package main

import (
	"flag"

	"github.com/ralreegorganon/cddadb/cmd/cddadb-map/metadata"
	"github.com/ralreegorganon/cddadb/cmd/cddadb-map/overmap"
	"github.com/ralreegorganon/cddadb/cmd/cddadb-map/rasterize"
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

	jsonRoot := "/Users/jj/code/Cataclysm-DDA/data/json"
	m := metadata.NewOvermap()
	err := m.BuildUpForJsonRoot(jsonRoot)
	if err != nil {
		log.Fatal(err)
	}

	save := "/Users/jj/code/Cataclysm-DDA/save/Grantsburg"
	o, err := overmap.FromSave(save)
	if err != nil {
		log.Fatal(err)
	}

	err = o.RenderToFilesAlt(m, "/Users/jj/Desktop/Grantsburg")
	if err != nil {
		log.Fatal(err)
	}

	err = rasterize.Blam("/Users/jj/Desktop/Grantsburg/o_10")
	if err != nil {
		log.Fatal(err)
	}
}
