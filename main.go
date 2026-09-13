package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"pdfDataExtractor/fileutils"
	"strconv"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"github.com/ledongthuc/pdf"
)

const (
	// Количество ячеек, которые являются шапкой таблицы
	HeaderColumnsCount = 3 // Горизонтальные
	HeaderRowsCount    = 2 // Вертикальные
)

var (
	DPI = 300
)

func main() {
	var (
		pdfFilePath   string
		buildingIndex int
	)

	flag.StringVar(&pdfFilePath, "pdfPath", "", "Путь к файлу .pdf для обработки")
	flag.IntVar(&buildingIndex, "buildingIndex", 0, "Индекс корпуса (0 - первый, 1 - второй)")
	flag.Parse()

	required := 0
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "pdfPath" {
			required += 1
		}

		if f.Name == "buildingIndex" {
			required += 1
		}
	})

	if required < 2 {
		fmt.Println("Ошибка: пропущен обязательный флаг!")
		flag.Usage() // Показываем справку
		os.Exit(1)
	}

	processPdf(pdfFilePath, buildingIndex)
}

// makeBaseImageFromPdf из страницы pdf делает изображение
func makeBaseImageFromPdf(pdfFilePath, PathToBasicImg string, pageNumber int) {
	err := exec.Command("magick", "-density", strconv.Itoa(DPI), // Утилита и флаги
		pdfFilePath+fmt.Sprintf("[%d]", pageNumber), // Файл pdf и страница
		"-background", "white", "-flatten", "-quality", "100", // Флаги
		PathToBasicImg, // Путь для сохранения
	).Run()
	if err != nil {
		fmt.Printf("Ошибка при создании изображения: %v\n", err)
		return
	}
}

// processPdf полный процесс обработки pdf файла
func processPdf(pdfFilePath string, buildingIndex int) {
	var (
		// Путь к папке с картинками
		imagesBaseFolder = "rImages"

		pdfPageCount int
	)

	// Получение количества страниц в файле
	{
		f, r, err := pdf.Open(pdfFilePath)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		pdfPageCount = r.NumPage()
	}
	start := time.Now()

	// Обработка каждой страницы
	for pageNumber := range pdfPageCount {
		var (
			// Максимальное число страниц файла, для корректного расчёта смещения
			maxPages = 10

			// Смещение в названиях папок и файлов, для работы нескольких корпусов
			// Пример: индекс страницы 0 и индекс корпуса 0 -> 0; индекс страницы 2 индекс корпуса 1 -> 11
			offset = buildingIndex * maxPages

			// Папка, куда будут сохранены специальные версии картинок (для групп, преподавателей, дебага)
			specialImagesFolder = path.Join(imagesBaseFolder, strconv.Itoa(pageNumber+offset))

			// Путь к основной картинке с расписанием
			PathToBasicImg = path.Join(imagesBaseFolder, fmt.Sprintf("%d.png", pageNumber+offset))

			pageInfo = &PageData{BuildingIndex: buildingIndex}
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
				makeBaseImageFromPdf(pdfFilePath, PathToBasicImg, pageNumber)
				fmt.Println("Страница запечена, DPI:", DPI)

				// Чтение получившейся картинки
				pageInfo.tempData.img = fileutils.ReadImg(PathToBasicImg)

				// Обрезка фона базовой картинки
				CropImage(
					pageInfo,
					PathToBasicImg,
				)

				// Требование от телеграмм api, что бы длинна + ширина до 10 000
				rect := pageInfo.tempData.img.Rect
				if rect.Dx()+rect.Dy() >= 10000 {
					DPI -= 30
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
				pdfFilePath,
				pageNumber,
				path.Join(specialImagesFolder, "temp.png"),
			)

			// Подготавливаем картинки с подписями над маленькими версиями расписания
			pageInfo.tempData.init(
				pageInfo.BuildingIndex,
				pageInfo.Date,
				pageInfo.DateFontSize,
				pageInfo.Cells[0][0].Border[0],
			)

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
		}
	}

	// TEMP таймер
	fmt.Println("Обработанно за", time.Since(start))
}
