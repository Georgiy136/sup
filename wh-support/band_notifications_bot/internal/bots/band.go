package bots

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	customerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"go.uber.org/ratelimit"

	jsoniter "github.com/json-iterator/go"
	mattermost "github.com/mattermost/mattermost/server/public/model"
	"github.com/sirupsen/logrus"
)

type cache interface {
	GetBandUserID(employeeID int64) (string, bool)
	SetBandUserID(employeeID int64, userID string)
}

type bandBotConfiguration struct {
	Url              string            `json:"band_url"`
	BotToken         string            `json:"bot_token"`
	RetryCount       int               `json:"retry_count"`
	ResponseTimeout  string            `json:"response_timeout"`
	RequiredUsers    map[string]string `json:"required_users"`
	ActionWebhookURL string            `json:"action_webhook_url"`
}

type BandBot struct {
	retryCount       int
	requiredUsers    map[string]string
	actionWebhookURL string
	client           *mattermost.Client4
	rateLimiter      ratelimit.Limiter
	cache            cache
}

func NewBandBot(cache cache) *BandBot {
	return &BandBot{
		cache: cache,
	}
}

func (b *BandBot) Configure(_ context.Context, config configs.Config) {
	const (
		defaultRetryCount      = 1
		defaultResponseTimeout = 15 * time.Second
		msgLimit               = 20
		msgTimeLimit           = 1 * time.Second
	)

	confRaw := config.GetByServiceKeyRequired("band_bot_config")

	botConfig := new(bandBotConfiguration)

	err := jsoniter.Unmarshal(confRaw, botConfig)
	if err != nil {
		logrus.Panicf("cannot unmarshal bot config: %v", err)
	}

	responseTimeout := defaultResponseTimeout
	if botConfig.ResponseTimeout != "" {
		responseTimeout, err = time.ParseDuration(botConfig.ResponseTimeout)
		if err != nil {
			logrus.Panicf("cannot parse response_timeout: %v", err)
		}
	}

	b.client = mattermost.NewAPIv4Client(botConfig.Url)
	b.rateLimiter = ratelimit.New(msgLimit, ratelimit.Per(msgTimeLimit))
	b.client.HTTPClient = &http.Client{
		Timeout: responseTimeout,
	}

	if botConfig.BotToken == "" {
		logrus.Panicf("band bot token is required")
	}
	b.client.SetOAuthToken(botConfig.BotToken)

	b.retryCount = defaultRetryCount

	if botConfig.RequiredUsers == nil {
		botConfig.RequiredUsers = map[string]string{}
	}
	b.requiredUsers = botConfig.RequiredUsers

	if botConfig.RetryCount > 0 {
		b.retryCount = botConfig.RetryCount
	}
	if botConfig.ActionWebhookURL == "" {
		logrus.Panicf("band bot action webhook URL is required")
	}
	b.actionWebhookURL = botConfig.ActionWebhookURL

	if err = b.initBot(); err != nil {
		logrus.Panicf("band bot init failed: %v", err)
	}
}

func (b *BandBot) initBot() error {
	me, _, err := b.client.GetMe(context.Background(), "")
	if err != nil {
		return fmt.Errorf("cannot get bot user: %s", err)
	}
	b.client.HTTPHeader["User-Agent"] = me.Username
	return nil
}

func (b *BandBot) SendNotification(ctx context.Context, recipient models.BandRecipient, msgNotification string, attachments ...models.BandAttachment) (*models.BandSentNotification, error) {
	b.rateLimiter.Take()

	bandAttachments, err := b.buildBandAttachments(attachments)
	if err != nil {
		return nil, fmt.Errorf("cannot build band attachments: %w", err)
	}

	return b.sendDirectNotification(ctx, msgNotification, recipient.EmployeeID, bandAttachments)
}

func (b *BandBot) SendChannelMessage(ctx context.Context, channelID, message string) error {
	if channelID == "" {
		return fmt.Errorf("empty channel id: %w", customerrors.ErrNilValue)
	}
	if message == "" {
		return fmt.Errorf("empty message: %w", customerrors.ErrNilValue)
	}

	b.rateLimiter.Take()
	_, resp, err := b.client.CreatePost(ctx, &mattermost.Post{ChannelId: channelID, Message: message})
	if err != nil {
		return fmt.Errorf("can't send channel message: %w", err)
	}
	if resp == nil {
		return fmt.Errorf("can't send channel message: %w", customerrors.ErrNilResponse)
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("can't send channel message: err status code %d", resp.StatusCode)
	}
	return nil
}

func (b *BandBot) sendDirectNotification(ctx context.Context, msgNotification string, directEmployeeID int64, attachments []*mattermost.MessageAttachment) (*models.BandSentNotification, error) {
	channel, err := b.getOrCreateDirectChannelWithBot(ctx, directEmployeeID)
	if err != nil {
		return nil, fmt.Errorf("cannot get or create direct channel with bot: %w", err)
	}

	if channel.LastPostAt == 0 {
		if err := b.sendHelloMessage(ctx, channel.Id); err != nil {
			return nil, fmt.Errorf("cannot send hello message: %w", err)
		}
	}

	post, err := b.sendMessage(ctx, channel.Id, msgNotification, attachments)
	if err != nil {
		return nil, fmt.Errorf("cannot send message: %w", err)
	}
	if post == nil {
		return nil, nil
	}

	return &models.BandSentNotification{
		PostID:    post.Id,
		ChannelID: channel.Id,
	}, nil
}

func (b *BandBot) sendMessage(ctx context.Context, channelId, msg string, attachments []*mattermost.MessageAttachment) (*mattermost.Post, error) {
	post := &mattermost.Post{
		ChannelId: channelId,
		Message:   msg,
	}
	if len(attachments) > 0 {
		mattermost.ParseMessageAttachment(post, attachments)
	}

	createdPost, resp, err := b.client.CreatePost(ctx, post)
	if err != nil {
		if strings.Contains(err.Error(), "У вас нет соответствующих прав") {
			logrus.Infof("can't create post: %v", err)
			return nil, nil
		}
		return nil, fmt.Errorf("can't create post: %w", err)
	}
	if resp == nil {
		return nil, fmt.Errorf("can't create post: %w", customerrors.ErrNilResponse)
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("can't create post: err status code %d", resp.StatusCode)
	}

	return createdPost, nil
}

func (b *BandBot) UpdatePost(ctx context.Context, postID, message string, attachments ...models.BandAttachment) error {
	if postID == "" {
		return fmt.Errorf("empty post id: %w", customerrors.ErrNilValue)
	}

	post := &mattermost.Post{Id: postID, Message: message}
	postPropsAttachments, err := b.buildBandAttachments(attachments)
	if err != nil {
		return fmt.Errorf("cannot build attachments: %w", err)
	}
	post.AddProp(mattermost.PostPropsAttachments, postPropsAttachments)
	_, resp, err := b.client.UpdatePost(ctx, postID, post)
	if err != nil {
		return fmt.Errorf("can't update post %s: %w", postID, err)
	}
	if resp == nil {
		return fmt.Errorf("can't update post %s: %w", postID, customerrors.ErrNilResponse)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("can't update post %s: err status code %d", postID, resp.StatusCode)
	}
	return nil
}

func (b *BandBot) GetPostMessage(ctx context.Context, postID string) (string, error) {
	if postID == "" {
		return "", fmt.Errorf("empty post id: %w", customerrors.ErrNilValue)
	}

	post, resp, err := b.client.GetPost(ctx, postID, "")
	if err != nil {
		return "", fmt.Errorf("can't get post %s: %w", postID, err)
	}
	if resp == nil {
		return "", fmt.Errorf("can't get post %s: %w", postID, customerrors.ErrNilResponse)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("can't get post %s: err status code %d", postID, resp.StatusCode)
	}
	return post.Message, nil
}

func (b *BandBot) OpenInteractiveDialog(ctx context.Context, triggerID string, conf models.BandInteractiveDialogConfig, state models.BandDialogState) error {
	if triggerID == "" {
		return fmt.Errorf("empty trigger id: %w", customerrors.ErrNilValue)
	}

	callbackURL, err := buildCallbackURL(b.actionWebhookURL, conf.Path)
	if err != nil {
		return fmt.Errorf("can't build dialog URL: %w", err)
	}

	stateRaw, err := jsoniter.MarshalToString(state)
	if err != nil {
		return fmt.Errorf("can't marshal dialog state: %w", err)
	}

	var elements []mattermost.DialogElement
	if len(conf.Elements) > 0 {
		elements = b.convertDialogElements(conf.Elements)
	}

	resp, err := b.client.OpenInteractiveDialog(ctx, mattermost.OpenDialogRequest{
		TriggerId: triggerID,
		URL:       callbackURL,
		Dialog: mattermost.Dialog{
			CallbackId:       conf.CallbackID,
			Title:            conf.Title,
			IntroductionText: conf.IntroductionText,
			State:            stateRaw,
			SubmitLabel:      conf.SubmitLabel,
			Elements:         elements,
		},
	})
	if err != nil {
		return fmt.Errorf("can't open dialog for ticket %d: %w", state.TicketID, err)
	}
	if resp == nil {
		return fmt.Errorf("can't open dialog for ticket %d: %w", state.TicketID, customerrors.ErrNilResponse)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("can't open dialog for ticket %d: err status code %d", state.TicketID, resp.StatusCode)
	}
	return nil
}

func (b *BandBot) convertDialogElements(elements []models.BandDialogElement) []mattermost.DialogElement {
	result := make([]mattermost.DialogElement, 0, len(elements))
	for _, el := range elements {
		mmEl := mattermost.DialogElement{
			DisplayName: el.DisplayName,
			Name:        el.Name,
			Type:        string(el.Type),
			Placeholder: el.Placeholder,
			Default:     el.DefaultValue,
			Optional:    !el.Required,
		}

		for _, opt := range el.Options {
			mmEl.Options = append(mmEl.Options, &mattermost.PostActionOptions{
				Text:  opt.Text,
				Value: opt.Value,
			})
		}
		result = append(result, mmEl)
	}
	return result
}

func (b *BandBot) GetEmployeeIDByUserID(ctx context.Context, userID string) (int64, error) {
	if userID == "" {
		return 0, fmt.Errorf("empty user id: %w", customerrors.ErrNilValue)
	}

	user, resp, err := b.client.GetUser(ctx, userID, "")
	if err != nil {
		return 0, fmt.Errorf("can't get band user %s: %w", userID, err)
	}
	if resp == nil {
		return 0, fmt.Errorf("can't get band user %s: %w", userID, customerrors.ErrNilResponse)
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("can't get band user %s: err status code %d", userID, resp.StatusCode)
	}
	if user.AuthData == nil || *user.AuthData == "" {
		return 0, fmt.Errorf("band user %s has empty auth data: %w", userID, customerrors.ErrUserNotFound)
	}

	employeeID, err := strconv.ParseInt(*user.AuthData, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("can't parse employee id from auth data %v: %w", *user.AuthData, err)
	}
	return employeeID, nil
}

func (b *BandBot) GetUsernameByEmployeeID(ctx context.Context, employeeID int64) (string, error) {
	userID, err := b.searchUserIDByEmployeeId(ctx, employeeID)
	if err != nil {
		return "", err
	}

	user, resp, err := b.client.GetUser(ctx, userID, "")
	if err != nil {
		return "", fmt.Errorf("can't get band user %s: %w", userID, err)
	}
	if resp == nil {
		return "", fmt.Errorf("can't get band user %s: %w", userID, customerrors.ErrNilResponse)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("can't get band user %s: err status code %d", userID, resp.StatusCode)
	}
	if user == nil || user.Username == "" {
		return "", fmt.Errorf("band user %s has empty username: %w", userID, customerrors.ErrUserNotFound)
	}
	return user.Username, nil
}

func (b *BandBot) buildBandAttachments(attachments []models.BandAttachment) ([]*mattermost.MessageAttachment, error) {
	bandAttachments := make([]*mattermost.MessageAttachment, 0, len(attachments))
	for _, a := range attachments {
		actions, err := b.buildBandActions(a.Actions)
		if err != nil {
			return nil, fmt.Errorf("can't build band actions: %w", err)
		}

		bandAttachments = append(bandAttachments, &mattermost.MessageAttachment{
			Color:      a.Color,
			AuthorName: a.AuthorName,
			Text:       a.Text,
			Actions:    actions,
		})
	}
	return bandAttachments, nil
}

func (b *BandBot) buildBandActions(actions []models.BandAction) ([]*mattermost.PostAction, error) {
	if len(actions) == 0 {
		return nil, nil
	}
	bandActions := make([]*mattermost.PostAction, 0, len(actions))
	for _, action := range actions {
		callbackURL, err := buildCallbackURL(b.actionWebhookURL, action.Path)
		if err != nil {
			return nil, fmt.Errorf("can't build action URL for path %s: %w", action.Path, err)
		}
		bandActions = append(bandActions, &mattermost.PostAction{
			Id:    action.ID,
			Type:  mattermost.PostActionTypeButton,
			Name:  action.Name,
			Style: action.Style,
			Integration: &mattermost.PostActionIntegration{
				URL:     callbackURL,
				Context: action.Context,
			},
		})
	}
	return bandActions, nil
}

func buildCallbackURL(baseUrl, path string) (string, error) {
	base, err := url.Parse(baseUrl)
	if err != nil {
		return "", fmt.Errorf("invalid base url: %w", err)
	}
	return base.JoinPath(path).String(), nil
}

func (b *BandBot) GetDirectChannelID(ctx context.Context, employeeID int64) (string, error) {
	channel, err := b.getOrCreateDirectChannelWithBot(ctx, employeeID)
	if err != nil {
		return "", fmt.Errorf("can't get or create direct channel: %w", err)
	}
	return channel.Id, nil
}

func (b *BandBot) getOrCreateDirectChannelWithBot(ctx context.Context, directEmployeeID int64) (channel *mattermost.Channel, err error) {
	botUserId, err := b.getMeUserId(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't get me user id: %w", err)
	}

	userId, err := b.searchUserIDByEmployeeId(ctx, directEmployeeID)
	if err != nil {
		return nil, fmt.Errorf("can't get user id by employee: %w", err)
	}

	channel, err = b.getOrCreateDirectChannel(ctx, botUserId, userId)
	if err != nil {
		return nil, fmt.Errorf("can't get or create direct channel: %w", err)
	}

	return channel, nil
}

func (b *BandBot) getOrCreateDirectChannel(ctx context.Context, userId1, userId2 string) (*mattermost.Channel, error) {
	channel, resp, err := b.client.CreateDirectChannel(ctx, userId1, userId2)
	if err != nil {
		return nil, fmt.Errorf("cannot create direct channel: %w", err)
	}

	if resp == nil {
		return nil, fmt.Errorf("cannot create direct channel: %w", customerrors.ErrNilResponse)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("cannot create direct channel; err status code: %d", resp.StatusCode)
	}

	return channel, nil
}

func (b *BandBot) getMeUserId(ctx context.Context) (string, error) {
	user, resp, err := b.client.GetMe(ctx, "")
	if err != nil {
		return "", fmt.Errorf("cannot get user info about me: %w", err)
	}

	if resp == nil {
		return "", fmt.Errorf("cannot get user info about me: %w", customerrors.ErrNilResponse)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cannot get user info about me, err status code: %d", resp.StatusCode)
	}

	return user.Id, nil
}

func (b *BandBot) searchUserIDByEmployeeId(ctx context.Context, employeeId int64) (string, error) {
	termEmployeeId := strconv.FormatInt(employeeId, 10)

	if userId, ok := b.requiredUsers[termEmployeeId]; ok {
		return userId, nil
	}

	if userId, ok := b.cache.GetBandUserID(employeeId); ok {
		return userId, nil
	}

	searchReq := &mattermost.UserSearch{
		Term: termEmployeeId,
	}

	users, resp, err := doWithRetry(ctx, b.retryCount, func() ([]*mattermost.User, *mattermost.Response, error) {
		return b.client.SearchUsers(ctx, searchReq)
	})
	if err != nil {
		if strings.Contains(err.Error(), "user not found") || strings.Contains(err.Error(), "Невозможно найти пользователя, соответствующего критериям поиска") {
			return "", fmt.Errorf("cannot search users, employee_id: %s, err: %w", termEmployeeId, customerrors.ErrUserNotFound)
		}
		return "", fmt.Errorf("cannot search users: %w", err)
	}

	if resp == nil {
		return "", fmt.Errorf("cannot search users: %w", customerrors.ErrNilValue)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cannot search users, err status code: %d", resp.StatusCode)
	}

	if len(users) == 0 {
		return "", fmt.Errorf("cannot find the user, employee_id: %s, err: %w", termEmployeeId, customerrors.ErrUserNotFound)
	}

	if len(users) > 1 {
		var countUsers int
		users, countUsers = b.deepSearchUserIDByEmployeeID(users, termEmployeeId)
		if countUsers == 0 {
			return "", fmt.Errorf("cannot deep search users: %w", customerrors.ErrUserNotFound)
		}

		if countUsers > 1 {
			return "", fmt.Errorf("can't search users: too many users by employeeID: %d, err: %w", len(users), customerrors.ErrTooManyUsers)
		}
	}

	b.cache.SetBandUserID(employeeId, users[0].Id)
	return users[0].Id, nil
}

func (b *BandBot) deepSearchUserIDByEmployeeID(users []*mattermost.User, termEmployeeId string) ([]*mattermost.User, int) {
	var result []*mattermost.User

	for i := range users {
		if users[i].AuthData != nil && *users[i].AuthData == termEmployeeId {
			result = append(result, users[i])
		}
	}

	return result, len(result)
}

func (b *BandBot) sendHelloMessage(ctx context.Context, channelId string) error {
	const helloMessage = `Вас приветствует Wh Support Notification Bot 👋
Я умею рассылать уведомления о новых заявках, созданых через сервис [Wh Support](https://support.wbwh.tech)`

	_, err := b.sendMessage(ctx, channelId, helloMessage, nil)
	return err
}
