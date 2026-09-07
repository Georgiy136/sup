package models

import (
	"time"

	jsoniter "github.com/json-iterator/go"
)

type CentrifugoPublishRequest struct {
	Channel string `json:"channel"`
	Data    any    `json:"data"`
}

type BodyByUploadIDRequest struct {
	IsSuccess bool   `json:"success"`
	Event     string `json:"event"`
	Meta      Meta   `json:"meta"`
}

type Meta struct {
	PublishedAt time.Time `json:"published_at"`
}

type CentrifugoResponse struct {
	Result jsoniter.RawMessage `json:"result"`
	Error  *CentrifugoError    `json:"error"`
}

type CentrifugoError struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}
