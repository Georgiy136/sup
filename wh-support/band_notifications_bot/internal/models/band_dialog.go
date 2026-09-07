package models

type BandInteractiveDialogConfig struct {
	CallbackID       string
	Title            string
	IntroductionText string
	SubmitLabel      string
	Path             string
	Elements         []BandDialogElement
}

type BandDialogElement struct {
	DisplayName  string                     `json:"display_name"`
	Name         string                     `json:"name"`
	Type         BandDialogElementType      `json:"type"`
	Placeholder  string                     `json:"placeholder"`
	Required     bool                       `json:"required"`
	DefaultValue string                     `json:"default_value"`
	Options      []*BandDialogElementOption `json:"options"`
}

type BandDialogElementOption struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

type BandDialogElementType string

const (
	BandDialogElementTypeText     BandDialogElementType = "text"
	BandDialogElementTypeTextarea BandDialogElementType = "textarea"
	BandDialogElementTypeSelect   BandDialogElementType = "select"
	BandDialogElementTypeBool     BandDialogElementType = "bool"
)
