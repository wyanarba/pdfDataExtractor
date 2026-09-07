package main

import (
	"image"
	"image/draw"
	"pdfDataExtractor/fileutils"
)

func CropImage(pageInfo *PageData, savePath string) {
	var img *image.RGBA = pageInfo.tempData.img

	cropInfo := CropInfo{} // Результаты работы функции
	bounds := img.Bounds()
	pageInfo.boundsBeforeCut = bounds

	BorderWith := 0 // Ширина внешней линии таблицы (обрамления)
	// Получение сколько удалять слева и ширины внешней линии таблицы
	{
		type lineInfo struct {
			countFound int         // Сколько раз было найдено (Суммарно)
			lengths    map[int]int // Длинна и сколько раз была найдено
		}
		lineLengths := map[int]lineInfo{} // Информация о первой чёрной линии в строке (X начала и информация)

		// Обработка изображения
		for y := 0; y < bounds.Max.Y; y++ {
			var (
				lineStart, lineEnd     int  // Начало и конец линии, для толщины линии контура
				lineStarted, lineEnded bool // Найдено ли начало
			)

			// Поиск линии
			for x := 0; x < bounds.Max.X; x++ {

				// Проверка на чёрный
				isDark := isDarkColor(
					normalizeColor(
						img.At(x, y).RGBA(),
					),
				)

				// Начало и конец линии
				if lineStarted == false {
					if isDark {
						lineStarted = true
						lineStart = x
					}
				} else {
					if isDark == false {
						lineEnd = x - 1
						lineEnded = true
						break
					}
				}
			}

			// Сохранение информации о линии
			if lineStarted && lineEnded {
				len := lineEnd - lineStart

				info, ok := lineLengths[lineStart]

				// Значение встречено впервые
				if ok == false {
					info.lengths = map[int]int{}
				}

				info.countFound++
				info.lengths[len]++

				lineLengths[lineStart] = info
			}
		}

		// Поиск самого частого x
		Key, mostFrequency := 0, 0
		for key := range lineLengths {
			if mostFrequency < lineLengths[key].countFound {
				mostFrequency = lineLengths[key].countFound
				Key = key
			}
		}
		cropInfo.CutOut.Min.X = Key // Сохранение x, как начала для обрезки

		// Поиск самой частой длинны
		lengths := lineLengths[Key].lengths
		mostFrequency = 0
		for key := range lengths {
			if mostFrequency < lengths[key] {
				mostFrequency = lengths[key]
				Key = key
			}
		}
		BorderWith = Key
	}

	// Получение сколько удалять сверху
	{
		minYTop := bounds.Max.Y

		for x := 0; x < bounds.Max.X; x++ {
			for y := 0; y < bounds.Max.Y; y++ {
				// Проверка на чёрный
				isDark := isDarkColor(
					normalizeColor(
						img.At(x, y).RGBA(),
					),
				)

				if isDark {
					if y < minYTop {
						minYTop = y
					}
					break
				}
			}
		}

		// Сохранение результатов
		cropInfo.CutOut.Min.Y = minYTop
	}

	// Получение сколько удалять слева
	{
		maxXLeft := 0

		for y := bounds.Max.Y - 1; y >= 0; y-- {
			for x := bounds.Max.X - 1; x >= 0; x-- {
				// Проверка на чёрный
				isDark := isDarkColor(
					normalizeColor(
						img.At(x, y).RGBA(),
					),
				)

				if isDark {
					if x > maxXLeft {
						maxXLeft = x
					}
					break
				}
			}
		}

		// Сохранение результатов
		cropInfo.CutOut.Max.X = maxXLeft
	}

	// Получение сколько удалять снизу
	{
		maxYBottom := 0

		for x := bounds.Max.X - 1; x >= 0; x-- {
			for y := bounds.Max.Y - 1; y >= 0; y-- {
				// Проверка на чёрный
				isDark := isDarkColor(
					normalizeColor(
						img.At(x, y).RGBA(),
					),
				)

				if isDark {
					if y > maxYBottom {
						maxYBottom = y
					}
					break
				}
			}
		}

		// Сохранение результатов
		cropInfo.CutOut.Max.Y = maxYBottom
	}

	// Отступы от краёв таблицы, что бы было красивее
	const (
		topMarginRatio    = 2.0
		commonMarginRatio = 1.7
	)
	cropInfo.Margin = IndentInfo{
		int(float32(BorderWith) * commonMarginRatio),
		int(float32(BorderWith) * topMarginRatio),
		int(float32(BorderWith) * commonMarginRatio),
		int(float32(BorderWith) * commonMarginRatio),
	}

	// Зона обрезки таблицы
	cropInfo.CutOut = image.Rect(
		cropInfo.CutOut.Min.X-cropInfo.Margin[0],
		cropInfo.CutOut.Min.Y-cropInfo.Margin[1],
		cropInfo.CutOut.Max.X+cropInfo.Margin[2],
		cropInfo.CutOut.Max.Y+cropInfo.Margin[3],
	)

	// Кроп
	croppedImg := image.NewRGBA(
		image.Rect(0, 0, cropInfo.CutOut.Dx(), cropInfo.CutOut.Dy()),
	)
	draw.Draw(croppedImg, croppedImg.Bounds(), img, cropInfo.CutOut.Min, draw.Src)
	pageInfo.boundsAfterCut = croppedImg.Bounds()

	// Сохранение
	fileutils.SaveImg(croppedImg, savePath)
	pageInfo.tempData.img = croppedImg

	pageInfo.Crop = cropInfo
}
