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
	modsRoot := "/Users/jj/code/Cataclysm-DDA/data/mods"
	m := metadata.NewOvermap()
	err := m.BuildUp(jsonRoot, modsRoot)
	if err != nil {
		log.Fatal(err)
	}

	save := "/Users/jj/code/Cataclysm-DDA/save/Hannastown"
	// save := "/Users/jj/Downloads/trinitycenter"
	o, err := overmap.FromSave(save)
	if err != nil {
		log.Fatal(err)
	}

	w, err := o.RenderToAttributes(m)
	if err != nil {
		log.Fatal(err)
	}

	// err = w.RenderToFiles("/Users/jj/Desktop/Hannastown")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// err = rasterize.Blam2("/Users/jj/Desktop/TrinityCenter/", w)
	err = rasterize.Blam2("/Users/jj/Desktop/Hannastown", w)
	if err != nil {
		log.Fatal(err)
	}
}
