package models

import (
	"time"

	jsoniter "github.com/json-iterator/go"
)

type PresenceRequest struct {
	Channel string `json:"channel"`
}

type PresenceResult struct {
	Presence map[string]PresenceClient `json:"presence"`
}

type PresenceClient struct {
	User   string `json:"user"`
	Client string `json:"client"`
}

type HistoryRequest struct {
	Channel string          `json:"channel"`
	Limit   int64           `json:"limit"`
	Reverse bool            `json:"reverse"`
	Since   *StreamPosition `json:"since,omitempty"`
}

type StreamPosition struct {
	Offset int64  `json:"offset"`
	Epoch  string `json:"epoch"`
}

type StreamResponse[T any] struct {
	Result *T           `json:"result"`
	Error  *StreamError `json:"error"`
}

type StreamError struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

type HistoryOffset struct {
	Epoch     string    `json:"epoch"`
	Offset    int64     `json:"offset"`
	UpdatedAt time.Time `json:"updated_at"`
}

type HistoryResult struct {
	Publications []Publication `json:"publications"`
	Epoch        string        `json:"epoch"`
	Offset       int64         `json:"offset"`
}

type Publication struct {
	Data   jsoniter.RawMessage `json:"data"`
	Offset int64               `json:"offset"`
}
