package cddadb

type Item struct {
	ID       string `json:"id" db:"id"`
	Abstract string `json:"abstract" db:"abstract"`
	Type     string `json:"type" db:"type"`
}
