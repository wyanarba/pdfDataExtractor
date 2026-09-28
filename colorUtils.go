package pdfDataExtractor

import (
	"image"
	"image/color"
	"image/draw"
)

func isDarkColor(r, b, g, a uint8) bool {
	return r < 200 && b < 200 && g < 200
}

func normalizeColor(r, g, b, a uint32) (r8, g8, b8, a8 uint8) {
	r8, b8, g8, a8 = uint8(r>>8), uint8(b>>8), uint8(g>>8), uint8(a>>8)
	return
}

func fillCellOnImage(cells []*CellData, img *image.RGBA, color *color.NRGBA) {
	// Дефолтный цвет выделения
	if color == nil {
		color = &ColorNN13
	}

	uniformColor := image.NewUniform(color)
	for _, cell := range cells {
		draw.Draw(img, cell.Rect, uniformColor, image.Point{}, draw.Over)
	}
}
