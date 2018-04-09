package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
)

type Overmap struct {
	templates map[string]overmapTerrain
	built     map[string]overmapTerrain
	symbols   map[int]string
	rotations [][]int
}

type overmapTerrain struct {
	ID         string   `json:"id"`
	Type       string   `json:"type"`
	Abstract   string   `json:"abstract"`
	Name       string   `json:"name"`
	Sym        int      `json:"sym"`
	Color      string   `json:"color"`
	CopyFrom   string   `json:"copy-from"`
	SeeCost    int      `json:"see_cost"`
	Extras     string   `json:"extras"`
	MonDensity int      `json:"mondensity"`
	Flags      []string `json:"flags"`
	Spawns     spawns   `json:"spawns"`
	MapGen     []mapGen `json:"mapgen"`
}

type spawns struct {
	Group      string `json:"group"`
	Population []int  `json:"population"`
	Chance     int    `json:"chance"`
}

type mapGen struct {
	Method string `json:"method"`
	Name   string `json:"name"`
}

const overmapTerrainTypeID = "overmap_terrain"

type inLoadOrder []string

func (s inLoadOrder) Len() int {
	return len(s)
}

func (s inLoadOrder) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s inLoadOrder) Less(i, j int) bool {
	c1 := strings.Count(s[i], "/")
	c2 := strings.Count(s[j], "/")

	if c1 == c2 {
		return s[i] < s[j]
	}
	return c1 < c2
}

func indexOf(slice []int, item int) int {
	for i := range slice {
		if slice[i] == item {
			return i
		}
	}
	return -1
}

var linearSuffixes = []string{
	"_isolated",
	"_end_south",
	"_end_west",
	"_ne",
	"_end_north",
	"_ns",
	"_es",
	"_nes",
	"_end_east",
	"_wn",
	"_ew",
	"_new",
	"_sw",
	"_nsw",
	"_esw",
	"_nesw"}

var linearSuffixSymbols = map[string]int{
	"_isolated":  0,
	"_end_south": 4194424,
	"_end_west":  4194417,
	"_ne":        4194413,
	"_end_north": 4194424,
	"_ns":        4194424,
	"_es":        4194412,
	"_nes":       4194420,
	"_end_east":  4194417,
	"_wn":        4194410,
	"_ew":        4194417,
	"_new":       4194422,
	"_sw":        4194411,
	"_nsw":       4194421,
	"_esw":       4194423,
	"_nesw":      4194414,
}

var rotationSuffixes = []string{
	"_north",
	"_east",
	"_south",
	"_west"}

func NewOvermap() *Overmap {
	symbols := map[int]string{
		4194424: "\u2502",
		4194417: "\u2500",
		4194413: "\u2514",
		4194412: "\u250c",
		4194411: "\u2510",
		4194410: "\u2518",
		4194420: "\u251c",
		4194422: "\u2534",
		4194421: "\u2524",
		4194423: "\u252c",
		4194414: "\u253c",
	}

	for i := 0; i < 128; i++ {
		symbols[i] = string(i)
	}

	rotations := make([][]int, 0)
	rotations = append(rotations, []int{60, 94, 62, 118})
	rotations = append(rotations, []int{4194410, 4194413, 4194412, 4194411})
	rotations = append(rotations, []int{4194417, 4194424, 4194417, 4194424})
	rotations = append(rotations, []int{4194420, 4194423, 4194421, 4194422})

	l := &Overmap{
		templates: make(map[string]overmapTerrain),
		built:     make(map[string]overmapTerrain),
		symbols:   symbols,
		rotations: rotations,
	}
	return l
}

func (o *Overmap) BuildUp(jsonRoot, modsRoot string) error {
	files, err := sourceFiles(jsonRoot, modsRoot)
	if err != nil {
		return err
	}

	err = o.loadTemplatesFromFiles(files)
	if err != nil {
		return err
	}

	err = o.buildTemplates()
	if err != nil {
		return err
	}

	return nil
}

func sourceFiles(jsonRoot, modsRoot string) ([]string, error) {
	files := []string{}

	err := filepath.Walk(jsonRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".json") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// need to filter to active mods for save
	// err = filepath.Walk(modsRoot, func(path string, info os.FileInfo, err error) error {
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if info.IsDir() {
	// 		return nil
	// 	}
	// 	if strings.HasSuffix(path, ".json") {
	// 		files = append(files, path)
	// 	}
	// 	return nil
	// })
	// if err != nil {
	// 	return nil, err
	// }

	sort.Sort(inLoadOrder(files))

	return files, nil
}

func (o *Overmap) loadTemplates(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	if !bytes.Contains(b, []byte(overmapTerrainTypeID)) {
		return nil
	}

	var temp []map[string]interface{}
	err = json.Unmarshal(b, &temp)
	if err != nil {
		return err
	}

	filteredOvermapTerrains := make([]map[string]interface{}, 0)
	for _, t := range temp {
		if t["type"].(string) == overmapTerrainTypeID {
			filteredOvermapTerrains = append(filteredOvermapTerrains, t)
		}
	}

	filteredText, err := json.Marshal(filteredOvermapTerrains)
	if err != nil {
		return err
	}

	var overmapTerrains []overmapTerrain
	err = json.Unmarshal(filteredText, &overmapTerrains)
	if err != nil {
		return err
	}

	for _, ot := range overmapTerrains {
		if ot.Type != overmapTerrainTypeID {
			continue
		}
		if ot.Abstract != "" {
			o.templates[ot.Abstract] = ot
		} else {
			o.templates[ot.ID] = ot
		}
	}

	return nil
}

func (o *Overmap) loadTemplatesFromFiles(files []string) error {
	for _, f := range files {
		f, err := os.Open(f)
		if err != nil {
			return err
		}
		defer f.Close()

		o.loadTemplates(f)
	}

	return nil
}

func (o *Overmap) buildTemplates() error {
	for _, ot := range o.templates {
		bt := make([]overmapTerrain, 0)
		t := ot
		bt = append(bt, t)
		for t.CopyFrom != "" {
			t = o.templates[t.CopyFrom]
			bt = append(bt, t)
		}

		b := overmapTerrain{}
		for i := len(bt) - 1; i >= 0; i-- {
			if err := mergo.Merge(&b, bt[i], mergo.WithOverride); err != nil {
				return err
			}
		}

		if ot.Abstract == "" {
			b.Abstract = ""
			b.CopyFrom = ""
			o.built[b.ID] = b

			rotate := true

			if b.Flags != nil {
				for _, f := range b.Flags {
					if f == "NO_ROTATE" {
						rotate = false
					} else if f == "LINEAR" {
						for _, suffix := range linearSuffixes {
							bs := overmapTerrain{}
							if err := mergo.Merge(&bs, b, mergo.WithOverride); err != nil {
								return err
							}
							bs.ID = b.ID + suffix
							bs.Sym = linearSuffixSymbols[suffix]
							o.built[bs.ID] = bs
						}
					}
				}
			}

			if rotate {
				for i, suffix := range rotationSuffixes {
					bs := overmapTerrain{}
					if err := mergo.Merge(&bs, b, mergo.WithOverride); err != nil {
						log.Fatal(err)
					}
					bs.ID = b.ID + suffix

					for _, r := range o.rotations {
						index := indexOf(r, b.Sym)
						if index != -1 {
							bs.Sym = r[(i+index+4)%4]
							break
						}
					}
					o.built[bs.ID] = bs
				}
			}
		}
	}

	colors := make(map[string]int)
	foo := make([]string, 0)
	for _, v := range o.built {
		if v.Sym == 82 {
			fmt.Println(v.Color)
		}

		if _, ok := colors[v.Color]; !ok {
			colors[v.Color] = 0
			foo = append(foo, v.Color)
		}
		colors[v.Color] = colors[v.Color] + 1
	}
	sort.Strings(foo)
	for _, s := range foo {
		fmt.Printf("case \"%v\":\n return white, black\n", s)
	}
	spew.Dump(colors)

	return nil
}

func (o *Overmap) Exists(id string) bool {
	_, ok := o.built[id]
	return ok
}

func (o *Overmap) Symbol(id string) string {
	if t, tok := o.built[id]; tok {
		if s, sok := o.symbols[t.Sym]; sok {
			return s
		}
	}
	return "?"
}

func (o *Overmap) Color(id string) (*image.Uniform, *image.Uniform) {
	white := image.NewUniform(color.RGBA{150, 150, 150, 255})
	black := image.NewUniform(color.RGBA{0, 0, 0, 255})
	red := image.NewUniform(color.RGBA{255, 0, 0, 255})
	green := image.NewUniform(color.RGBA{0, 110, 0, 255})
	brown := image.NewUniform(color.RGBA{92, 51, 23, 255})
	blue := image.NewUniform(color.RGBA{0, 0, 200, 255})
	magenta := image.NewUniform(color.RGBA{139, 58, 98, 255})
	cyan := image.NewUniform(color.RGBA{0, 150, 180, 255})
	gray := image.NewUniform(color.RGBA{150, 150, 150, 255})
	darkGray := image.NewUniform(color.RGBA{99, 99, 99, 255})
	lightRed := image.NewUniform(color.RGBA{255, 150, 150, 255})
	lightGreen := image.NewUniform(color.RGBA{0, 255, 0, 255})
	yellow := image.NewUniform(color.RGBA{255, 255, 0, 255})
	lightBlue := image.NewUniform(color.RGBA{100, 100, 255, 255})
	lightMagenta := image.NewUniform(color.RGBA{254, 0, 254, 255})
	lightCyan := image.NewUniform(color.RGBA{0, 240, 255, 255})

	if c, tok := o.built[id]; tok {
		switch c.Color {
		case "black_yellow":
			return black, yellow
		case "blue":
			return blue, black
		case "brown":
			return brown, black
		case "c_yellow_green":
			return yellow, green
		case "cyan":
			return cyan, black
		case "dark_gray":
			return darkGray, black
		case "dark_gray_magenta":
			return darkGray, magenta
		case "green":
			return green, black
		case "h_dark_gray":
			return darkGray, black
		case "h_yellow":
			return yellow, black
		case "i_blue":
			return black, blue
		case "i_brown":
			return black, brown
		case "i_cyan":
			return black, cyan
		case "i_green":
			return black, green
		case "i_light_blue":
			return black, lightBlue
		case "i_light_cyan":
			return black, lightCyan
		case "i_light_gray":
			return black, gray
		case "i_light_green":
			return black, lightGreen
		case "i_light_red":
			return black, lightRed
		case "i_magenta":
			return black, magenta
		case "i_pink":
			return black, lightMagenta
		case "i_red":
			return black, red
		case "i_yellow":
			return black, yellow
		case "light_blue":
			return lightBlue, black
		case "light_cyan":
			return lightCyan, black
		case "light_gray":
			return gray, black
		case "light_green":
			return lightGreen, black
		case "light_green_yellow":
			return lightGreen, yellow
		case "light_red":
			return lightRed, black
		case "magenta":
			return magenta, black
		case "pink":
			return lightMagenta, black
		case "pink_magenta":
			return lightMagenta, magenta
		case "red":
			return red, black
		case "white":
			return white, black
		case "white_magenta":
			return white, magenta
		case "white_white":
			return white, white
		case "yellow":
			return yellow, black
		case "yellow_cyan":
			return yellow, cyan
		case "yellow_magenta":
			return yellow, magenta
		default:
			return white, black
		}
	}
	return white, black
}
