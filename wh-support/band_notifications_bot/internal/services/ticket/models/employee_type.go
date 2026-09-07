package models

type TypeOfEmployee string

const (
	TypeOfEmployeeCreator  TypeOfEmployee = "CRT" // Создатель заявки
	TypeOfEmployeeGroup    TypeOfEmployee = "GRP" // Группа исполнения или подтверждения
	TypeOfEmployeeEmployee TypeOfEmployee = "EMP" // Сотрудник, работающий с заявкой
	TypeOfEmployeeObserver TypeOfEmployee = "FVR" // Наблюдатель заявки
)

func (t TypeOfEmployee) IsCreator() bool {
	return t == TypeOfEmployeeCreator
}

func (t TypeOfEmployee) IsGroup() bool {
	return t == TypeOfEmployeeGroup
}

func (t TypeOfEmployee) IsEmployee() bool {
	return t == TypeOfEmployeeEmployee
}

func (t TypeOfEmployee) IsObserver() bool {
	return t == TypeOfEmployeeObserver
}

func (t TypeOfEmployee) IsEmpty() bool {
	return t == ""
}

func (t TypeOfEmployee) String() string {
	return string(t)
}
