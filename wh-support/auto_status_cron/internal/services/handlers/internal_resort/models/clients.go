package internalresortmodels

type RequestDataForGoodsValidation struct {
	GoodsIDs []int64 `json:"goods_ids"`
}

type ResponseDataForGoodsValidation struct {
	Comment  string  `json:"comment"`
	GoodsIDs []int64 `json:"goods_ids"`
}

type RequestCreateInternalOrder struct {
	GoodsIDs       []int64  `json:"shk_ids"`
	NmID           int64    `json:"nm_id"`
	ContragentCode string   `json:"contragent_code"`
	ContragentID   int64    `json:"contragent_id"`
	Comment        string   `json:"comment"`
	PhotoURLs      []string `json:"photo_urls,omitempty"`
}

type ResponseCreateInternalOrder struct {
	GoodsIDsErrProcessing []GoodsIDErr `json:"shk_ids_err_processing"`
}

type GoodsIDErr struct {
	GoodsID int64  `json:"shk_id"`
	Error   string `json:"error"`
}
