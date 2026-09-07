package models

type BandAttachment struct {
	Color      string
	AuthorName string
	Text       string
	Actions    []BandAction
}

type BandAction struct {
	ID      string
	Name    string
	Style   string
	Path    string
	Context map[string]any
}
