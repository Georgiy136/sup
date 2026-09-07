package handlerutils

import (
	"fmt"
	"strings"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

func DecodeMapToStructureWithErrorUnset(src map[string]interface{}, ext interface{}) error {
	decodeConfig := &mapstructure.DecoderConfig{
		ErrorUnset: true,
		Result:     &ext,
	}

	decoder, err := mapstructure.NewDecoder(decodeConfig)
	if err != nil {
		return fmt.Errorf("can't create decoder err: %w", err)
	}

	err = decoder.Decode(src)
	if err != nil {
		logrus.Infof("err ext: %+v", ext)
		return fmt.Errorf("can't decode ext err: %w", err)
	}

	return nil
}

func DecodeMapToStructureWithoutErrorUnset(src map[string]interface{}, ext interface{}) error {
	decodeConfig := &mapstructure.DecoderConfig{
		ErrorUnset: false,
		Result:     &ext,
	}

	decoder, err := mapstructure.NewDecoder(decodeConfig)
	if err != nil {
		return fmt.Errorf("can't create decoder err: %w", err)
	}

	err = decoder.Decode(src)
	if err != nil {
		logrus.Infof("err ext: %+v", ext)
		return fmt.Errorf("can't decode ext err: %w", err)
	}

	return nil
}

func GenSeqFromInterval(start, end int64) ([]int64, error) {
	if start > end {
		return nil, fmt.Errorf("invalid street section start: %d > %d", start, end)
	}

	sections := make([]int64, 0, end-start+1)

	for sec := start; sec <= end; sec++ {
		sections = append(sections, sec)
	}

	return sections, nil
}

func GroupTicketsByStatuses(tickets []models.TicketCommonInfo, statuses []string) (map[string][]models.TicketCommonInfo, map[int64]error) {
	groupsTickets := make(map[string][]models.TicketCommonInfo, len(statuses))

	ticketsErr := make(map[int64]error)

	for _, status := range statuses {
		groupsTickets[status] = make([]models.TicketCommonInfo, 0)
	}

	for _, ticket := range tickets {
		if _, ok := groupsTickets[ticket.StatusID]; !ok {
			ticketsErr[ticket.TicketID] = fmt.Errorf("ticket %d by category %d has unknown status: %s", ticket.TicketID, ticket.CategoryID, ticket.StatusID)
			continue
		}

		groupsTickets[ticket.StatusID] = append(groupsTickets[ticket.StatusID], ticket)
	}

	return groupsTickets, ticketsErr
}

func MergeMapsOverwrite[K comparable, V any](dst, src map[K]V) map[K]V {
	if dst == nil {
		return src
	}

	for k, v := range src {
		if _, exists := dst[k]; exists {
			logrus.Errorf("duplicate key while merge map: %v", k)
		}
		dst[k] = v
	}

	return dst
}

func ExtractValueFromText(text, startValue, endValue string) string {
	start := strings.Index(text, startValue)

	if start == -1 {
		return ""

	}
	start += len(startValue)
	rest := text[start:]
	end := strings.Index(rest, endValue)

	if end == -1 {
		return ""
	}
	return rest[:end]
}
