package imageUtils

import (
	"image"
	"image/color"
	"image/draw"
	"log"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

// RenderTextRectangle рисует прямоугольник с текстом и белым фоном, по заданным параметрам
func RenderTextRectangle(text string, fontBytes []byte, fontSize, dpi float64, padding int) image.Image {
	// Немного изменённая функция от клода, не хочу разбираться в этих отступах у глифов шрифтов

	f, err := truetype.Parse(fontBytes)
	if err != nil {
		log.Fatal(err)
	}

	face := truetype.NewFace(f, &truetype.Options{
		Size:    fontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})

	// Измеряем текст
	d := &font.Drawer{Face: face}
	textWidth := d.MeasureString(text).Ceil()
	metrics := face.Metrics()
	textHeight := (metrics.Ascent + metrics.Descent).Ceil()

	width := textWidth + padding*2
	height := textHeight + padding*2

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(f)
	c.SetFontSize(fontSize)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(color.Black))
	c.SetHinting(font.HintingFull)

	pt := freetype.Pt(padding, padding+metrics.Ascent.Ceil())
	if _, err := c.DrawString(text, pt); err != nil {
		log.Fatal(err)
	}

	return img
}
