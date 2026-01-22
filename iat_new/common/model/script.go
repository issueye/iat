package model

type Script struct {
	Base
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"` // JS code
}
