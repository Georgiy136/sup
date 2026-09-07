package redis

import (
	"context"
	"fmt"
	"time"

	ticketmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/models"

	jsoniter "github.com/json-iterator/go"
)

const (
	bandTicketPostsKeyFormat     = "ticket:%d:channel:%s:band_posts"
	bandTicketPostsCacheDuration = 14 * 24 * time.Hour
)

func (s *Storage) SaveBandTicketPost(ctx context.Context, post ticketmodels.TicketPost) error {
	raw, err := jsoniter.MarshalToString(post)
	if err != nil {
		return fmt.Errorf("marshal band ticket post: %w", err)
	}

	key := getBandTicketPostsKey(post.TicketID, post.ChannelID)
	pipe := s.Client.Pipeline()
	pipe.LPush(ctx, key, raw)
	pipe.Expire(ctx, key, bandTicketPostsCacheDuration)
	if _, err = pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis save band ticket post: %w", err)
	}
	return nil
}

func (s *Storage) GetBandTicketPostsByChannel(ctx context.Context, ticketID int64, channelID string) ([]ticketmodels.TicketPost, error) {
	rawPosts, err := s.Client.LRange(ctx, getBandTicketPostsKey(ticketID, channelID), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("redis get band ticket posts: %w", err)
	}
	if len(rawPosts) == 0 {
		return nil, nil
	}

	posts := make([]ticketmodels.TicketPost, 0, len(rawPosts))
	for _, raw := range rawPosts {
		var post ticketmodels.TicketPost
		if err = jsoniter.UnmarshalFromString(raw, &post); err != nil {
			return nil, fmt.Errorf("unmarshal band ticket post error: %w", err)
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func (s *Storage) DeleteBandTicketPosts(ctx context.Context, ticketID int64, channelID string) error {
	if err := s.Client.Del(ctx, getBandTicketPostsKey(ticketID, channelID)).Err(); err != nil {
		return fmt.Errorf("redis delete band ticket posts: %w", err)
	}
	return nil
}

func (s *Storage) DeleteBandTicketPostsByOperation(ctx context.Context, ticketID int64, channelID, operation string) error {
	posts, err := s.GetBandTicketPostsByChannel(ctx, ticketID, channelID)
	if err != nil {
		return err
	}

	stablePosts := make([]ticketmodels.TicketPost, 0, len(posts))
	for _, post := range posts {
		if post.Operation != operation {
			stablePosts = append(stablePosts, post)
		}
	}

	if len(stablePosts) == len(posts) {
		return nil
	}

	key := getBandTicketPostsKey(ticketID, channelID)
	if err = s.Client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis delete band ticket posts by operation: %w", err)
	}
	if len(stablePosts) == 0 {
		return nil
	}

	pipe := s.Client.Pipeline()
	for _, post := range stablePosts {
		raw, err := jsoniter.MarshalToString(post)
		if err != nil {
			return fmt.Errorf("marshal stable band ticket post: %w", err)
		}
		pipe.LPush(ctx, key, raw)
	}
	pipe.Expire(ctx, key, bandTicketPostsCacheDuration)
	if _, err = pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis save stable band ticket posts: %w", err)
	}
	return nil
}

func getBandTicketPostsKey(ticketID int64, channelID string) string {
	return fmt.Sprintf(bandTicketPostsKeyFormat, ticketID, channelID)
}
