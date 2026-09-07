package models

type TranslationGetByKeyRequest struct {
	Lang           string           `json:"lang"`
	KeysWithValues []KeysWithValues `json:"keys_with_values"`
}

type KeysWithValues struct {
	KeyId         string         `json:"key_id"`
	MessageValues map[string]any `json:"message_values"`
}

type TranslationGetByKeyResponse struct {
	Data []TranslationData `json:"data"`
}

type TranslationData struct {
	LocalizationkeyId   string `json:"localizationkey_id"`
	LocalizationLang    string `json:"localization_lang"`
	LocalizationMessage string `json:"localization_message"`
}
