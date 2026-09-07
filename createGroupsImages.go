package main

import (
	"image"
	"path"
	"pdfDataExtractor/fileutils"
	im "pdfDataExtractor/imageUtils"
	"sync"
)

func CreateGroupsImages(pageInfo *PageData, saveFolderPath string) {
	var img *image.RGBA = pageInfo.tempData.img

	var (
		// Карта групп
		groups = map[string]*CellData{}
	)

	// Перебор всех ячеек
	{
		var cell *CellData

		for i := range pageInfo.CountCols {
			for j := range pageInfo.CountRows {
				cell = pageInfo.Cells[i][j]

				if cell.CellType == CellType_Group && cell.AllText != "" {
					groups[cell.AllText[:len(cell.AllText)-1]] = cell
				}
			}
		}
	}

	// Ширина первой внешней рамки таблицы
	var fBorder = pageInfo.Cells[0][0].Border[0]

	// Функция для создания самих изображений
	var makeImages = func(groupName string) {
		var (
			// Верхняя и нижняя ячейка в расписании группы
			upperCell *CellData = groups[groupName]
			lowerCell *CellData = getLastCellBeforeUnderGroupCell(
				pageInfo,
				pageInfo.Cells[upperCell.I][upperCell.J+1],
			)

			// Копия верхней ячейки, у которой изменены размеры Rect (размером со всё расписание группы)
			mutantCell CellData = *upperCell
		)
		mutantCell.Rect.Max.Y = lowerCell.Rect.Max.Y

		// Создание маленьких изображений с расписанием группы
		{
			// Линия с самой мини таблицей расписания
			shLine := GetLineCellAndTime(
				pageInfo,
				img,
				&mutantCell,
				IndentInfo{
					0,
					0,
					0,
					0,
				},
			)

			resultImg := im.MergeImagesVerticallyCentred(
				[]image.Image{
					pageInfo.tempData.buildNumLine,
					pageInfo.tempData.dateLine,
					pageInfo.tempData.spacerLine,
					shLine,
				},
			)

			resultImg = im.AddBorder(resultImg, int(float32(fBorder)*0.6))

			// Сохранение изображения
			fileutils.SaveImg(
				resultImg,
				path.Join(
					saveFolderPath,
					groupName+"S.png",
				),
			)
		}

		// Создание больших картинок с расписанием группы
		{
			// Копирование исходника
			resultImg := image.NewRGBA(img.Bounds())
			copy(resultImg.Pix, img.Pix)

			im.HighlightRect(
				resultImg,
				mutantCell.Rect,
				ColorHighlightMain,
			)

			// Сохранение изображения
			fileutils.SaveImg(
				resultImg,
				path.Join(
					saveFolderPath,
					groupName+".png",
				),
			)
		}
	}

	//
	wg := sync.WaitGroup{}
	for groupName := range groups {
		wg.Add(1)
		go func() {
			defer wg.Done()
			makeImages(groupName)
		}()
	}
	wg.Wait()

	// // Параллельная обработка
	// wg := sync.WaitGroup{}
	// for teacherName := range teachers {
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		makeImages(teacherName)
	// 	}()
	// }
	// wg.Wait()
}
