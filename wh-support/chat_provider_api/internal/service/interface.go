package service

import (
	"context"
	"io"

	mattermost "github.com/mattermost/mattermost/server/public/model"
)

type BandClient interface {
	ChannelID() string
	GetBotID() string
	UploadFileStream(ctx context.Context, body io.Reader, filename string, fileSize int64) (*mattermost.FileInfo, error)
	GetFileLink(ctx context.Context, fileID string) (string, error)
	GetFileInfosForPost(ctx context.Context, postID string) ([]*mattermost.FileInfo, error)
	CreatePost(ctx context.Context, post *mattermost.Post) (*mattermost.Post, error)
	GetPostThreadWithOpts(ctx context.Context, postID string, opts mattermost.GetPostsOptions) (*mattermost.PostList, error)
}

type EmployeeInfoStore interface {
	GetEmployeeName(ctx context.Context, employeeID int64) (string, error)
}
