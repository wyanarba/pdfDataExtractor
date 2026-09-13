package main

import (
	"image"
	"image/color"
	"math"
	"pdfDataExtractor/fileutils"
	"regexp"

	//"ansi"

	"github.com/ledongthuc/pdf"
)

type pdfTextBox struct {
	Text     string
	X, Y     float64
	FontSize float64
}

type imgTextBox struct {
	Text     string
	Coords   image.Point
	FontSize int
}

func ExtractText(pageInfo *PageData, pdfFilePath string, pdfPage int, tempImgPath string) {
	// Чтение файла pdf
	f, r, err := pdf.Open(pdfFilePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Открытие нужной страницы
	page := r.Page(pdfPage + 1)
	if page.V.IsNull() {
		panic("page not found")
	}

	// Переменные
	var (
		// Фразы
		tb           = pdfTextBox{}   // Текущая фраза
		imgTextBoxes = []imgTextBox{} // Фразы в изображении

		// PointsToPixels
		ratio     float64 = float64(DPI) / 72.0           // Коэффициент ppi / const(72)
		imgHeight         = pageInfo.boundsBeforeCut.Dy() // Высота таблицы (нужна для PointsToPixels)

		// Смещение после обрезки
		xOffset = pageInfo.Crop.CutOut.Min.X // по x
		yOffset = pageInfo.Crop.CutOut.Min.Y // по y

		// Регулярное выражение для поиска даты
		re = regexp.MustCompile(`на\s\d{2}\.\d{2}\.\d{4}\sг\.\s\([а-яА-ЯёЁ]+\)`)
	)

	// Само извлечение текста
	for _, literal := range page.Content().Text {

		// Завершающий фразу символ
		if literal.S[0] == 239 || literal.S[0] == '\n' {

			// PointsToPixels
			imgTb := imgTextBox{
				Text: tb.Text, // Текст

				Coords: image.Pt( // Координаты
					int(math.Round(tb.X*ratio))-xOffset,           // X
					imgHeight-int(math.Round(tb.Y*ratio))-yOffset, // Y
				),

				FontSize: int(math.Round(tb.FontSize * ratio)), // Размер шрифта
			}

			// Смещение по Y и X (Примерный центр глифа по коробке)
			imgTb.Coords.Y -= imgTb.FontSize / 4
			imgTb.Coords.X += imgTb.FontSize / 4

			// Защита от текста вне cropped изображения
			if imgTb.Coords.In(pageInfo.boundsAfterCut) == false {
				continue
			}

			// Определения ячейки
			isFound := false
			for i := range pageInfo.CountCols {
				for j := range pageInfo.CountRows {
					cell := pageInfo.Cells[i][j]

					// Найдена
					if imgTb.Coords.In(cell.CompareRect) {

						// add сохраняет информацию о текст боксе в ячейку
						add := func(cell *CellData) {
							cell.AllText += imgTb.Text + "\n"
						}

						// Обработка особых видов ячеек
						switch cell.CellType {
						// Часть от объединенной ячейки
						case CellType_PartOfMerged:
							add(cell.mergedCell)

						// Разделённая ячейка на 2 подгруппы
						case CellType_Divided:
							if imgTb.Coords.In(cell.dividedParts[0].CompareRect) {
								add(cell.dividedParts[0])
							} else {
								add(cell.dividedParts[1])
							}
						}

						add(cell)
						isFound = true
						break
					}
				}

				if isFound {
					break
				}
			}

			// Проверка, не дата ли эта строка
			dateStr := re.FindString(imgTb.Text)
			if dateStr != "" {
				// Дата без "на "
				pageInfo.Date = string(
					[]rune(dateStr)[3:],
				)
				pageInfo.DateFontSize = tb.FontSize
			}

			// Сохранение
			imgTextBoxes = append(imgTextBoxes, imgTb)
			tb = pdfTextBox{}

			continue
		}

		// Первая буква в фразе
		if tb.Text == "" {
			tb.X = literal.X
			tb.Y = literal.Y
		}

		// Изменение Y
		if literal.Y < tb.Y {
			tb.Y = literal.Y
		}

		// Размер шрифта
		if literal.FontSize > tb.FontSize {
			tb.FontSize = literal.FontSize
		}

		// Добавление буквы
		tb.Text += literal.S
	}

	// Отрисовка на изображении
	{
		var img *image.RGBA = fileutils.ReadImg(tempImgPath)

		// Сама отрисовка
		for i := range imgTextBoxes {
			imgTb := &imgTextBoxes[i]

			img.Set(imgTb.Coords.X, imgTb.Coords.Y, color.NRGBA{255, 0, 0, 255})
		}

		// Кодируем изображение в PNG
		fileutils.SaveImg(img, tempImgPath)
	}

}

func convertPointsToPixels() {

}
