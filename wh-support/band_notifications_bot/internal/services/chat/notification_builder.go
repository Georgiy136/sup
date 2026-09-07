package chat

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/consts"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	chatmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/chat/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/employee_tags"
)

const (
	messagePreviewLen   = 100
	attachmentColor     = "#2E8B57"
	UnknownEmployeeName = "Неизвестный сотрудник"

	formatNewChatMessage = "💬 Новое сообщение в чате по заявке *№%d*"
	formatTaggedInChat   = "🔔 Вас упомянули в чате по заявке *№%d*"
	formatBandUsername   = "@%s"
)

var (
	formatChatLink = "[Открыть обсуждение в Wh Support](" + consts.SupportSiteURL + "/my-tickets/%d?mode=my_tickets_active&tab=discussion)"
)

func BuildNotificationMessage(ticketID int64) string {
	return fmt.Sprintf(formatNewChatMessage, ticketID)
}

func BuildTagNotificationMessage(ticketID int64) string {
	return fmt.Sprintf(formatTaggedInChat, ticketID)
}

func FormatBandUsername(username string) string {
	return fmt.Sprintf(formatBandUsername, username)
}

func BuildAttachment(p chatmodels.BandAttachmentParams) models.BandAttachment {
	preview := employee_tags.ReplaceTaggedEmployeeIDs(
		messagePreview(p.Message),
		p.EmployeeLabels,
		p.DefaultEmployeeLabel,
	)

	return models.BandAttachment{
		Color:      attachmentColor,
		AuthorName: formatAuthor(p.SenderEmployeeName, p.CreatedDt),
		Text:       preview + "\n\n" + fmt.Sprintf(formatChatLink, p.TicketID),
	}
}

func messagePreview(msg string) string {
	parts := strings.Split(msg, " ")
	if len(parts) == 0 {
		return msg
	}

	var b strings.Builder
	length := utf8.RuneCountInString(parts[0])
	b.WriteString(parts[0])

	for _, part := range parts[1:] {
		partLen := utf8.RuneCountInString(part)
		if length+1+partLen > messagePreviewLen {
			return b.String() + "…"
		}
		b.WriteString(" ")
		b.WriteString(part)
		length += 1 + partLen
	}
	return b.String()
}

func formatAuthor(name, createdDt string) string {
	t, err := time.Parse(time.RFC3339Nano, createdDt)
	if err != nil {
		logrus.Errorf("can't parse created_dt %s: %v", createdDt, err)
		return name
	}
	return fmt.Sprintf("%s · %s", name, t.Format("15:04"))
}
