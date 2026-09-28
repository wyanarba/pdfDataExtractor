package pdfDataExtractor

import (
	_ "embed"
	"path"
	"sync"

	"image"

	"regexp"
	"sort"
	"strings"

	"github.com/wyanarba/pdfDataExtractor/fileutils"
	im "github.com/wyanarba/pdfDataExtractor/imageUtils"
)

// Шрифты для рендера текста
var (
	//go:embed times.ttf
	timesFont []byte

	//go:embed timesB.ttf
	timesBoldFont []byte
)

func CreateTeachersImages(pageInfo *PageData, saveFolderPath string) {
	// Загрузка изображения
	var (
		img   *image.RGBA = pageInfo.tempData.img
		imgCh             = pageInfo.tempData.imgCh
	)

	var (
		// Регулярное выражение для ФИО формата Иванов И.И. с любым количеством пробелов
		re = regexp.MustCompile(`[А-ЯЁ][а-яё]+\s*[А-ЯЁ]\.\s*[А-ЯЁ]\.`)

		// Карта преподавателей
		teachers = map[string][]*CellData{}
	)

	// processCell обрабатывает ячейку, сохраняя, найденных преподавателей в тексте ячейки, в карту
	processCell := func(cell *CellData) {
		foundStrs := re.FindAllString(cell.AllText, -1)

		if len(foundStrs) == 0 {
			return
		}

		for _, name := range foundStrs {
			// Форматирование
			name = strings.Replace(name, " ", "", -1)
			name = strings.Replace(name, ".", "", -1)

			// Сохранение
			cells, isOk := teachers[name]

			// Первая запись
			if isOk == false {
				cells = []*CellData{}
			}

			cells = append(cells, cell)
			teachers[name] = cells
		}
	}

	// Перебор всех ячеек
	{
		for i := range pageInfo.CountCols {
			for j := range pageInfo.CountRows {
				cell := pageInfo.Cells[i][j]

				switch cell.CellType {
				// Обработка ячейки разделенной на 2 подгруппы
				case CellType_Divided:
					processCell(cell.DividedParts[0])
					processCell(cell.DividedParts[1])

				// Для объединённых ячеек обработка будет вызвана отдельно
				case CellType_PartOfMerged:
					continue

				// Прочие ячейки
				default:
					processCell(cell)
				}
			}
		}

		for _, mc := range pageInfo.mergedCell {
			processCell(mc)
		}
	}

	// Сортировка слайсов в значениях карт, по расстоянию до ближайшей CellType_Group
	{
		for key := range teachers {
			slice := teachers[key]

			sort.Slice(slice, func(i, j int) bool {
				if slice[i].CellType == CellType_Merged {
					return false
				}

				return getDistanceToGroupCell(pageInfo, slice[i]) < getDistanceToGroupCell(pageInfo, slice[j])
			})
		}
	}

	// Ширина первой внешней рамки таблицы
	var fBorder = pageInfo.Cells[0][0].Border[0]

	// Изменённые преподаватели
	var (
		changedTeachers = make(chan string)
	)

	// Функция для создания самих изображений
	var makeImages = func(teacherName string) {
		// Создание маленьких изображений с расписанием преподавателей
		{
			var (
				// Изображения время + ячейка пары
				lines []image.Image

				// Информация о ячейке, для получения линий
				lastGroupCell *CellData // Ячейка последней группы
				lastOffset    = -101    // Смещение последней пары от названия группы
			)

			// Создание пар
			for index, currentCell := range teachers[teacherName] {
				// Учёт разделённой пары (по подгруппам)
				if currentCell.CellType == CellType_PartOfDivided {
					currentCell = currentCell.dividedParentCell
				}

				// Проверка, нужно ли отрисовывать ячейку заголовка группы
				needNewGroupLine := true
				{
					// Ячейка текущей группы
					currentGroupCell := getUpperGroupCell(pageInfo, currentCell)
					// Смещение текущей пары от названия группы
					currentOffset := getDistanceToGroupCell(pageInfo, currentCell)

					// Проверка
					if lastGroupCell == currentGroupCell && lastOffset+1 == currentOffset {
						needNewGroupLine = false
					}

					lastGroupCell = currentGroupCell
					lastOffset = currentOffset
				}

				// Нужно отрисовать ячейку заголовка группы
				if needNewGroupLine {
					// Определение типа отступов
					var margin = IndentInfo{}
					{
						// Для меньшей толщины общей линии
						margin[3] = -3

						if index != 0 {
							margin[1] = lastGroupCell.Border[1] / 2 * -1
						}
					}

					// Получение линии с заголовком группы
					line := GetLineCellAndTime(
						pageInfo,
						imgCh,
						lastGroupCell,
						margin,
					)

					if line != nil {
						// Сохранение линии
						lines = append(
							lines,
							im.MergeImagesHorizontally(
								[]image.Image{
									line,
								},
							),
						)
					}
				}

				// Определение типа отступов
				var margin = IndentInfo{}
				{
					// Для хорошего размера самой нижней линии в вырезке
					if index == len(teachers[teacherName])-1 {
						margin[3] = 1
					}

					if needNewGroupLine && index != 0 {
						margin[1] = -1
					}
				}

				// Получение линии
				line := GetLineCellAndTime(
					pageInfo,
					imgCh,
					currentCell,
					margin,
				)

				// Сохранение линии
				lines = append(
					lines,
					im.MergeImagesHorizontally(
						[]image.Image{
							line,
						},
					),
				)
			}

			// Промежуточный результат вырезки из расписания, без даты
			resultImg := im.MergeImagesVertically(lines)

			resultImg = im.MergeImagesVerticallyCentred(
				[]image.Image{
					pageInfo.tempData.buildNumLine,
					pageInfo.tempData.dateLine,
					pageInfo.tempData.spacerLine,
					resultImg,
				},
			)

			resultImg = im.AddBorder(resultImg, int(float32(fBorder)*0.6))

			// Сохранение изображения
			fileutils.SaveImg(
				resultImg,
				path.Join(
					saveFolderPath,
					teacherName+"S.png",
				),
			)
		}

		// Создание больших картинок с преподавателями
		{
			// Копирование исходника
			resultImg := image.NewRGBA(img.Bounds())
			copy(resultImg.Pix, img.Pix)

			for _, currentCell := range teachers[teacherName] {
				im.HighlightRect(resultImg, currentCell.Rect, ColorHighlightMain)

				// Прямоугольник для выделения времени
				var timeRect image.Rectangle

				// Для CellType_Merged
				if currentCell.CellType == CellType_Merged {
					lastCell := currentCell.partsOfMergedCell[len(currentCell.partsOfMergedCell)-1]

					timeRect = image.Rect(
						pageInfo.Cells[0][currentCell.J].Rect.Min.X,
						pageInfo.Cells[0][currentCell.J].Rect.Min.Y,
						pageInfo.Cells[2][lastCell.J].Rect.Max.X,
						pageInfo.Cells[2][lastCell.J].Rect.Max.Y,
					)
				} else {
					timeRect = image.Rect(
						pageInfo.Cells[0][currentCell.J].Rect.Min.X,
						pageInfo.Cells[0][currentCell.J].Rect.Min.Y,
						pageInfo.Cells[2][currentCell.J].Rect.Max.X,
						pageInfo.Cells[2][currentCell.J].Rect.Max.Y,
					)
				}

				im.HighlightRect(
					resultImg,
					timeRect,
					ColorHighlightMain,
				)

				if currentCell.IsChanged {
					changedTeachers <- teacherName
				}
			}

			drawAllChangeMarks(pageInfo, resultImg)

			// Сохранение изображения
			fileutils.SaveImg(
				resultImg,
				path.Join(
					saveFolderPath,
					teacherName+".png",
				),
			)
		}
	}

	// Найденные преподаватели
	var (
		foundTeachers = make(map[string]struct{})

		wg   = sync.WaitGroup{}
		pool = make(chan struct{}, Settings.MaxWorkers)
	)

	// Параллельная обработка
	for teacherName := range teachers {
		foundTeachers[teacherName] = struct{}{}

		wg.Add(1)
		go func() {
			pool <- struct{}{}

			makeImages(teacherName)

			<-pool
			wg.Done()
		}()
	}

	// Ожидаем завершения
	go func() {
		wg.Wait()
		close(changedTeachers)
	}()
	for teacherName := range changedTeachers {
		pageInfo.ChangedTeachers[teacherName] = struct{}{}
	}

	// Все преподаватели
	for teacherName := range foundTeachers {
		pageInfo.FoundTeachers = append(pageInfo.FoundTeachers, teacherName)
	}
}

func GetLineCellAndTime(fileInfo *PageData, img image.Image, targetCell *CellData, margin IndentInfo) image.Image {
	if targetCell == nil {
		return nil
	}

	// Область ячейки
	rectCell := image.Rect(
		targetCell.Rect.Min.X, // Для ровного слияния общей линии
		targetCell.Rect.Min.Y-targetCell.Border[1]+1-margin[1], // Для ровного слияния верхней линии
		targetCell.Rect.Max.X+targetCell.Border[2]+margin[2],
		targetCell.Rect.Max.Y+targetCell.Border[3]-1+margin[3], // Для ровного слияния нижней линии
	)

	// Область времени
	timeRect := image.Rect(
		fileInfo.Cells[0][0].Rect.Min.X-fileInfo.Cells[0][0].Border[0]-margin[0],
		rectCell.Min.Y,
		fileInfo.Cells[HeaderColumnsCount-1][0].Rect.Max.X+fileInfo.Cells[HeaderColumnsCount-1][0].Border[2]-1, // Для ровного слияния общей линии
		rectCell.Max.Y,
	)

	// Создание результата
	result := im.MergeImagesHorizontally(
		[]image.Image{
			im.CropRect(img, timeRect),
			im.CropRect(img, rectCell),
		},
	)

	return result
}
