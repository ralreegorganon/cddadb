package main

import (
	"flag"

	"github.com/ralreegorganon/cddadb/cmd/cddadb-map/metadata"
	"github.com/ralreegorganon/cddadb/cmd/cddadb-map/overmap"
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

	// asdf
	/*
		fontForeGroundColor, fontBackGroundColor := image.NewUniform(color.RGBA{255, 255, 255, 255}), image.NewUniform(color.RGBA{0, 0, 0, 255})

		background := image.NewRGBA(image.Rect(0, 0, 4096, 4096))

		draw.Draw(background, background.Bounds(), fontBackGroundColor, image.ZP, draw.Src)

		fontBytes, _ := ioutil.ReadFile("/Users/jj/Library/Fonts/SourceCodePro-Regular.otf")
		font, err := truetype.Parse(fontBytes)
		if err != nil {
			log.Fatal(err)
		}

		fontsize := 24.0
		ctx := freetype.NewContext()
		ctx.SetDPI(72)
		ctx.SetFont(font)
		ctx.SetFontSize(fontsize)
		ctx.SetClip(background.Bounds())
		ctx.SetDst(background)
		ctx.SetSrc(fontForeGroundColor)

		var UTF8text = rows[:]

		pt := freetype.Pt(10, 10+int(ctx.PointToFixed(fontsize)>>6))

		for _, str := range UTF8text {
			_, err := ctx.DrawString(str, pt)
			if err != nil {
				fmt.Println(err)
				return
			}
			pt.Y += ctx.PointToFixed(fontsize)
		}

		// Save
		outFile, err := os.Create("utf8text.png")
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		defer outFile.Close()
		buff := bufio.NewWriter(outFile)

		err = png.Encode(buff, background)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		// flush everything out to file
		err = buff.Flush()
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	*/

}
