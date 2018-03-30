package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

func init() {
	f := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(f)
}

func main() {
	flag.Parse()

	db, err := sqlx.Open("postgres", "postgres://cddadb:cddadb@localhost:9432/cddadb?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	txn, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := txn.Prepare(pq.CopyIn("item", "id", "type", "source", "raw"))
	if err != nil {
		log.Fatal(err)
	}

	itemsRoot := "f:/code/cpp/Cataclysm-DDA/data/json/items"
	files := []string{}

	err = filepath.Walk(itemsRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.Name() == "obsolete.json" {
			return nil
		}
		files = append(files, path)
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fmt.Printf("processing file: %s\n", f)

		itemText, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatal(err)
		}

		var data []map[string]interface{}
		err = json.Unmarshal(itemText, &data)
		if err != nil {
			log.Fatal(err)
		}

		for _, d := range data {
			raw, err := json.Marshal(d)
			if err != nil {
				log.Fatal(err)
			}

			//fmt.Printf("%s\n", string(raw))
			_, err = stmt.Exec(d["id"], d["type"], f, string(raw))
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Fatal(err)
	}

	err = stmt.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = txn.Commit()
	if err != nil {
		log.Fatal(err)
	}

}
