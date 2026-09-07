package utils

import (
	"fmt"
	"strconv"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/models"
)

func ExtractProps(props map[string]any) (*models.Props, error) {
	ticketID, err := getInt64(props, "ticket_id")
	if err != nil {
		return nil, err
	}
	employeeID, err := getInt64(props, "employee_id")
	if err != nil {
		return nil, err
	}
	employeeName, err := getString(props, "employee_name")
	if err != nil {
		return nil, err
	}
	return &models.Props{
		TicketID:     ticketID,
		EmployeeID:   employeeID,
		EmployeeName: employeeName,
	}, nil
}

func getString(props map[string]any, key string) (string, error) {
	value, ok := props[key]
	if !ok {
		return "", fmt.Errorf("%s is missing", key)
	}
	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%s is not a string", key)
	}
	return str, nil
}

func getInt64(props map[string]any, key string) (int64, error) {
	str, err := getString(props, key)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s parse error: %w", key, err)
	}
	return value, nil
}

func ToRFC3339Nano(unixMs int64) string {
	if unixMs <= 0 {
		return ""
	}
	return time.UnixMilli(unixMs).Format(time.RFC3339Nano)
}

func FromRFC3339NanoToUnix(value string) (int64, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return 0, fmt.Errorf("parse RFC3339Nano time: %w", err)
	}

	return parsed.UnixMilli(), nil
}
