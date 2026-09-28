package pdfDataExtractor

import (
	"image"
	"image/draw"
	"slices"

	"github.com/wyanarba/pdfDataExtractor/fileutils"
	im "github.com/wyanarba/pdfDataExtractor/imageUtils"
)

// GetCellsFromImage извлекает ячейки из переданного обрезанного изображения таблицы расписания.
// Так-же определяет их тип.
func GetCellsFromImage(pageInfo *PageData, savePath string) {
	// Копирование изображения
	var img *image.RGBA = image.NewRGBA(pageInfo.tempData.img.Rect)
	copy(img.Pix, pageInfo.tempData.img.Pix)

	bounds := img.Bounds()

	var (
		lenRay      = pageInfo.Crop.Margin[1] * 5 // Длина луча проверки
		xLeftCorner = pageInfo.Crop.Margin[1]     // Координата X начала таблицы
		//yLeftCorner = fileInfo.Crop.Margin[1]    // Координата Y начала таблицы
		tempBuffer = []int{} // Временный буфер, для результатов сканирования (до группировки)

		// Минимальные и максимальные точки в прямоугольнике (область пересечения 2х линий)
		maxPoints = [2][]int{} // X1, Y1
		minPoints = [2][]int{} // X2, Y2

		X, Y = 0, 1
	)

	// Получение пересечений самой левой вертикальной линии
	// И пересечений по Y
	{
		// Сканирование
		for y := 0; y < bounds.Max.Y; y++ {
			if castRay(img, lenRay, y, xLeftCorner, East) {
				tempBuffer = append(tempBuffer, y)
			}
		}

		// Объединение точек
		min, max := chunkSequence(tempBuffer)
		minPoints[Y] = append(minPoints[Y], min...)
		maxPoints[Y] = append(maxPoints[Y], max...)
	}

	// Проверка на валидность самой верхней горизонтальной линии
	// (один раз надписи над таблицей были обёрнуты в единую ячейку таблицы, что вызывало ошибки)
	// Линия должна содержать не менее 5 пересечений (значит ячейка под ней не является объединенной)
	{
		tempBuffer = []int{}
		lenRay = minPoints[Y][1] - maxPoints[Y][0] // От второй горизонтальной линии до первой
		rayStartCoord := minPoints[Y][0]           // Высота (начальная точка по x) для начала луча

		// Сканирование
		for x := 0; x < bounds.Max.X; x++ {
			if castRay(img, lenRay, x, rayStartCoord, South) {
				tempBuffer = append(tempBuffer, x)
			}
		}

		// Объединение точек
		min, _ := chunkSequence(tempBuffer)

		if len(min) < 6 { // Самая верхняя линия не является частью обычной шапки таблицы
			// Удаляем информацию о ней
			minPoints[Y] = minPoints[Y][1:]
			maxPoints[Y] = maxPoints[Y][1:]
		}
	}

	// Получение координат пересечений по X
	{
		tempBuffer = []int{}
		lenRay = minPoints[Y][2] - maxPoints[Y][0] // От третьей горизонтальной линии до первой
		rayStartCoord := minPoints[Y][0]           // Высота (начальная точка по x) для начала луча

		// Сканирование
		for x := 0; x < bounds.Max.X; x++ {
			if castRay(img, lenRay, x, rayStartCoord, South) {
				tempBuffer = append(tempBuffer, x)
			}
		}

		// Объединение точек
		min, max := chunkSequence(tempBuffer)
		minPoints[X] = append(minPoints[X], min...)
		maxPoints[X] = append(maxPoints[X], max...)
	}

	// Создание ячеек таблицы
	{
		// Размерность таблицы
		pageInfo.CountCols = len(minPoints[X]) - 1
		pageInfo.CountRows = len(minPoints[Y]) - 1

		// Инициализация таблицы
		{
			// Создаём колонки
			pageInfo.Cells = make([][]*CellData, pageInfo.CountCols)

			// Создаём строки
			for i := range pageInfo.Cells {
				pageInfo.Cells[i] = make([]*CellData, pageInfo.CountRows)
			}
		}

		CellsHeight := []int{} // Высоты ячеек (для отделения CellType_Group)

		// Создание ячеек
		for i := 0; i < pageInfo.CountCols; i++ {
			for j := 0; j < pageInfo.CountRows; j++ {
				// Создание ячейки
				cell := MakeCellData(
					i, j,
					image.Rect(
						maxPoints[X][i]+1,
						maxPoints[Y][j]+1,
						minPoints[X][i+1],
						minPoints[Y][j+1],
					),
					IndentInfo{
						maxPoints[X][i] - minPoints[X][i] + 1,
						maxPoints[Y][j] - minPoints[Y][j] + 1,
						maxPoints[X][i+1] - minPoints[X][i+1] + 1,
						maxPoints[Y][j+1] - minPoints[Y][j+1] + 1,
					},
				)

				// Определение типа ячейки (служебная или обычная)
				if i < HeaderColumnsCount || j < HeaderRowsCount {
					cell.CellType = CellType_Service
				} else {
					cell.CellType = CellType_Basic
				}

				// Сохранение высоты ячейки
				if i == 0 && j >= HeaderRowsCount {
					CellsHeight = append(CellsHeight, cell.Rect.Dy())
				}

				// Сохранение ячейки
				pageInfo.Cells[i][j] = &cell
			}
		}

		// Отделение CellType_Group (разделение на 2 группы по высоте basic ячеек)
		{
			hc := make([]int, len(CellsHeight))
			copy(hc, CellsHeight)

			// 1. Сортируем
			slices.Sort(hc)
			splitThreshold := 0

			for range 10 {
				// 2. Ищем самый большой разрыв между соседними числами
				maxGap := 0
				splitIndex := 0

				for i := 1; i < len(hc); i++ {
					gap := hc[i] - hc[i-1]
					if gap > maxGap {
						maxGap = gap
						splitIndex = i // Запоминаем место разреза
					}
				}

				splitThreshold = hc[splitIndex] // Наименьшее значение из больших

				// Защита от случаев когда одна строка не естественно высокая и всё портит
				// (Group / Regular)
				if len(hc[:splitIndex])/len(hc[splitIndex:]) > 2 {
					hc = hc[:splitIndex]
				} else {
					break
				}
			}

			// Если размер меньшее splitThreshold, изменяем тип ячеек в строке на CellType_Group
			for idx, height := range CellsHeight {
				if height < splitThreshold {

					// Изменение строки в таблице
					j := idx + HeaderRowsCount
					for i := HeaderColumnsCount; i < pageInfo.CountCols; i++ {

						pageInfo.Cells[i][j].CellType = CellType_Group
					}
				}
			}
		}
	}

	// Поиск CellType_Divided и CellType_PartOfMerged
	{
		// Смещение для исключения влияния толстых линий
		offset := (maxPoints[X][0]-minPoints[X][0])/2 + 1

		for i := range pageInfo.CountCols {
			for j := range pageInfo.CountRows {
				cell := pageInfo.Cells[i][j]

				if cell.CellType != CellType_Basic {
					continue
				}

				// Проверка на разделение на 2 части
				{
					tempBuffer = []int{}
					scanStartCoord := cell.Rect.Min.X + offset
					scanEndCoord := cell.Rect.Max.X - offset
					rayStartCoord := cell.Rect.Min.Y + offset
					rayLen := (cell.Rect.Max.Y - offset) - rayStartCoord

					// Сканирование
					for x := scanStartCoord; x <= scanEndCoord; x++ {
						if castRay(img, rayLen, x, rayStartCoord, South) {
							tempBuffer = append(tempBuffer, x)
						}
					}

					min, max := chunkSequence(tempBuffer)

					if len(min) > 1 {
						Settings.Logger.Error("Мега ошибка", i, j, len(min))
					}

					if len(min) == 1 {
						cell.CellType = CellType_Divided

						// Создание ячеек
						leftCell := MakeCellData(
							cell.I,
							cell.J,

							image.Rect(
								cell.Rect.Min.X,
								cell.Rect.Min.Y,
								min[0],
								cell.Rect.Max.Y,
							),

							IndentInfo{
								cell.Border[0],
								cell.Border[1],
								max[0] - min[0],
								cell.Border[3],
							},
						)

						rightCell := MakeCellData(
							cell.I,
							cell.J,

							image.Rect(
								max[0]+1,
								cell.Rect.Min.Y,
								cell.Rect.Max.X,
								cell.Rect.Max.Y,
							),

							IndentInfo{
								max[0] - min[0] + 1,
								cell.Border[1],
								cell.Border[2],
								cell.Border[3],
							},
						)

						// Служебная информация
						leftCell.dividedParentCell = cell
						rightCell.dividedParentCell = cell
						leftCell.CellType, rightCell.CellType = CellType_PartOfDivided, CellType_PartOfDivided

						// Сохранение
						pageInfo.dividedPartCells = append(pageInfo.dividedPartCells, &leftCell, &rightCell)
						cell.DividedParts = [2]*CellData{&leftCell, &rightCell}
					}

				}

				if cell.CellType != CellType_Basic {
					continue
				}

				// Проверка на объединённую ячейку
				{
					tempBuffer = []int{}
					scanStartCoord := cell.Rect.Min.Y - cell.Border[1]
					scanEndCoord := cell.Rect.Max.Y + cell.Border[3]
					rayStartCoord := cell.Rect.Min.X + offset
					rayLen := (cell.Rect.Max.X - offset) - rayStartCoord

					// Сканирование
					for y := scanStartCoord; y <= scanEndCoord; y++ {
						if castRay(img, rayLen, y, rayStartCoord, East) {
							tempBuffer = append(tempBuffer, y)
						}
					}

					min, _ := chunkSequence(tempBuffer)
					if len(min) < 2 {
						cell.CellType = CellType_PartOfMerged

						mrg := pageInfo.Cells[i][j-1].MergedCell
						if mrg == nil {
							mrg = &CellData{}
							pageInfo.mergedCell = append(pageInfo.mergedCell, mrg)
							cell.MergedCell = mrg
							mrg.partsOfMergedCell = append(mrg.partsOfMergedCell, cell)

							mrg.Rect = cell.Rect
							mrg.Border = cell.Border
							mrg.CellType = CellType_Merged

							mrg.CompareRect = image.Rect(
								mrg.Rect.Min.X-mrg.Border[0],
								mrg.Rect.Min.Y-mrg.Border[1],
								mrg.Rect.Max.X,
								mrg.Rect.Max.Y,
							)

							mrg.I = cell.I
							mrg.J = cell.J
						} else {
							cell.MergedCell = mrg
							mrg.partsOfMergedCell = append(mrg.partsOfMergedCell, cell)

							mrg.Rect.Max.Y = cell.Rect.Max.Y
							mrg.Border[2] = cell.Border[2]

							mrg.CompareRect = image.Rect(
								mrg.Rect.Min.X-mrg.Border[0],
								mrg.Rect.Min.Y-mrg.Border[1],
								mrg.Rect.Max.X,
								mrg.Rect.Max.Y,
							)
						}

					}
				}
			}
		}
	}

	// Отрисовка таблицы
	{
		for i := range pageInfo.CountCols {
			for j := range pageInfo.CountRows {
				cell := pageInfo.Cells[i][j]

				// Базовый цвет
				colorFill := ColorNN1
				switch cell.CellType {

				// Сервисный цвет
				case CellType_Service:
					colorFill = ColorNN2

				// Цвет группы
				case CellType_Group:
					colorFill = ColorNN3

				// Цвет части разделённой ячейки
				case CellType_Divided:
					colorFill = ColorNN6

				case CellType_PartOfMerged:
					continue
				}

				if cell.CellType == CellType_Divided {
					im.HighlightRect(img, cell.DividedParts[0].Rect, ColorNN6)
					im.HighlightRect(img, cell.DividedParts[1].Rect, ColorNN6)
				} else {
					im.HighlightRect(img, cell.Rect, colorFill)
				}

			}
		}

		for _, mrg := range pageInfo.mergedCell {
			uniformColor := image.NewUniform(ColorNN7)
			draw.Draw(img, mrg.Rect, uniformColor, image.Point{}, draw.Over)
		}
	}

	// Кодируем изображение в PNG
	fileutils.SaveImg(img, savePath)
}

// Вспомогательные функции

type Direction rune

const (
	North Direction = 'N' // Север
	West  Direction = 'W' // Запад
	South Direction = 'S' // Юг
	East  Direction = 'E' // Восток
)

// castRay бросает луч длинны lenRay в направлении direction.
// Координаты начала луча задаются через rayOffsetCoord (изменяется в внешнем цикле, вне функции, например для направления 'E' это Y)
// и rayStartCoord (изменяется во внутреннем цикле, в функции, например для направления 'E' это Y).
// Функция принимает изображение img, и возвращает true, если все пиксели на пути луча были чёрными.
func castRay(img *image.RGBA, lenRay, rayOffsetCoord, rayStartCoord int, direction Direction) (isDone bool) {
	var (
		step       int
		finalPoint int
		i          = rayStartCoord
	)
	isDone = true

	if direction == 'N' {
		// Север (y вверх)
		step = -1
		finalPoint = rayStartCoord - lenRay

		for ; i > finalPoint; i += step {
			isDark := isDarkColor(
				normalizeColor(
					img.At(rayOffsetCoord, i).RGBA(),
				),
			)

			if isDark == false {
				isDone = false
				break
			}
		}

	} else if direction == 'W' {
		// Запад (x влево, на минус)))   )
		step = -1
		finalPoint = rayStartCoord - lenRay

		for ; i > finalPoint; i += step {
			isDark := isDarkColor(
				normalizeColor(
					img.At(i, rayOffsetCoord).RGBA(),
				),
			)

			if isDark == false {
				isDone = false
				break
			}
		}

	} else if direction == 'S' {
		// Юг (y вниз)
		step = 1
		finalPoint = rayStartCoord + lenRay

		for ; i < finalPoint; i += step {
			isDark := isDarkColor(
				normalizeColor(
					img.At(rayOffsetCoord, i).RGBA(),
				),
			)

			if isDark == false {
				isDone = false
				break
			}
		}

	} else if direction == 'E' {
		// Восток (x вправо, на плюс)))   )
		step = 1
		finalPoint = rayStartCoord + lenRay

		for ; i < finalPoint; i += step {
			isDark := isDarkColor(
				normalizeColor(
					img.At(i, rayOffsetCoord).RGBA(),
				),
			)

			if isDark == false {
				isDone = false
				break
			}
		}
	}

	return
}

// chunkSequence из переданной последовательности координат seq извлекает координаты minPoints и maxPoints (начала и конца) узлов (узел - это пересечение линий в таблице).
func chunkSequence(seq []int) (minPoints, maxPoints []int) {
	const (
		mergeDistance = 3
	)

	// Защита
	if len(seq) == 0 {
		return
	}

	slices.Sort(seq)
	seq = append(seq, 10e9) // Для того что бы последняя точка нормально обработалась
	lastEl := seq[0]
	minEl := seq[0]

	for _, el := range seq {
		if el-lastEl > mergeDistance {
			minPoints = append(minPoints, minEl)
			maxPoints = append(maxPoints, lastEl)

			minEl = el
		}

		lastEl = el
	}

	return
}
