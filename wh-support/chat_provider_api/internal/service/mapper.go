package service

import (
	"context"
	"errors"
	"fmt"

	mattermost "github.com/mattermost/mattermost/server/public/model"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/utils"
)

func (c *ChatProviderService) buildThreadPage(
	ctx context.Context,
	postList *mattermost.PostList,
	skipRoot bool,
) (models.ThreadMessagesPage, error) {
	if postList == nil {
		return models.ThreadMessagesPage{}, errors.New("post list is nil")
	}
	if len(postList.Order) == 0 {
		return models.ThreadMessagesPage{}, nil
	}

	botID := c.bandClient.GetBotID()
	messagesOrder := make([]string, 0, len(postList.Order))
	messages := make(map[string]models.ChatMessage, len(postList.Order))

	for _, postID := range postList.Order {
		post, ok := postList.Posts[postID]
		if !ok || post == nil {
			continue
		}
		if post.UserId != botID {
			logrus.Warnf("[buildThreadPage] post is not from bot; post=%v", post)
			continue
		}
		if skipRoot && post.RootId == "" {
			continue
		}

		msg, err := c.postToChatMessage(ctx, post)
		if err != nil {
			return models.ThreadMessagesPage{}, err
		}

		messagesOrder = append(messagesOrder, postID)
		messages[postID] = *msg
	}

	hasMore := postList.HasNext != nil && *postList.HasNext
	var (
		before         string
		beforeCreateAt string
	)
	if hasMore && len(messagesOrder) > 0 {
		lastID := messagesOrder[len(messagesOrder)-1]
		before = lastID
		if lastPost, ok := postList.Posts[lastID]; ok && lastPost != nil {
			beforeCreateAt = utils.ToRFC3339Nano(lastPost.CreateAt)
		}
	}

	return models.ThreadMessagesPage{
		MessagesOrder:  messagesOrder,
		Messages:       messages,
		HasMore:        hasMore,
		Before:         before,
		BeforeCreateAt: beforeCreateAt,
	}, nil
}

func (c *ChatProviderService) postToChatMessage(ctx context.Context, post *mattermost.Post) (*models.ChatMessage, error) {
	const op = "PostToChatMessage"
	if post == nil {
		return nil, errors.New("post is nil")
	}

	outFiles := make([]models.FileInfo, 0, len(post.FileIds))
	if len(post.FileIds) != 0 {
		files, err := c.bandClient.GetFileInfosForPost(ctx, post.Id)
		if err != nil {
			logrus.Errorf("[%s] get post files: %v", op, err)
		}

		for _, f := range files {
			if f == nil {
				continue
			}
			outFiles = append(outFiles, models.FileInfo{
				FileID:    f.Id,
				FileName:  f.Name,
				MessageID: post.Id,
				MimeType:  f.MimeType,
			})
		}
	}

	props, err := utils.ExtractProps(post.GetProps())
	if err != nil {
		return nil, fmt.Errorf("[%s] can't extract post props: %w; props=%v", op, err, post.GetProps())
	}

	return &models.ChatMessage{
		ID:           post.Id,
		Message:      post.Message,
		EmployeeID:   props.EmployeeID,
		EmployeeName: props.EmployeeName,
		Files:        outFiles,
		CreatedDt:    utils.ToRFC3339Nano(post.CreateAt),
	}, nil
}
