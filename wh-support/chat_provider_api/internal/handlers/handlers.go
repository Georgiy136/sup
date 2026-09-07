package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/models"
	chat_utils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/utils"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"
)

type Handlers struct {
	chat       ChatService
	errBuilder core_errors.ErrorBuilder
}

func New(chat ChatService) *Handlers {
	return &Handlers{
		chat:       chat,
		errBuilder: core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
	}
}

type ChatService interface {
	CreateChat(ctx context.Context, employeeID, ticketID int64) (*models.CreateChatResponse, error)
	CreateChatMessage(ctx context.Context, employeeID int64, chatID, message string, ticketID int64, fileIDs []string) (*models.CreatePostResponse, error)
	GetChat(ctx context.Context, chatID string) (*models.GetChatResponse, error)
	GetChatHistory(ctx context.Context, chatID string, countPost int64, fromPost string, fromCreateAt int64) (*models.HistoryChatResponse, error)
	UploadFileStream(ctx context.Context, filename string, body io.Reader, fileSize int64) (*models.FileInfo, error)
	GetDownloadLink(ctx context.Context, fileID string) (*models.GetDownloadLinkResponse, error)
}

func (h *Handlers) CreateChat(ctx *gin.Context, params map[string]any) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		h.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), errors.New("can't parse employee_id parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID int64 `json:"ticket_id" binding:"required,gt=0"`
	})
	if !ok {
		h.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		h.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	response, err := h.chat.CreateChat(ctx.Request.Context(), employeeID, body.TicketID)
	if err != nil {
		h.handleServiceError(ctx, err)
		return
	}
	utils.BindObjectToRestData(ctx, response)
}

func (h *Handlers) CreatePost(ctx *gin.Context, params map[string]any) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		h.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), errors.New("can't parse employee_id parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		ChatID   string   `json:"chat_id" binding:"required,lte=30"`
		Message  string   `json:"message" binding:"required,lte=5000"`
		TicketID int64    `json:"ticket_id" binding:"required,gt=0"`
		FileIDs  []string `json:"file_ids" binding:"omitempty,dive,lte=30"`
	})
	if !ok {
		h.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		h.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	response, err := h.chat.CreateChatMessage(
		ctx.Request.Context(),
		employeeID,
		body.ChatID,
		body.Message,
		body.TicketID,
		body.FileIDs,
	)
	if err != nil {
		h.handleServiceError(ctx, err)
		return
	}
	utils.BindObjectToRestData(ctx, response)
}

func (h *Handlers) GetChat(ctx *gin.Context, params map[string]any) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		ChatID string `json:"chat_id" binding:"required,lte=30"`
	})
	if !ok {
		h.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		h.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	response, err := h.chat.GetChat(ctx.Request.Context(), body.ChatID)
	if err != nil {
		h.handleServiceError(ctx, err)
		return
	}
	utils.BindObjectToRestData(ctx, response)
}

func (h *Handlers) GetHistoryChat(ctx *gin.Context, params map[string]any) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		ChatID       string `json:"chat_id" binding:"required,lte=30"`
		CountPost    int64  `json:"count_post" binding:"required,gt=0,lte=200"`
		FromPost     string `json:"from_post" binding:"required,lte=30"`
		FromCreateAt string `json:"from_create_at" binding:"required,IsRFC3339Nano"`
	})
	if !ok {
		h.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		h.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	fromCreateAt, err := chat_utils.FromRFC3339NanoToUnix(body.FromCreateAt)
	if err != nil {
		h.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("invalid from_create_at: %w", err))
		return
	}

	response, err := h.chat.GetChatHistory(ctx.Request.Context(), body.ChatID, body.CountPost, body.FromPost, fromCreateAt)
	if err != nil {
		h.handleServiceError(ctx, err)
		return
	}
	utils.BindObjectToRestData(ctx, response)
}

func (h *Handlers) UploadFile(ctx *gin.Context, _ map[string]any) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		h.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("can't read file from multipart form: %w", err))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("can't open uploaded file: %w", err))
		return
	}
	response, err := h.chat.UploadFileStream(ctx.Request.Context(), fileHeader.Filename, file, fileHeader.Size)
	if err != nil {
		h.handleServiceError(ctx, err)
		return
	}
	utils.BindObjectToRestData(ctx, response)
}

func (h *Handlers) DownloadFile(ctx *gin.Context, params map[string]any) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		FileID string `json:"file_id" binding:"required,lte=30"`
	})
	if !ok {
		h.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		h.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	response, err := h.chat.GetDownloadLink(ctx.Request.Context(), body.FileID)
	if err != nil {
		h.handleServiceError(ctx, err)
		return
	}
	utils.BindObjectToRestData(ctx, response)
}

func (h *Handlers) handleServiceError(ctx *gin.Context, err error) {
	if ctx.Request.Context().Err() != nil {
		logrus.Infof("client disconnected: %v", err)
		h.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, err)
		return
	}
	if httpErr, ok := errors.AsType[*models.HTTPError](err); ok {
		switch httpErr.StatusCode {
		case http.StatusBadRequest:
			h.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, httpErr)
		case http.StatusNotFound:
			h.errBuilder.BindError(ctx, errors_keys.ErrBizDbResponseEmpty, httpErr)
		default:
			h.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), httpErr)
		}
		return
	}
	h.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), err)
}
