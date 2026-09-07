package models

type ThreadMessagesPage struct {
	MessagesOrder  []string
	Messages       map[string]ChatMessage
	HasMore        bool
	Before         string
	BeforeCreateAt string
}
