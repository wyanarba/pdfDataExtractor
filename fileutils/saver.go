package fileutils

import (
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
)

func ReadImg(filePath string) (img *image.RGBA) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Не удалось открыть файл c картинкой (CropImage): %v", err)
	}
	defer file.Close()
	imgSrc, _, err := image.Decode(file)
	if err != nil {
		log.Fatalf("Не удалось декодировать: %v", err)
	}
	// Копирование изображения и конвертация
	bounds := imgSrc.Bounds()
	img = image.NewRGBA(bounds)
	draw.Draw(img, bounds, imgSrc, bounds.Min, draw.Src)

	return
}

func SaveImg(img image.Image, filePath string) {
	outFile, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Не удалось создать итоговый файл %s: %v", filePath, err)
	}
	defer outFile.Close()
	err = png.Encode(outFile, img)
	if err != nil {
		log.Fatalf("Ошибка при сохранении: %v", err)
	}
}
