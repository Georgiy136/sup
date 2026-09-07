package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// Извлекает ticket_id из названия канала
// Поддерживаемые форматы:
//   - ticket:12345
//   - ticket:12345:chat
//   - ticket:12345:updates
func ExtractTicketIDFromChannel(channel string) (int64, error) {
	if !strings.HasPrefix(channel, "ticket:") {
		return 0, fmt.Errorf("channel must start with 'ticket:', got: %s", channel)
	}

	withoutPrefix := strings.TrimPrefix(channel, "ticket:")

	parts := strings.Split(withoutPrefix, ":")
	if len(parts) == 0 || parts[0] == "" {
		return 0, fmt.Errorf("ticket_id is missing in channel: %s", channel)
	}

	ticketID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid ticket_id in channel %s: %w", channel, err)
	}

	if ticketID <= 0 {
		return 0, fmt.Errorf("ticket_id must be positive, got: %d", ticketID)
	}

	return ticketID, nil
}
