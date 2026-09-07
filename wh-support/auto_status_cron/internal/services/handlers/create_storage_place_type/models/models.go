package createstorageplacetypemodels

type ExtTicketInfoForCreateStoragePlaceType struct {
	PlaceTypeName string `mapstructure:"place_type_name"`
	StickerType   string `mapstructure:"sticker_type"`
	StickerPrefix struct {
		ID string `mapstructure:"id"`
	} `mapstructure:"sticker_prefix"`
}

type RequestForCreateStoragePlaceType struct {
	PlaceTypeName string `json:"place_type_name"`
	EmployeeID    int64  `json:"employee_id"`
	LocLang       string `json:"loc_lang"`
	PlaceTypeID   *int64 `json:"place_type_id"`
	StickerType   string `json:"sticker_type"`
	StickerPrefix string `json:"sticker_prefix"`
	IsDel         bool   `json:"is_del"`
}

type ResponseDataFromCreateStoragePlaceTypeApi struct {
	Data StoragePlaceType `json:"data"`
}

type StoragePlaceType struct {
	PlaceTypeID int64 `json:"place_type_id"`
}

type AddedPerformInfoStoragePlaceType struct {
	PlaceTypeID int64 `json:"place_type_id"`
}
