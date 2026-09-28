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

func RenderTextWithoutBackground(text string, fontBytes []byte, fontSize, dpi float64, padding int, textColor color.Color) image.Image {
	// Что то там клодик
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		log.Fatal(err)
	}

	face := truetype.NewFace(f, &truetype.Options{
		Size:    fontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})

	// Измеряем текст (с запасом, т.к. Ascent/Descent часто больше реальной высоты глифов)
	d := &font.Drawer{Face: face}
	textWidth := d.MeasureString(text).Ceil()
	metrics := face.Metrics()
	lineHeight := (metrics.Ascent + metrics.Descent).Ceil()

	extra := lineHeight / 2
	width := textWidth + extra*2
	height := lineHeight + extra*2

	tmp := image.NewRGBA(image.Rect(0, 0, width, height))

	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(f)
	c.SetFontSize(fontSize)
	c.SetClip(tmp.Bounds())
	c.SetDst(tmp)
	c.SetSrc(image.NewUniform(textColor))
	c.SetHinting(font.HintingFull)

	pt := freetype.Pt(extra, extra+metrics.Ascent.Ceil())
	if _, err := c.DrawString(text, pt); err != nil {
		log.Fatal(err)
	}

	// Обрезаем по реальным непрозрачным пикселям
	inkBounds := opaqueBounds(tmp)
	if inkBounds.Empty() {
		return image.NewRGBA(image.Rect(0, 0, padding*2, padding*2))
	}

	outW := inkBounds.Dx() + padding*2
	outH := inkBounds.Dy() + padding*2
	out := image.NewRGBA(image.Rect(0, 0, outW, outH))

	draw.Draw(out, image.Rect(padding, padding, padding+inkBounds.Dx(), padding+inkBounds.Dy()),
		tmp, inkBounds.Min, draw.Src)

	return out
}

func opaqueBounds(img *image.RGBA) image.Rectangle {
	// Что то там клодик
	b := img.Bounds()
	minX, minY := b.Max.X, b.Max.Y
	maxX, maxY := b.Min.X, b.Min.Y
	found := false

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				found = true
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	if !found {
		return image.Rectangle{}
	}
	return image.Rect(minX, minY, maxX+1, maxY+1)
}
