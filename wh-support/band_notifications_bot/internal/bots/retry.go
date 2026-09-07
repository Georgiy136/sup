package bots

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
)

const (
	retryDelay        = 200 * time.Millisecond
	defaultRetryCount = 2
)

func doWithRetry[T any](
	ctx context.Context,
	retryCount int,
	fn func() (T, *model.Response, error),
) (T, *model.Response, error) {
	var (
		result T
		resp   *model.Response
		err    error
	)
	if retryCount <= 0 {
		retryCount = defaultRetryCount
	}

	for attempt := 1; attempt <= retryCount; attempt++ {
		if ctx.Err() != nil {
			return result, resp, fmt.Errorf("request cancelled: %w", ctx.Err())
		}

		result, resp, err = fn()
		if err == nil {
			return result, resp, nil
		}

		if !isRetryableBandError(resp) || attempt == retryCount {
			return result, resp, err
		}

		select {
		case <-ctx.Done():
			return result, resp, fmt.Errorf("request cancelled: %w", ctx.Err())
		case <-time.After(retryDelay):
		}
	}

	return result, resp, err
}

func isRetryableBandError(resp *model.Response) bool {
	if resp == nil {
		return true
	}

	switch resp.StatusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusBadGateway,         // 502
		http.StatusServiceUnavailable, // 503
		http.StatusGatewayTimeout:     // 504
		return true
	default:
		return false
	}
}
