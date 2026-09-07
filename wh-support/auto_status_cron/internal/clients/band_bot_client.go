package clients

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"golang.org/x/time/rate"
)

type BandBotConfiguration struct {
	Url                   string `json:"band_url"`
	BotToken              string `json:"bot_token"`
	ChannelID             string `json:"band_channel_id"`
	RetryCount            int    `json:"retry_count"`
	ResponseTimeout       string `json:"response_timeout"`
	ResponseTimeoutParsed time.Duration
}

type BandBotClient struct {
	config      BandBotConfiguration
	client      *model.Client4
	rateLimiter *rate.Limiter
}

func NewBandBotClient() *BandBotClient {
	return &BandBotClient{}
}

func (b *BandBotClient) Configure(ctx context.Context, config configs.Config) {
	const (
		defaultRetryCount      = 1
		numReq                 = 3
		durReq                 = 1 * time.Second
		defaultResponseTimeout = 1 * time.Second
		key                    = "band_bot_client"
	)

	rawConfig := config.GetByServiceKeyRequired(key)

	var cfg BandBotConfiguration
	if err := jsoniter.Unmarshal(rawConfig, &cfg); err != nil {
		logrus.Panicf("cannot unmarshal - %s: %s", key, err)
	}

	var validate = validator.New()
	if err := validate.Struct(cfg); err != nil {
		logrus.Panicf("error validate config - %s: %s", key, err)
	}

	logrus.Debugf("client band_bot_client inited url- %s", cfg.Url)

	b.config = cfg
	b.config.RetryCount += defaultRetryCount
	timeout, err := time.ParseDuration(b.config.ResponseTimeout)
	if err != nil {
		b.config.ResponseTimeoutParsed = defaultResponseTimeout
		return
	}
	b.config.ResponseTimeoutParsed = timeout

	b.rateLimiter = rate.NewLimiter(rate.Every(durReq/numReq), numReq)
	b.client = model.NewAPIv4Client(cfg.Url)

	b.client.SetOAuthToken(b.config.BotToken)
	if b.config.RetryCount < 0 {
		b.config.RetryCount = 0
	}
}

func (b *BandBotClient) SendMessage(message string) (*string, error) {
	if len(message) == 0 {
		return nil, nil
	}
	var err error
	rootPost := &model.Post{
		ChannelId: b.config.ChannelID,
		Message:   message,
	}

	for range b.config.RetryCount {
		post, _, er := b.sendMessage(rootPost)
		if er != nil {
			err = er
			continue
		}
		return &post.Id, nil
	}
	return nil, err
}

func (b *BandBotClient) sendMessage(post *model.Post) (*model.Post, *model.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), b.config.ResponseTimeoutParsed)
	defer cancel()
	return b.client.CreatePost(ctx, post)
}
