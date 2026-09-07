package models

type AuthTokenParams struct {
	AuthTokenTTLDuration string `json:"auth_token_ttl_duration"`
	RefreshWindow        string `json:"refresh_window"`
}
type TicketAccessInfo struct {
	ChatID               string  `json:"chat_id"`
	TicketID             int64   `json:"ticket_id"`
	CategoryID           int64   `json:"category_id"`
	CreateEmployeeID     int64   `json:"create_employee_id"`
	GroupEmployeeIDs     []int64 `json:"group_employee_ids"`
	FavouriteEmployeeIDs []int64 `json:"favourite_employee_ids"`
}

type GenerateTokenResponse struct {
	Token string `json:"token"`
}

// формат ошибки для всех Centrifugo proxy endpoints https://centrifugal.dev/docs/server/proxy#proxy-error-response
type CentrifugoErrorResponse struct {
	Error CentrifugoError `json:"error"`
}

// структура ошибки Centrifugo https://centrifugal.dev/docs/server/proxy#error
type CentrifugoError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
