package closewh

// Статусы тикетов
const (
	statusIdDeactivateWh       = "UPD"
	statusIdGenerateSignDocGD1 = "GD1"
	statusIdGenerateSignDocGD2 = "GD2"

	statusIdDeleteStoragePlaceDLP = "DLP"
	statusIdDeleteStoragePlacePL1 = "PL1"
	statusIdDeleteStoragePlacePL2 = "PL2"
	statusIdDeleteStoragePlacePL3 = "PL3"
	statusIdDeleteStoragePlacePL4 = "PL4"

	statusIdGetGoodsUP1 = "UP1"
	statusIdGetGoodsUP2 = "UP2"

	statusIdCloseWhDl1 = "DL1"
	statusIdCloseWhDl2 = "DL2"
	statusIdCloseWhDl3 = "DL3"
	statusIdCloseWhDl4 = "DL4"
	statusIdCloseWhDl5 = "DL5"

	statusIdUploadStoragePlacesLD1 = "LD1"
	statusIdUploadStoragePlacesLD2 = "LD2"
)

// Лимиты
const (
	goodsMaxTotal         = 25000
	goodsMaxTotalPriceSum = 50000.0
)

// Типы офисов и сотрудники для документа подписи
const (
	officeTypeSC = 0 // Складской комплекс
	officeTypeSR = 1 // Сортировочный центр

	officeTypeNameSK = "СК"
	officeTypeNameSC = "СЦ"

	employeeNameSKEmployee         = "Фадеев Вячеслав Владимирович"
	employeeNameSKEmployeeGenitive = "Фадеева Вячеслава Владимировича"
	employeeNameSCEmployee         = "Бусарев Евгений Александрович"
	employeeNameSCEmployeeGenitive = "Бусарева Евгения Александровича"
)
