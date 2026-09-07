package models

type ProdTypes struct {
	Data []ProdType `json:"data"`
}

type ProdType struct {
	ProdtypeCode            string `json:"prodtype_code"`
	ProdTypePartID          int64  `json:"prodtype_part_id"`
	IsForTarification       bool   `json:"is_for_tarification"`
	IsCredit                bool   `json:"is_credit"`
	IsUseImproverMultiplier bool   `json:"is_use_improver_multiplier"`
	IsDeleted               bool   `json:"is_deleted"`
	ProdtypeID              int64  `json:"prodtype_id"`
	ProdtypeName            string `json:"prodtype_name"`
}

type ProdTypeParts struct {
	Data []ProdTypePart `json:"data"`
}

type ProdTypePart struct {
	EmployeeID       int64  `json:"employee_id"`
	ProdTypePartID   int64  `json:"prodtypepart_id"`
	ProdTypePartName string `json:"prodtypepart_name"`
	Dt               string `json:"dt"`
	IsDeleted        bool   `json:"is_deleted"`
}

type SelectorProdType struct {
	ProdTypeID   int64  `json:"id"`
	ProdTypeName string `json:"name"`
}
