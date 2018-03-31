package main

import (
	"encoding/json"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
)

type Item struct {
	ID       string `json:"id"`
	Abstract string `json:"abstract"`
	Type     string `json:"type"`
	CopyFrom string `json:"copy-from"`
	Raw      json.RawMessage
}

func DoIt(sources map[string][]byte) error {
	templates := make(map[string]map[string]interface{})
	for _, t := range sources {
		var data []map[string]interface{}
		err := json.Unmarshal(t, &data)
		if err != nil {
			log.Fatal(err)
		}

		for _, d := range data {
			if keyExists(d, "abstract") {
				templates[d["abstract"].(string)] = d
			} else if keyExists(d, "id") {
				templates[d["id"].(string)] = d
			} else {
				log.Fatal("no id or abstract")
			}
		}
	}

	for id, t := range templates {
		itemType := t["type"].(string)
		switch itemType {
		case "TOOLMOD":
			buildToolmod(t[id].(string), templates)
		}
	}
	spew.Dump(templates["mod_battery"])

	return nil
}

func buildToolmod(id string, templates map[string]map[string]interface{}) {

}

func keyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}
