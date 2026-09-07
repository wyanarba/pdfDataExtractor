package imageUtils

import (
	"image"
	"image/color"
	"image/draw"
)

func MergeImagesHorizontally(images []image.Image) image.Image {
	var (
		// Размерность выходного изображения
		width  = 0 // Сумма всех
		height = 0 // Максимальное из всех

		// Итоговое изображение
		newImg *image.RGBA

		// Ширина уже отрисованных изображений
		currentWidth = 0
	)

	// Расчёт размера итогового изображения
	for _, img := range images {
		width += img.Bounds().Dx()
		height = max(height, img.Bounds().Dy())
	}

	// Создание итогового изображения (подложки)
	newImg = image.NewRGBA(image.Rect(0, 0, width, height))
	white := color.RGBA{255, 255, 255, 255}
	draw.Draw(newImg, newImg.Bounds(), &image.Uniform{white}, image.Point{}, draw.Src)

	for _, img := range images {
		draw.Draw(
			newImg,
			image.Rect(
				currentWidth,
				0,
				img.Bounds().Dx()+currentWidth,
				img.Bounds().Dy(),
			),
			img,
			image.Point{},
			draw.Src,
		)

		currentWidth += img.Bounds().Dx()
	}

	return newImg
}

func MergeImagesVertically(images []image.Image) image.Image {
	var (
		// Размерность выходного изображения
		width  = 0 // Максимальное из всех
		height = 0 // Сумма всех

		// Итоговое изображение
		newImg *image.RGBA

		// Высота уже отрисованных изображений
		currentHeight = 0
	)

	// Расчёт размера итогового изображения
	for _, img := range images {
		width = max(width, img.Bounds().Dx())
		height += img.Bounds().Dy()
	}

	// Создание итогового изображения (подложки)
	newImg = image.NewRGBA(image.Rect(0, 0, width, height))
	white := color.RGBA{255, 255, 255, 255}
	draw.Draw(newImg, newImg.Bounds(), &image.Uniform{white}, image.Point{}, draw.Src)

	for _, img := range images {
		draw.Draw(
			newImg,
			image.Rect(
				0,
				currentHeight,
				img.Bounds().Dx(),
				img.Bounds().Dy()+currentHeight,
			),
			img,
			image.Point{},
			draw.Src,
		)

		currentHeight += img.Bounds().Dy()
	}

	return newImg
}

func MergeImagesVerticallyCentred(images []image.Image) image.Image {
	var (
		// Размерность выходного изображения
		width  = 0 // Максимальное из всех
		height = 0 // Сумма всех

		// Итоговое изображение
		newImg *image.RGBA

		// Высота уже отрисованных изображений
		currentHeight = 0
	)

	// Расчёт размера итогового изображения
	for _, img := range images {
		width = max(width, img.Bounds().Dx())
		height += img.Bounds().Dy()
	}

	// Создание итогового изображения (подложки)
	newImg = image.NewRGBA(image.Rect(0, 0, width, height))
	white := color.RGBA{255, 255, 255, 255}
	draw.Draw(newImg, newImg.Bounds(), &image.Uniform{white}, image.Point{}, draw.Src)

	for _, img := range images {
		// Смещение для оцентровки блока
		indent := (width - img.Bounds().Dx()) / 2

		draw.Draw(
			newImg,
			image.Rect(
				indent,
				currentHeight,
				indent+img.Bounds().Dx(),
				img.Bounds().Dy()+currentHeight,
			),
			img,
			image.Point{},
			draw.Src,
		)

		currentHeight += img.Bounds().Dy()
	}

	return newImg
}
