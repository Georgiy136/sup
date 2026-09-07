package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"

	mattermost "github.com/mattermost/mattermost/server/public/model"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/constant"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/utils"
)

type ChatProviderService struct {
	bandClient        BandClient
	employeeInfoStore EmployeeInfoStore
}

func New(
	bandClient BandClient,
	employeeInfoStore EmployeeInfoStore,

) *ChatProviderService {
	return &ChatProviderService{
		bandClient:        bandClient,
		employeeInfoStore: employeeInfoStore,
	}
}

func (c *ChatProviderService) CreateChat(ctx context.Context, employeeID, ticketID int64) (*models.CreateChatResponse, error) {
	const op = "CreateChat"
	employeeName, err := c.employeeInfoStore.GetEmployeeName(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("[%s] can't get employee name: %w", op, err)
	}
	if employeeName == "" {
		employeeName = strconv.Itoa(int(employeeID))
	}

	chatName := fmt.Sprintf(constant.ChatNameTemplate, ticketID)

	props := mattermost.StringInterface{
		"employee_id":   strconv.FormatInt(employeeID, 10),
		"ticket_id":     strconv.FormatInt(ticketID, 10),
		"employee_name": employeeName,
	}

	post, err := c.bandClient.CreatePost(ctx, &mattermost.Post{
		ChannelId: c.bandClient.ChannelID(),
		Message:   chatName,
		Props:     props,
	})
	if err != nil {
		return nil, fmt.Errorf("[%s] can't create band chat: %w", op, err)
	}

	return &models.CreateChatResponse{
		TicketID: ticketID,
		Chat: models.ChatMeta{
			ID:         post.Id,
			ChatName:   chatName,
			EmployeeID: employeeID,
			CreatedDt:  utils.ToRFC3339Nano(post.CreateAt),
		},
	}, nil
}

func (c *ChatProviderService) CreateChatMessage(ctx context.Context, employeeID int64, chatID, message string, ticketID int64, files []string) (*models.CreatePostResponse, error) {
	const op = "CreateChatMessage"
	employeeName, err := c.employeeInfoStore.GetEmployeeName(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("[%s] can't get employee name by employee_id=%d: %w", op, employeeID, err)
	}
	if employeeName == "" {
		employeeName = strconv.Itoa(int(employeeID))
	}

	props := mattermost.StringInterface{
		"employee_id":   strconv.FormatInt(employeeID, 10),
		"ticket_id":     strconv.FormatInt(ticketID, 10),
		"employee_name": employeeName,
	}

	post, err := c.bandClient.CreatePost(ctx, &mattermost.Post{
		ChannelId: c.bandClient.ChannelID(),
		Message:   message,
		RootId:    chatID,
		Props:     props,
		FileIds:   files,
	})
	if err != nil {
		return nil, fmt.Errorf("[%s] can't create chat post: %w", op, err)
	}

	msg, err := c.postToChatMessage(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("[%s] can't convert created post: %w", op, err)
	}

	return &models.CreatePostResponse{
		TicketID: ticketID,
		Message:  *msg,
	}, nil
}

func (c *ChatProviderService) UploadFileStream(ctx context.Context, filename string, body io.Reader, fileSize int64) (*models.FileInfo, error) {
	const op = "UploadFileStream"
	fileInfo, err := c.bandClient.UploadFileStream(ctx, body, filename, fileSize)
	if err != nil {
		return nil, fmt.Errorf("[%s] can't upload file: %w", op, err)
	}

	return &models.FileInfo{
		FileID:   fileInfo.Id,
		FileName: fileInfo.Name,
		MimeType: fileInfo.MimeType,
	}, nil
}

func (c *ChatProviderService) GetDownloadLink(ctx context.Context, fileID string) (*models.GetDownloadLinkResponse, error) {
	const op = "GetDownloadLink"
	downloadURL, err := c.bandClient.GetFileLink(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("[%s] can't get file link: %w", op, err)
	}

	return &models.GetDownloadLinkResponse{DownloadURL: downloadURL}, nil
}

func (c *ChatProviderService) GetChat(ctx context.Context, chatID string) (*models.GetChatResponse, error) {
	const op = "GetChat"
	opts := mattermost.GetPostsOptions{
		PerPage:   constant.DefaultMessagesLimit,
		Direction: "up",
	}
	postList, err := c.bandClient.GetPostThreadWithOpts(ctx, chatID, opts)
	if err != nil {
		return nil, fmt.Errorf("[%s] can't get chat thread: %w", op, err)
	}

	page, err := c.buildThreadPage(ctx, postList, true)
	if err != nil {
		return nil, fmt.Errorf("[%s] can't convert chat thread: %w", op, err)
	}

	rootMsg, rootMessage, err := c.getRootMessage(ctx, postList, chatID)
	if err != nil {
		return nil, fmt.Errorf("[%s] can't get root message: %w", op, err)
	}

	return &models.GetChatResponse{
		Chat: models.ChatThread{
			Created:       false,
			ChatName:      rootMsg.Message,
			EmployeeID:    rootMessage.EmployeeID,
			CreatedDt:     rootMessage.CreatedDt,
			RootMessage:   *rootMessage,
			MessagesOrder: page.MessagesOrder,
			Messages:      page.Messages,
		},
		HasMore:        page.HasMore,
		Limit:          len(page.Messages),
		Before:         page.Before,
		BeforeCreateAt: page.BeforeCreateAt,
	}, nil
}

func (c *ChatProviderService) GetChatHistory(ctx context.Context, chatID string, countPost int64, fromPost string, fromCreateAt int64) (*models.HistoryChatResponse, error) {
	const op = "GetChatHistory"
	opts := mattermost.GetPostsOptions{
		PerPage:      int(countPost) + 1,
		FromPost:     fromPost,
		FromCreateAt: fromCreateAt,
		Direction:    "up",
	}
	postList, err := c.bandClient.GetPostThreadWithOpts(ctx, chatID, opts)
	if err != nil {
		return nil, fmt.Errorf("[%s] can't get chat history: %w", op, err)
	}

	_, rootMessage, err := c.getRootMessage(ctx, postList, chatID)
	if err != nil {
		return nil, fmt.Errorf("[%s] can't get root message: %w", op, err)
	}

	page, err := c.buildThreadPage(ctx, postList, true)
	if err != nil {
		return nil, fmt.Errorf("[%s] can't convert chat history: %w", op, err)
	}

	return &models.HistoryChatResponse{
		MessagesOrder:  page.MessagesOrder,
		RootMessage:    *rootMessage,
		Messages:       page.Messages,
		HasMore:        page.HasMore,
		Limit:          len(page.Messages),
		Before:         page.Before,
		BeforeCreateAt: page.BeforeCreateAt,
	}, nil
}

func (c *ChatProviderService) getRootMessage(ctx context.Context, postList *mattermost.PostList, chatID string) (*mattermost.Post, *models.ChatMessage, error) {
	if postList == nil {
		return nil, nil, errors.New("post list is nil")
	}
	rootMsg, exists := postList.Posts[chatID]
	if !exists || rootMsg == nil {
		return nil, nil, errors.New("root message not found")
	}

	rootMessage, err := c.postToChatMessage(ctx, rootMsg)
	if err != nil {
		return nil, nil, fmt.Errorf("can't convert root message: %w", err)
	}

	return rootMsg, rootMessage, nil
}
