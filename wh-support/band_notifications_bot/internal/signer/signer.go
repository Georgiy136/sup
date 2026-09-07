package signer

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type Signer struct {
	cfg Config
}

func New() *Signer {
	return &Signer{}
}

func (s *Signer) Configure(_ context.Context, conf configs.Config) {
	const key = "signer_config"
	if err := jsoniter.Unmarshal(conf.GetByServiceKeyRequired(key), &s.cfg); err != nil {
		logrus.Panicf("error parsing '%s' configs - %v", key, err)
	}
}

func (s *Signer) Sign(parts ...string) string {
	return sign(s.cfg.SigningKey, parts...)
}

func sign(secret string, data ...string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := io.WriteString(mac, strings.Join(data, "\n")); err != nil {
		logrus.Errorf("gen hmac sign error: %v", err)
	}
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *Signer) Compare(signature string, parts ...string) bool {
	return s.Sign(parts...) == signature
}
