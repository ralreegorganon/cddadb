package overmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/ralreegorganon/cddadb/cmd/cddadb-map/metadata"
)

type Overmap struct {
	Chunks []OvermapChunk
}

type OvermapChunk struct {
	X      int
	Y      int
	Layers [][]TerrainGroup `json:"layers"`
	//RegionID string           `json:"region_id"`
	//MonsterGroups   string `json:"monster_groups"`
	//Cities          string `json:"cities"`
	//RoadsOut        string `json:"roads_out"`
	//Radios          string `json:"radios"`
	//MonsterMap      string `json:"monster_map"`
	//TrackedVehicles string `json:"tracked_vehicles"`
	//ScentTraces     string `json:"scent_traces"`
	//NPCs            string `json:"npcs"`
}

type TerrainGroup struct {
	OvermapTerrainID string
	Count            float64
}

func (tg *TerrainGroup) UnmarshalJSON(bs []byte) error {
	arr := []interface{}{}
	json.Unmarshal(bs, &arr)
	tg.OvermapTerrainID = arr[0].(string)
	tg.Count = arr[1].(float64)
	return nil
}

func keyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}

func indexOf(slice []int, item int) int {
	for i := range slice {
		if slice[i] == item {
			return i
		}
	}
	return -1
}

func overmapChunkFiles(root string) ([]string, error) {
	files := []string{}
	re := regexp.MustCompile(`o\.-?\d\.-?\d$`)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		isOvermapChunk := re.MatchString(path)
		if isOvermapChunk {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func chunkFileNameToCoordinates(chunkFile string) (int, int, error) {
	_, file := filepath.Split(chunkFile)
	parts := strings.Split(file, ".")
	x, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	y, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, err
	}
	return x, y, nil
}

func FromSave(save string) (*Overmap, error) {
	chunkFiles, err := overmapChunkFiles(save)
	if err != nil {
		return nil, err
	}

	chunks := make([]OvermapChunk, 0)

	for _, f := range chunkFiles {
		t, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, err
		}

		lines := strings.Split(string(t), "\n")

		if lines[0] != "# version 26" {
			return nil, fmt.Errorf("unsupported version: %v", lines[0])
		}

		var buffer bytes.Buffer
		for i := 1; i < len(lines); i++ {
			buffer.WriteString(lines[i])
		}

		pruned := buffer.Bytes()

		var chunk OvermapChunk
		err = json.Unmarshal(pruned, &chunk)
		if err != nil {
			return nil, err
		}

		x, y, err := chunkFileNameToCoordinates(f)
		if err != nil {
			return nil, err
		}
		chunk.X = x
		chunk.Y = y

		chunks = append(chunks, chunk)
	}

	o := &Overmap{
		Chunks: chunks,
	}
	return o, nil
}

func (o *Overmap) RenderToFiles(m *metadata.Overmap, root string) error {
	err := os.MkdirAll(root, os.ModePerm)
	if err != nil {
		return err
	}

	for _, c := range o.Chunks {
		for _, l := range c.Layers {
			for _, e := range l {
				if exists := m.Exists(e.OvermapTerrainID); !exists {
					return fmt.Errorf("couldn't find terrain: %s", e.OvermapTerrainID)
				}
			}
		}
	}

	for _, c := range o.Chunks {
		textMap := [21][180][180]string{}
		for li, l := range c.Layers {
			lmi := 0
			linearMap := [32400]string{}
			for _, e := range l {
				s := m.Symbol(e.OvermapTerrainID)
				for i := 0; i < int(e.Count); i++ {
					linearMap[lmi] = s
					lmi++
				}
			}

			for x := 0; x < 180; x++ {
				for y := 0; y < 180; y++ {
					textMap[li][x][y] = linearMap[x*180+y]
				}
			}

			rows := [180]string{}
			for i := 0; i < 180; i++ {
				rows[i] = strings.Join(textMap[li][i][:], "")
			}
			layer := strings.Join(rows[:], "\n")

			filename := filepath.Join(root, fmt.Sprintf("o.%v.%v.%v", c.X, c.Y, li))
			f, err := os.Create(filename)
			if err != nil {
				return err
			}
			defer f.Close()
			f.WriteString(layer)
		}
	}
	return nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

type World struct {
	Layers []WorldLayer
}

type WorldLayer struct {
	Rows []WorldRow
}

type WorldRow struct {
	Cells []WorldCell
}

type WorldCell struct {
	Symbol string
}

func (o *Overmap) RenderToAttributes(m *metadata.Overmap) (*World, error) {
	for _, c := range o.Chunks {
		for _, l := range c.Layers {
			for _, e := range l {
				if exists := m.Exists(e.OvermapTerrainID); !exists {
					return nil, fmt.Errorf("couldn't find terrain: %s", e.OvermapTerrainID)
				}
			}
		}
	}

	cXMax := 0
	cXMin := 0
	cYMax := 0
	cYMin := 0

	for _, c := range o.Chunks {
		if c.X > cXMax {
			cXMax = c.X
		}
		if c.Y > cYMax {
			cYMax = c.Y
		}
		if c.X < cXMin {
			cXMin = c.X
		}
		if c.Y < cYMin {
			cYMin = c.Y
		}
	}

	cXSize := abs(cXMax) + abs(cXMin) + 1
	cYSize := abs(cYMax) + abs(cYMin) + 1

	chunkCapacity := cXSize * cYSize

	doneChunks := make(map[int]bool)
	cells := make([]WorldCell, 680400*chunkCapacity)
	for _, c := range o.Chunks {
		ci := c.X + (0 - cXMin) + cXSize*(c.Y+0-cYMin)
		doneChunks[ci] = true
		fmt.Printf("processing x:%v, y:%v as %v\n", c.X, c.Y, ci)
		for li, l := range c.Layers {
			lzp := 0
			for _, e := range l {
				s := m.Symbol(e.OvermapTerrainID)
				for i := 0; i < int(e.Count); i++ {
					tmi := ci*680400 + li*32400 + lzp
					cells[tmi] = WorldCell{Symbol: s}
					lzp++
				}
			}
		}
	}

	for i := 0; i < chunkCapacity; i++ {
		if _, ok := doneChunks[i]; !ok {
			fmt.Printf("filling in blank chunk: %v\n", i)
			for e := 0; e < 680400; e++ {
				cells[i*680400+e] = WorldCell{Symbol: " "}
			}
		}
	}

	// chunk*680400 + layer*32400 + row*180 + cell
	// chunk =  xchunk + ychunk*xchunksize

	worldLayers := make([]WorldLayer, 21)
	for l := 0; l < 21; l++ {
		worldLayers[l].Rows = make([]WorldRow, 180*cYSize)
		for r := 0; r < 180*cYSize; r++ {
			worldLayers[l].Rows[r].Cells = make([]WorldCell, 180*cXSize)
		}
	}

	for li := 0; li < 21; li++ {
		for xi := 0; xi < cXSize; xi++ {
			for yi := 0; yi < cYSize; yi++ {
				for ri := 0; ri < 180; ri++ {
					for ci := 0; ci < 180; ci++ {
						worldLayers[li].Rows[yi*180+ri].Cells[xi*180+ci] = cells[(xi+yi*cXSize)*680400+li*32400+ri*180+ci]
					}
				}
			}
		}
	}

	// for li := 0; li < 21; li++ {
	// 	for xi := 0; xi < cXSize; xi++ {
	// 		for yi := 0; yi < cYSize; yi++ {
	// 			for ri := 0; ri < 180; ri++ {
	// 				worldCells := make([]WorldCell, 180*cXSize)
	// 				for ci := 0; ci < 180; ci++ {
	// 					worldCells[xi*180+ci] = cells[(xi+yi*cXSize)*680400+li*32400+ri*180+ci]
	// 				}
	// 				worldRows[yi*180+ri] = WorldRow{Cells: worldCells}
	// 			}
	// 		}
	// 	}
	// 	worldLayers[li] = WorldLayer{Rows: worldRows}
	// }
	world := World{Layers: worldLayers}

	index := (0+0*cXSize)*680400 + 0*32400 + 0*180 + 0
	fmt.Printf("checking out index: %v,  %v \n", index, cells[index])
	fmt.Printf("wut: %v", world.Layers[0].Rows[0].Cells[0])
	//fmt.Printf("first cell: %v", w.Layers[0].Rows[0].Cells[0])

	// worldLayers := make([]WorldLayer, 21)
	// for li := 0; li < 21; li++ {
	// 	rows := make([]string, 180*cYSize)
	// 	worldRows := make([]WorldRow, 180*cYSize)
	// 	for xi := 0; xi < cXSize; xi++ {
	// 		for yi := 0; yi < cYSize; yi++ {
	// 			ci := xi + yi*cXSize
	// 			for i := 0; i < 180; i++ {
	// 				start := ci*680400 + li*32400 + i*180
	// 				end := ci*680400 + li*32400 + i*180 + 180
	// 				chunkrow := strings.Join(textMap[start:end], "")
	// 				rows[yi*180+i] = rows[yi*180+i] + chunkrow
	// 			}
	// 		}
	// 	}

	// }

	return &world, nil
}

func (w *World) RenderToFiles(root string) error {
	err := os.MkdirAll(root, os.ModePerm)
	if err != nil {
		return err
	}
	for i, l := range w.Layers {
		filename := filepath.Join(root, fmt.Sprintf("o_%v", i))
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		var b strings.Builder
		for _, r := range l.Rows {
			for _, c := range r.Cells {
				b.WriteString(c.Symbol)
			}
			b.WriteString("\n")
		}
		f.WriteString(b.String())
	}
	return nil
}

func (o *Overmap) RenderToFilesAlt(m *metadata.Overmap, root string) error {
	err := os.MkdirAll(root, os.ModePerm)
	if err != nil {
		return err
	}

	missingTerrain := make(map[string]int)
	for _, c := range o.Chunks {
		for _, l := range c.Layers {
			for _, e := range l {
				if exists := m.Exists(e.OvermapTerrainID); !exists {
					if _, ok := missingTerrain[e.OvermapTerrainID]; !ok {
						missingTerrain[e.OvermapTerrainID] = 0
					} else {
						missingTerrain[e.OvermapTerrainID] = missingTerrain[e.OvermapTerrainID] + 1
					}
				}
			}
		}
	}

	if len(missingTerrain) > 0 {
		for k, v := range missingTerrain {
			fmt.Printf("missing terrain: %v x %v\n", k, v)
		}
	}

	cXMax := 0
	cXMin := 0
	cYMax := 0
	cYMin := 0

	for _, c := range o.Chunks {
		if c.X > cXMax {
			cXMax = c.X
		}
		if c.Y > cYMax {
			cYMax = c.Y
		}
		if c.X < cXMin {
			cXMin = c.X
		}
		if c.Y < cYMin {
			cYMin = c.Y
		}
	}

	cXSize := abs(cXMax) + abs(cXMin) + 1
	cYSize := abs(cYMax) + abs(cYMin) + 1

	chunkCapacity := cXSize * cYSize

	doneChunks := make(map[int]bool)
	textMap := make([]string, 680400*chunkCapacity)
	for _, c := range o.Chunks {
		ci := c.X + (0 - cXMin) + cXSize*(c.Y+0-cYMin)
		doneChunks[ci] = true
		fmt.Printf("processing x:%v, y:%v as %v\n", c.X, c.Y, ci)
		for li, l := range c.Layers {
			lzp := 0
			for _, e := range l {

				var s string
				if m.Exists(e.OvermapTerrainID) {
					s = m.Symbol(e.OvermapTerrainID)
				} else {
					s = "?"
				}

				for i := 0; i < int(e.Count); i++ {
					tmi := ci*680400 + li*32400 + lzp
					textMap[tmi] = s
					lzp++
				}
			}

			// rows := [180]string{}
			// for i := 0; i < 180; i++ {
			// 	start := ci*680400 + li*32400 + i*180
			// 	end := ci*680400 + li*32400 + i*180 + 180
			// 	rows[i] = strings.Join(textMap[start:end], "")
			// }

			// layer := strings.Join(rows[:], "\n")

			// filename := filepath.Join(root, fmt.Sprintf("o.%v.%v.%v", c.X, c.Y, li))
			// f, err := os.Create(filename)
			// if err != nil {
			// 	return err
			// }
			// defer f.Close()
			// f.WriteString(layer)
		}
	}

	for i := 0; i < chunkCapacity; i++ {
		if _, ok := doneChunks[i]; !ok {
			for e := 0; e < 680400; e++ {
				textMap[i*680400+e] = " "
			}
		}
	}

	for li := 0; li < 21; li++ {
		rows := make([]string, 180*cYSize)
		for xi := 0; xi < cXSize; xi++ {
			for yi := 0; yi < cYSize; yi++ {
				ci := xi + yi*cXSize
				for i := 0; i < 180; i++ {
					start := ci*680400 + li*32400 + i*180
					end := ci*680400 + li*32400 + i*180 + 180
					chunkrow := strings.Join(textMap[start:end], "")
					rows[yi*180+i] = rows[yi*180+i] + chunkrow
				}
			}
		}
		layer := strings.Join(rows[:], "\n")
		filename := filepath.Join(root, fmt.Sprintf("o_%v", li))
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		f.WriteString(layer)
	}

	fmt.Printf("dimensions x:%v, y:%v\n", cXSize, cYSize)

	return nil
}
