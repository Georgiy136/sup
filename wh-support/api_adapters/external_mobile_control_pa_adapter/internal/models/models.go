package models

type DeviceInfo struct {
	DeviceTypeID       string `json:"device_type_id"`
	DeviceSerialNumber string `json:"device_serial_number"`
}

type CheckDeviceResponse struct {
	DevicesInfo []DeviceInfo `json:"value"`
	IsValid     bool         `json:"is_valid"`
	ErrMsg      string       `json:"err_msg"`
}
