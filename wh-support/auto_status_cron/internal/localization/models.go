package localization

// LocalizedErrors список локализованных ошибок в новом формате
type LocalizedErrors struct {
	KeysWithValues []LocalizationKey `json:"keys_with_values"`
}

// LocalizationKey ключ локализации с параметрами
type LocalizationKey struct {
	KeyID         string            `json:"key_id"`
	MessageValues map[string]string `json:"message_values"`
}

// Add добавляет ошибку в список
func (s *LocalizedErrors) Add(keyID string, values map[string]string) {
	s.KeysWithValues = append(s.KeysWithValues, LocalizationKey{
		KeyID:         keyID,
		MessageValues: values,
	})
}

// New создаёт список с одной ошибкой
func New(keyID string, values map[string]string) *LocalizedErrors {
	errs := &LocalizedErrors{}
	errs.Add(keyID, values)
	return errs
}
