package pdfDataExtractor

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"log"
	"os"
	"os/exec"
	"path"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"github.com/ledongthuc/pdf"
	"github.com/wyanarba/pdfDataExtractor/fileutils"
)

const (
	// Количество ячеек, которые являются шапкой таблицы
	HeaderColumnsCount = 3 // Горизонтальные
	HeaderRowsCount    = 2 // Вертикальные
)

func ProcessPdf(pdfFilePath string, buildingIndex int) *FileData {
	Settings.Logger.Debug("aaaa")

	var fileData = FileData{
		PdfFileName:   pdfFilePath,
		buildingIndex: buildingIndex,
	}

	processPdf(&fileData)

	data, err := json.MarshalIndent(fileData, "", "    ")
	if err != nil {
		Settings.Logger.Error(err)
	}
	file, _ := os.OpenFile(
		path.Join(Settings.BasePdfPath, strconv.Itoa(buildingIndex)+"_Data.json"),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0777,
	)
	if err != nil {
		Settings.Logger.Error(err)
	}
	defer file.Close()
	file.Write(data)

	return &fileData
}

// makeBaseImageFromPdf из страницы pdf делает изображение
func makeBaseImageFromPdf(pdfFilePath, PathToBasicImg string, pageNumber, DPI int) {
	err := exec.Command("magick", "-density", strconv.Itoa(DPI), // Утилита и флаги
		pdfFilePath+fmt.Sprintf("[%d]", pageNumber), // Файл pdf и страница
		"-background", "white", "-flatten", "-quality", "100", // Флаги
		PathToBasicImg, // Путь для сохранения
	).Run()
	if err != nil {
		Settings.Logger.Error("Ошибка при создании изображения: ", err)
		return
	}
}

// processPdf полный процесс обработки pdf файла
func processPdf(fileData *FileData) error {

	// Получение количества страниц в файле
	{
		f, r, err := pdf.Open(path.Join(Settings.BasePdfPath, fileData.PdfFileName))
		if err != nil {
			panic(err)
		}
		defer f.Close()

		fileData.PagesCount = r.NumPage()

		fileData.PagesInfo = make([]PageData, fileData.PagesCount)
	}

	// Получение хэша файла
	{
		var (
			data []byte
			err  error
		)
		if data, err = os.ReadFile(path.Join(Settings.BasePdfPath, fileData.PdfFileName)); err != nil {
			errors.Join(errors.New("Не удалось прочитать файл, для получения хэша"), err)
		} else {
			hash := sha256.Sum256(data)
			fileData.PdfFileHash = fmt.Sprintf("%x", hash)
		}
	}

	start := time.Now()

	// Обработка каждой страницы
	var (
		pool = make(chan struct{}, Settings.MaxWorkers)
		wg   = sync.WaitGroup{}
	)
	for pageNumber := range fileData.PagesCount {
		wg.Add(1)
		go func() {
			pool <- struct{}{}

			Settings.Logger.Info("Запущен ", pageNumber+1)
			processPdfPage(fileData, pageNumber, Settings.BaseImagesPath)

			<-pool
			wg.Done()
		}()

	}
	wg.Wait()

	// TEMP таймер
	Settings.Logger.Info("Обработано за", time.Since(start))

	return nil
}

func processPdfPage(fileData *FileData, pageNumber int, imagesBaseFolder string) {
	// Добавление страницы
	fileData.PagesInfo[pageNumber] = PageData{
		BuildingIndex: fileData.buildingIndex,
		PdfPageNumber: pageNumber,
		DPI:           300,

		doFindChanges:   true,
		ChangedGroups:   make(map[string]struct{}),
		ChangedTeachers: make(map[string]struct{}),

		GroupCells:      make(map[string]([][2]int)),
		TeacherCells:    make(map[string]([][2]int)),
		oldGroupCells:   make(map[string]([][2]int)),
		oldTeacherCells: make(map[string]([][2]int)),
	}

	var (
		// Максимальное число страниц файла, для корректного расчёта смещения
		maxPages = 10

		// Смещение в названиях папок и файлов, для работы нескольких корпусов
		// Пример: индекс страницы 0 и индекс корпуса 0 -> 0; индекс страницы 2 индекс корпуса 1 -> 11
		offset = fileData.buildingIndex * maxPages

		// Папка, куда будут сохранены специальные версии картинок (для групп, преподавателей, дебага)
		specialImagesFolder = path.Join(imagesBaseFolder, strconv.Itoa(pageNumber+offset))

		// Путь к основной картинке с расписанием
		PathToBasicImg = path.Join(imagesBaseFolder, fmt.Sprintf("%d.png", pageNumber+offset))

		pageInfo = &fileData.PagesInfo[pageNumber]
	)

	// Очистка specialImagesFolder от старого расписания
	{
		// Проверка на существование папки
		info, err := os.Stat(specialImagesFolder)
		if err == nil && info.IsDir() {

			// Удаление папки с её содержимым
			if err := os.RemoveAll(specialImagesFolder); err != nil {
				log.Fatalf("Очистка папки не удалась: %v", err)
			}
		}

		// Создание новой пустой папки.
		if err := os.MkdirAll(specialImagesFolder, 0755); err != nil {
			log.Fatalf("Ошибка при создании папки: %v", err)
		}
	}

	// Обработка изображения
	{
		// Уменьшать DPI, пока фото не подойдёт под требования телеграмма
		imgIsDone := false
		for imgIsDone != true {
			//Pdf page -> .png
			makeBaseImageFromPdf(
				path.Join(Settings.BasePdfPath, fileData.PdfFileName),
				PathToBasicImg,
				pageNumber,
				pageInfo.DPI,
			)
			Settings.Logger.Info("Страница запечена, DPI:", pageInfo.DPI)

			// Чтение получившейся картинки
			pageInfo.tempData.img = fileutils.ReadImg(PathToBasicImg)

			// Обрезка фона базовой картинки
			CropImage(
				pageInfo,
				PathToBasicImg,
			)

			pageInfo.tempData.imgCh = fileutils.ReadImg(PathToBasicImg)

			// Требование от телеграмм api, что бы длинна + ширина до 10 000
			rect := pageInfo.tempData.img.Rect
			if rect.Dx()+rect.Dy() >= 10000 {
				pageInfo.DPI -= 30
			} else {
				imgIsDone = true
			}
		}

		// Извлекаем информацию о ячейках таблицы по фото
		GetCellsFromImage(
			pageInfo,
			path.Join(specialImagesFolder, "temp.png"),
		)

		// Извлекаем текст из страницы .pdf и распределяем его по ячейкам
		ExtractText(
			pageInfo,
			path.Join(Settings.BasePdfPath, fileData.PdfFileName),
			pageNumber,
			path.Join(specialImagesFolder, "temp.png"),
		)

		// Подготавливаем картинки с подписями над маленькими версиями расписания
		pageInfo.tempData.init(
			pageInfo.BuildingIndex,
			pageInfo.Date,
			pageInfo.DPI,
			pageInfo.DateFontSize,
			pageInfo.Cells[0][0].Border[0],
		)

		// Отслеживание изменений
		trackPageChanges(pageInfo, fileData.PdfFileHash)
		drawAllChangeMarks(pageInfo, pageInfo.tempData.imgCh)

		// Создаём изображения для преподавателей
		CreateTeachersImages(
			pageInfo,
			specialImagesFolder,
		)

		// Создаём изображения для групп
		CreateGroupsImages(
			pageInfo,
			specialImagesFolder,
		)

		// Сохранение картинки с надписями об изменениях
		fileutils.SaveImg(pageInfo.tempData.imgCh, PathToBasicImg)

		// Определяем тип изменений страницы (ChangesData)
		{
			if pageInfo.ChangesInfo == "" {
				if (len(pageInfo.ChangedGroups) == 0) && (len(pageInfo.ChangedTeachers) == 0) {
					pageInfo.ChangesInfo = NonChanged
				} else {
					pageInfo.ChangesInfo = Changed
				}
			}
		}
	}
}

func trackPageChanges(newPage *PageData, newHash string) {
	newPage.ChangesInfo = NewPage

	var (
		oldPage PageData
		oldHash string
	)

	// Получение старой версии страницы
	{
		// Нет функции для загрузки предыдущей версии страницы
		if Settings.LoadFileInfoFunc == nil {
			return
		}

		dInfo, err := Settings.LoadFileInfoFunc(newPage.DateDayNumber, newPage.BuildingIndex)

		// Страница не найдена
		if err != nil || dInfo.Day != newPage.DateDayNumber {
			return
		}

		// Чтение oldFileInfo
		var oldFileInfo = &FileData{}
		{
			// 1. Открываем файл для чтения
			file, err := os.Open(path.Join(Settings.BaseStoragePath, dInfo.Hash, "file.json"))
			if err != nil {
				log.Fatalf("Ошибка при открытии файла: %v", err)
			}
			defer file.Close() // Обязательно закрываем файл в конце

			// 2. Создаем дешифратор и декодируем прямо из файла в переменную
			decoder := json.NewDecoder(file)
			if err := decoder.Decode(oldFileInfo); err != nil {
				log.Fatalf("Ошибка при разборе JSON: %v", err)
			}
		}

		// Неверно прочитанный файл
		if oldFileInfo.PagesCount < dInfo.Page || oldFileInfo.PagesInfo[dInfo.Page].DateDayNumber != newPage.DateDayNumber {
			return
		}

		oldPage = oldFileInfo.PagesInfo[dInfo.Page]
		oldHash = oldFileInfo.PdfFileHash
	}

	// Копирует информацию о изменении ячеек
	copyChangeStatus := func(newCell, oldCell *CellData) {
		if oldCell.IsChanged {
			newCell.IsChanged = true
		}
		if oldCell.IsChangedEarlier {
			newCell.IsChangedEarlier = true
		}
	}

	newPage.ChangesInfo = ""

	// Сравнивает две ячейки по тексту
	compareCell := func(newCell, oldCell *CellData) {
		var (
			newLines = []string{}
			oldLines = []string{}
		)

		// Получение линий
		newLines = strings.Split(newCell.AllText, "\n")
		oldLines = strings.Split(oldCell.AllText, "\n")

		// Изменена ячейка ранее
		if oldCell.IsChangedEarlier || oldCell.IsChanged {
			newCell.IsChangedEarlier = true
		}

		// Разное количество линий
		if len(oldLines) != len(newLines) {
			newCell.IsChanged = true
			return
		}

		// Сравнение по линиям
		for idx := range newLines {
			if slices.Index(oldLines, newLines[idx]) == -1 {
				newCell.IsChanged = true
				return
			}
		}
	}

	// Тот же самый файл что и раньше
	if newHash == oldHash {
		// Сравнение таблиц

		newPage.doFindChanges = false

		for colIdx := range newPage.CountCols {
			for rowIdx := range newPage.CountRows {
				var (
					newCell = newPage.Cells[colIdx][rowIdx]
					oldCell = oldPage.Cells[colIdx][rowIdx]
				)

				switch newCell.CellType {
				// Разделённая на две
				case CellType_Divided:
					copyChangeStatus(newCell.DividedParts[0], oldCell.DividedParts[0])
					copyChangeStatus(newCell.DividedParts[1], oldCell.DividedParts[1])
					copyChangeStatus(newCell, oldCell)

				// Часть от большой ячейки
				case CellType_PartOfDivided:
					copyChangeStatus(newCell, oldCell)
					if newCell.IsChanged {
						newCell.MergedCell.IsChanged = true
					}
					if newCell.IsChangedEarlier {
						newCell.MergedCell.IsChangedEarlier = true
					}

				// Обычная ячейка
				default:
					copyChangeStatus(newCell, oldCell)
				}
			}
		}

		newPage.ChangedGroups = oldPage.ChangedGroups
		newPage.ChangedTeachers = oldPage.ChangedTeachers

		return
	}

	// Сравнение первой страницы
	{
		// Сравнение таблиц
		var (
			colCount = min(newPage.CountCols, oldPage.CountCols)
			rowCount = min(newPage.CountRows, oldPage.CountRows)
		)

		// Если размерность таблиц не одинаковая
		if (oldPage.CountCols != newPage.CountCols) || (oldPage.CountRows != newPage.CountRows) {
			newPage.ChangesInfo = NewPageDimensionsTable
			newPage.doFindChanges = false
			return
		}

		for colIdx := range colCount {
			for rowIdx := range rowCount {
				var (
					newCell = newPage.Cells[colIdx][rowIdx]
					oldCell = oldPage.Cells[colIdx][rowIdx]
				)

				// Разные типы ячеек
				if newCell.CellType != oldCell.CellType {
					newCell.IsChanged = true
					if newCell.CellType == CellType_PartOfMerged {
						newCell.MergedCell.IsChanged = true
					}
				}

				switch newCell.CellType {
				// Разделённая на две
				case CellType_Divided:
					compareCell(newCell.DividedParts[0], oldCell.DividedParts[0])
					compareCell(newCell.DividedParts[1], oldCell.DividedParts[1])
					compareCell(newCell, oldCell)

				// Часть от большой ячейки
				case CellType_PartOfDivided:
					compareCell(newCell, oldCell)
					if newCell.IsChanged {
						newCell.MergedCell.IsChanged = true
					}
					if newCell.IsChangedEarlier {
						newCell.MergedCell.IsChangedEarlier = true
					}

				// Обычная ячейка
				default:
					compareCell(newCell, oldCell)
				}
			}
		}

		newPage.oldGroupCells = oldPage.GroupCells
		newPage.oldTeacherCells = oldPage.TeacherCells
	}
}

func drawAllChangeMarks(pageInfo *PageData, img *image.RGBA) {

	for i := range pageInfo.CountCols {
		for j := range pageInfo.CountRows {
			cell := pageInfo.Cells[i][j]

			switch cell.CellType {
			case CellType_PartOfMerged:
				continue
			case CellType_Divided:
				cell.DividedParts[0].drawChangedMark(pageInfo, img)
				cell.DividedParts[1].drawChangedMark(pageInfo, img)
			default:
				cell.drawChangedMark(pageInfo, img)
			}
		}
	}

	for _, cell := range pageInfo.mergedCell {
		cell.drawChangedMark(pageInfo, img)
	}
}
