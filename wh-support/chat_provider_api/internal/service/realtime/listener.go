package realtime

import (
	"context"
	"fmt"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/utils"

	jsoniter "github.com/json-iterator/go"
	mattermost "github.com/mattermost/mattermost/server/public/model"
	"github.com/sirupsen/logrus"
)

const reconnectDelay = 5 * time.Second

func (r *Realtime) startListener(ctx context.Context) {
	defer close(r.eventCh)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("[Realtime] listener stopped")
			return
		default:
			if err := r.listenBandEvents(ctx); err != nil {
				logrus.Errorf("realtime listener error: %v", err)

				select {
				case <-ctx.Done():
					return
				case <-time.After(reconnectDelay):
				}
			}
		}
	}
}

func (r *Realtime) listenBandEvents(ctx context.Context) error {
	const op = "listenBandEvents"

	wsClient, err := r.band.NewWebSocketClient()
	if err != nil {
		return fmt.Errorf("[%s] create websocket client: %w", op, err)
	}
	defer wsClient.Close()
	logrus.Infof("[%s] websocket client created", op)

	go wsClient.Listen()

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-wsClient.EventChannel:
			if !ok {
				if wsClient.ListenError != nil {
					return fmt.Errorf("[%s] websocket event channel closed: %w", op, wsClient.ListenError)
				}
				return fmt.Errorf("[%s] websocket event channel closed", op)
			}

			if event == nil || event.EventType() != mattermost.WebsocketEventPosted {
				continue
			}
			data := event.GetData()
			if data == nil {
				continue
			}

			rawPost, ok := data["post"].(string)
			if !ok || rawPost == "" {
				continue
			}

			post := mattermost.Post{}
			if err = jsoniter.UnmarshalFromString(rawPost, &post); err != nil {
				logrus.Errorf("[%s] can't parse post: %v", op, err)
				continue
			}

			if post.RootId == "" {
				continue
			}

			if post.UserId != r.band.GetBotID() {
				continue
			}

			props, err := utils.ExtractProps(post.GetProps())
			if err != nil {
				logrus.Errorf("[%s] can't extract post props: %v; props=%v", op, err, post.GetProps())
				continue
			}

			files := make([]models.FileInfo, 0, len(post.FileIds))
			if len(post.FileIds) > 0 {
				fileInfos, err := r.band.GetFileInfosForPost(ctx, post.Id)
				if err != nil {
					logrus.Errorf("[%s] can't get file infos for post_id=%s: %v", op, post.Id, err)
					continue
				}
				for _, fileInfo := range fileInfos {
					if fileInfo == nil {
						continue
					}
					files = append(files, models.FileInfo{
						FileID:    fileInfo.Id,
						FileName:  fileInfo.Name,
						MessageID: post.Id,
						MimeType:  fileInfo.MimeType,
					})
				}
			}

			postEvent := models.RealtimePostEvent{
				Message:      post.Message,
				MessageID:    post.Id,
				ChatID:       post.RootId,
				TicketID:     props.TicketID,
				EmployeeID:   props.EmployeeID,
				EmployeeName: props.EmployeeName,
				CreatedDt:    utils.ToRFC3339Nano(post.CreateAt),
				Files:        files,
			}

			logrus.Infof("[%s] got mattermost event post, ticket %d", op, props.TicketID)

			select {
			case r.eventCh <- postEvent:
			case <-ctx.Done():
				return nil
			}
		case <-wsClient.PingTimeoutChannel:
			return fmt.Errorf("[%s] websocket ping timeout", op)
		case resp, ok := <-wsClient.ResponseChannel:
			if !ok {
				if wsClient.ListenError != nil {
					return fmt.Errorf("[%s] websocket response channel closed: %w", op, wsClient.ListenError)
				}
				return fmt.Errorf("[%s] websocket response channel closed", op)
			}

			if resp == nil {
				continue
			}
			if status := resp.Status; status == "error" {
				logrus.Errorf("[%s] websocket response error: %v", op, resp.Data)
			}
		}
	}
}
