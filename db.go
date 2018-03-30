package cddadb

import (
	"github.com/jmoiron/sqlx"
)

type DB struct {
	*sqlx.DB
}

func (db *DB) Open(connectionString string) error {
	d, err := sqlx.Open("postgres", connectionString)
	if err != nil {
		return err
	}
	db.DB = d
	return nil
}

func (db *DB) GetItems() ([]*Item, error) {
	items := []*Item{}
	err := db.Select(&items, `
		select 
			*
		from 
			item
	`)
	if err != nil {
		return nil, err
	}
	return items, nil
}
