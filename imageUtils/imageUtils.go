package imageUtils

import (
	"image"
	"image/color"
	"image/draw"
)

// Функция для вырезки прямоугольника из изображения
func CropRect(img image.Image, rect image.Rectangle) image.Image {
	cropped := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	draw.Draw(cropped, cropped.Bounds(), img, rect.Min, draw.Src)
	return cropped
}

func AddBorder(img image.Image, size int) image.Image {
	var (
		// Размерность выходного изображения
		width  = img.Bounds().Dx() + size*2
		height = img.Bounds().Dy() + size*2
	)

	// Создание итогового изображения (подложки)
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))
	white := color.RGBA{255, 255, 255, 255}
	draw.Draw(newImg, newImg.Bounds(), &image.Uniform{white}, image.Point{}, draw.Src)

	// Отрисовка самого изображения
	draw.Draw(
		newImg,
		image.Rect(
			size,
			size,
			size+img.Bounds().Dx(),
			size+img.Bounds().Dy(),
		),
		img,
		image.Point{},
		draw.Src,
	)

	return newImg
}
