package cddadb

type Item struct {
	ID   string `json:"id" db:"id"`
	Type string `json:"type" db:"type"`
	Name string `json:"name" db:"name"`
}
