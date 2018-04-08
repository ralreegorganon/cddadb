package rasterize

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/freetype/truetype"

	"github.com/golang/freetype"
	"golang.org/x/image/font"
)

var (
	dpi      = flag.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontfile = flag.String("fontfile", "/Users/jj/Downloads/Topaz-8/Topaz-8.ttf", "filename of the ttf font")
	hinting  = flag.String("hinting", "none", "none | full")
	size     = flag.Float64("size", 24, "font size in points")
	spacing  = flag.Float64("spacing", 1, "line spacing (e.g. 2 means double spaced)")
	wonb     = flag.Bool("whiteonblack", true, "white text on a black background")
)

func Blam(path string) error {
	fontBytes, err := ioutil.ReadFile(*fontfile)
	if err != nil {
		return err
	}

	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return err
	}

	rawText, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	text := strings.Split(string(rawText), "\n")

	face := truetype.NewFace(f, &truetype.Options{
		Size:    *size,
		DPI:     *dpi,
		Hinting: font.HintingNone,
	})
	a := font.MeasureString(face, text[0])
	fmt.Println(a)

	fg, bg := image.Black, image.White
	if *wonb {
		fg, bg = image.White, image.Black
	}

	// blue := color.RGBA{0, 0, 255, 255}
	// bz := image.NewUniform(blue)

	rgba := image.NewRGBA(image.Rect(0, 0, 23056, 21600))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
	c := freetype.NewContext()
	c.SetDPI(*dpi)
	c.SetFont(f)
	c.SetFontSize(*size)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)
	switch *hinting {
	default:
		c.SetHinting(font.HintingNone)
	case "full":
		c.SetHinting(font.HintingFull)
	}

	fmt.Println(c.PointToFixed(*size * *spacing))

	// Draw the text.
	// pt := freetype.Pt(0, 0+int(c.PointToFixed(*size)>>6))
	// for _, s := range text {
	// 	_, err = c.DrawString(s, pt)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	pt.Y += c.PointToFixed(*size * *spacing)
	// }

	pt := freetype.Pt(0, 0+int(c.PointToFixed(*size)>>6))
	for _, s := range text {
		for _, cc := range s {
			awidth, ok := face.GlyphAdvance(cc)
			if ok != true {
				return fmt.Errorf("fuck")
			}
			//fmt.Printf("%v\n", pt.X)
			c.DrawString(string(cc), pt)
			pt.X += awidth
		}
		pt.X = c.PointToFixed(0)
		pt.Y += c.PointToFixed(*size * *spacing)
	}

	// Save that RGBA image to disk.
	outFile, err := os.Create("/Users/jj/Desktop/Grantsburg/rastertestb.png")
	if err != nil {
		return err
	}
	defer outFile.Close()
	b := bufio.NewWriter(outFile)
	err = png.Encode(b, rgba)
	if err != nil {
		return err
	}
	err = b.Flush()
	if err != nil {
		return err
	}

	return nil
}
