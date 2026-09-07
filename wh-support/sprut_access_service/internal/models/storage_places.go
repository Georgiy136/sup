package models

type SprutStoragePlacesResponse struct {
	Data SprutStoragePlacesData `json:"data"`
}

type SprutStoragePlacesData struct {
	Sections []SprutSection `json:"sections"`
}

type SprutSection struct {
	Section int64                `json:"section"`
	Places  []SprutStoragePlaces `json:"places"`
}

type SprutStoragePlaces struct {
	PlaceID   int64  `json:"place_id"`
	PlaceName string `json:"place_name"`
}

type StoragePlace struct {
	PlaceID   int64  `json:"place_id"`
	PlaceName string `json:"place_name"`
}
