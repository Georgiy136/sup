package models

type StickerPrefixes struct {
	Data []StickerPrefix `json:"data"`
}

type StickerPrefix struct {
	StickerPrefix      string `json:"sticker_prefix"`
	StickerDescription string `json:"sticker_description"`
}
