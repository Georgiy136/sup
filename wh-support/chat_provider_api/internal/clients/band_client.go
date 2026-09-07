package clients

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type BandConfiguration struct {
	URL                string `json:"band_url"`
	WebsocketURL       string `json:"websocket_url"`
	BotToken           string `json:"bot_token"`
	ChannelID          string `json:"band_channel_id"`
	RetryCount         int64  `json:"retry_count"`
	InsecureSkipVerify *bool  `json:"insecure_skip_verify"`
}

type BandClient struct {
	client *model.Client4
	botID  string
	cfg    BandConfiguration
}

func NewBandClient() *BandClient {
	return &BandClient{}
}

func (b *BandClient) Configure(ctx context.Context, config configs.Config) {
	const key = "band_bot_client"

	if err := jsoniter.Unmarshal(config.GetByServiceKeyRequired(key), &b.cfg); err != nil {
		logrus.Panicf("cannot unmarshal - %s: %s", key, err)
	}
	if b.cfg.URL == "" || b.cfg.BotToken == "" || b.cfg.ChannelID == "" {
		logrus.Panicf("required config fields are empty in %s", key)
	}

	b.client = model.NewAPIv4Client(b.cfg.URL)
	b.client.SetOAuthToken(b.cfg.BotToken)
	if b.cfg.InsecureSkipVerify != nil && *b.cfg.InsecureSkipVerify {
		logrus.Warnf("[%s] tls_insecure_skip_verify is enabled, TLS certificate verification is disabled", key)
		b.client.HTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
		}
	}
	b.initBotID(ctx)
}

func (b *BandClient) initBotID(ctx context.Context) {
	me, _, err := b.client.GetMe(ctx, "")
	if err != nil {
		logrus.Panicf("cannot get bot user: %s", err)
	}
	b.botID = me.Id
	b.client.HTTPHeader["User-Agent"] = me.Username
}

func (b *BandClient) ChannelID() string {
	return b.cfg.ChannelID
}

func (b *BandClient) GetBotID() string { return b.botID }

func (b *BandClient) CreatePost(ctx context.Context, post *model.Post) (*model.Post, error) {
	p, resp, err := b.client.CreatePost(ctx, post)
	if err != nil && resp != nil {
		logrus.WithFields(logrus.Fields{
			"post": fmt.Sprintf("%+v", post),
			"resp": fmt.Sprintf("%+v", resp),
			"err":  fmt.Sprintf("%+v", err),
		}).Debug("create post error")
		return nil, models.NewHTTPError(resp.StatusCode, fmt.Errorf("create post error: %w", err))
	}
	if err != nil {
		return nil, fmt.Errorf("create post error: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create post unexpected status: %d", resp.StatusCode)
	}
	return p, nil
}

func (b *BandClient) GetPostThreadWithOpts(ctx context.Context, postID string, opts model.GetPostsOptions) (*model.PostList, error) {
	posts, resp, err := b.client.GetPostThreadWithOpts(ctx, postID, "", opts)
	if err != nil && resp != nil {
		return nil, models.NewHTTPError(resp.StatusCode, fmt.Errorf("get post thread error: %w", err))
	}
	if err != nil {
		return nil, fmt.Errorf("get post thread error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get post thread unexpected status: %d", resp.StatusCode)
	}
	return posts, nil
}

func (b *BandClient) GetFileInfosForPost(ctx context.Context, postID string) ([]*model.FileInfo, error) {
	files, resp, err := b.client.GetFileInfosForPost(ctx, postID, "")
	if err != nil && resp != nil {
		return nil, models.NewHTTPError(resp.StatusCode, fmt.Errorf("get file info for post error: %w", err))
	}
	if err != nil {
		return nil, fmt.Errorf("get file info for post error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get file info for post unexpected status: %d", resp.StatusCode)
	}
	return files, nil
}

func (b *BandClient) UploadFileStream(ctx context.Context, body io.Reader, filename string, fileSize int64) (*model.FileInfo, error) {
	us, resp, err := b.client.CreateUpload(ctx, &model.UploadSession{
		ChannelId: b.cfg.ChannelID,
		Filename:  filename,
		FileSize:  fileSize,
		Type:      model.UploadTypeAttachment,
	})
	if err != nil && resp != nil {
		return nil, models.NewHTTPError(resp.StatusCode, fmt.Errorf("create upload session: %w", err))
	}
	if err != nil {
		return nil, fmt.Errorf("create upload session: %w", err)
	}

	fileInfo, resp, err := b.client.UploadData(ctx, us.Id, body)
	if err != nil && resp != nil {
		return nil, models.NewHTTPError(resp.StatusCode, fmt.Errorf("upload file data: %w", err))
	}
	if err != nil {
		return nil, fmt.Errorf("upload file data: %w", err)
	}

	return fileInfo, nil
}

func (b *BandClient) GetFileLink(ctx context.Context, fileID string) (string, error) {
	fileLink, resp, err := b.client.GetFileLink(ctx, fileID)
	if err != nil && resp != nil {
		return "", models.NewHTTPError(resp.StatusCode, fmt.Errorf("get file link error: %w", err))
	}
	if err != nil {
		return "", fmt.Errorf("get file link error: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get file link unexpected status: %d", resp.StatusCode)
	}
	return fileLink, nil
}

func (b *BandClient) NewWebSocketClient() (*model.WebSocketClient, error) {
	dialer := websocket.DefaultDialer
	if b.cfg.InsecureSkipVerify != nil && *b.cfg.InsecureSkipVerify {
		dialer = &websocket.Dialer{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		}
	}

	wsClient, err := model.NewWebSocketClient4WithDialer(dialer, b.cfg.WebsocketURL, b.cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("create mattermost websocket client: %w", err)
	}
	return wsClient, nil
}
