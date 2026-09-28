package imageUtils

import (
	"image"
	"image/color"
	"image/draw"
)

// Функция для выделения области на исходном изображении
func HighlightRect(img *image.RGBA, rect image.Rectangle, color color.NRGBA) {
	uniformColor := image.NewUniform(color)
	draw.Draw(img, rect, uniformColor, image.Point{}, draw.Over)
}
