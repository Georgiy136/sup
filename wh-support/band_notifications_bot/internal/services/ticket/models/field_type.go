package models

type FieldDataType string

const (
	FieldTypeSelector FieldDataType = "selector"
	FieldTypeDate     FieldDataType = "date"
	FieldTypeBoolean  FieldDataType = "boolean"
	FieldTypeNumber   FieldDataType = "number"
	FieldTypeString   FieldDataType = "string"
)
