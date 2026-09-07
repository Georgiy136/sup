package internalresortmodels

type ExtTicketInfoCheckUploadedGoodsIDs struct {
	GoodsIDs []ValidGoodsID `mapstructure:"goods_ids"`
	OfficeId struct {
		ID   int64  `mapstructure:"id"`
		Name string `mapstructure:"name"`
	} `mapstructure:"office_id"`
}

type ExtTicketInfoSendTaskToProfileTeamBySeller struct {
	NmId struct {
		ID   int64  `mapstructure:"id"`
		Name string `mapstructure:"name"`
		NmID int64  `mapstructure:"nm_id"`
	} `mapstructure:"nm_id"`
	Proof    string `mapstructure:"proof"`
	NmIDNew  int64  `mapstructure:"nm_id_new"`
	OfficeId struct {
		ID   int64  `mapstructure:"id"`
		Name string `mapstructure:"name"`
	} `mapstructure:"office_id"`
	ValidGoodsIDs []ValidGoodsID `mapstructure:"valid_shk"`
	URLs          []struct {
		URL URL `mapstructure:"url"`
	} `mapstructure:"urls"`
}

type ExtTicketInfoSendTaskToProfileTeamByEmployee struct {
	Proof    string `mapstructure:"proof"`
	NmIDNew  int64  `mapstructure:"nm_id_new"`
	OfficeId struct {
		ID   int64  `mapstructure:"id"`
		Name string `mapstructure:"name"`
	} `mapstructure:"office_id"`
	ValidGoodsIDs []ValidGoodsID `mapstructure:"valid_shk"`
	EmployeeID    struct {
		ID   int64  `mapstructure:"id"`
		Name string `mapstructure:"name"`
	} `mapstructure:"employee_id"`
	URLs []struct {
		URL URL `mapstructure:"url"`
	} `mapstructure:"urls"`
}

type InfoForSendTaskToProfileTeam struct {
	Proof          string
	ContragentCode string
	ContragentID   int64
	NmIDNew        int64
	ValidGoodsIDs  []ValidGoodsID
	URLs           []struct {
		URL URL `mapstructure:"url"`
	} `mapstructure:"urls"`
}

type PerformInfoCheckUploadedGoodsIDs struct {
	ValidGoodsIDs    []ValidGoodsID    `mapstructure:"valid_shk" json:"valid_shk"`
	NotValidGoodsIDs []NotValidGoodsID `mapstructure:"not_valid_shk" json:"not_valid_shk"`
}

type PerformInfoSendTaskToProfileTeam struct {
	NotValidGoodsIDs []NotValidGoodsID `mapstructure:"error_shk" json:"error_shk"`
}

type ValidGoodsID struct {
	GoodsID GoodsID `mapstructure:"shk_id" json:"shk_id"`
}

type NotValidGoodsID struct {
	Comment  Comment  `mapstructure:"comment" json:"comment"`
	GoodsIDs GoodsIDs `mapstructure:"good_ids" json:"good_ids"`
}

type GoodsID struct {
	Value         int64  `mapstructure:"value" json:"value"`
	OrderID       int64  `mapstructure:"order_id" json:"order_id"`
	FrontDataName string `mapstructure:"front_data_name" json:"front_data_name"`
}

type Comment struct {
	Value         string `mapstructure:"value" json:"value"`
	OrderID       int64  `mapstructure:"order_id" json:"order_id"`
	FrontDataName string `mapstructure:"front_data_name" json:"front_data_name"`
}

type GoodsIDs struct {
	Value         []int64 `mapstructure:"value" json:"value"`
	OrderID       int64   `mapstructure:"order_id" json:"order_id"`
	FrontDataName string  `mapstructure:"front_data_name" json:"front_data_name"`
}

type URL struct {
	Value         string `mapstructure:"value" json:"value"`
	OrderID       int64  `mapstructure:"order_id" json:"order_id"`
	FrontDataName string `mapstructure:"front_data_name" json:"front_data_name"`
}
