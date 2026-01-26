package model

type Project struct {
	Base
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
}
