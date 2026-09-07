package models

type BusinessProcesses struct {
	Data []BusinessProcess `json:"data"`
}
type BusinessProcess struct {
	ProcessID   string `json:"process_id"`
	ProcessName string `json:"process_name"`
}
