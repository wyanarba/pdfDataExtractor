package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"pdfDataExtractor/fileutils"
	"slices"
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

func processPDF() {
	var (
		// Путь к папке с картинками
		imagesBaseFolder = "rImages"

		pdfFilePath  string = "npo29.pdf"
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
	t := start

	for pageNumber := range pdfPageCount {
		var (
			// Папка, куда будут сохранены специальные версии картинок (для групп, преподавателей, дебага)
			specialImagesFolder = path.Join(imagesBaseFolder, strconv.Itoa(pageNumber))

			// Путь к основной картинке с расписанием
			PathToBasicImg = path.Join(imagesBaseFolder, fmt.Sprintf("%d.png", pageNumber))

			pageInfo = &PageData{BuildingNum: 0}
		)

		//Pdf page -> .png
		{
			err := exec.Command("magick", "-density", "400", // Утилита и флаги
				pdfFilePath+fmt.Sprintf("[%d]", pageNumber), // Файл pdf и страница
				"-background", "white", "-flatten", "-quality", "100", // Флаги
				PathToBasicImg, // Путь для сохранения
			).Run()
			if err != nil {
				fmt.Printf("Ошибка при создании изображения: %v\n", err)
				return
			}
			fmt.Println("Изображение получено")
		}
		// TEMP таймер
		fmt.Println("Запекание\\t", time.Since(t))
		t = time.Now()

		// Чтение получившейся картинки
		pageInfo.tempData.img = fileutils.ReadImg(PathToBasicImg)

		// Очистка specialImagesFolder от старого расписания
		{
			if err := os.RemoveAll(specialImagesFolder); err != nil {
				log.Fatalf("Ошибка при очистке: %v", err)
			}

			// Создание чистой папки.
			if err := os.MkdirAll(specialImagesFolder, 0755); err != nil {
				log.Fatalf("Ошибка при создании папки: %v", err)
			}
		}

		// TEMP таймер
		fmt.Println("Подготовка\\t", time.Since(t))
		t = time.Now()
		// Обработка изображения
		{
			// Обрезка фона базовой картинки
			CropImage(
				pageInfo,
				PathToBasicImg,
			)
			// TEMP таймер
			fmt.Println("Обрезка фона\\t", time.Since(t))
			t = time.Now()

			// Извлекаем информацию о ячейках таблицы по фото
			GetCellsFromImage(
				pageInfo,
				path.Join(specialImagesFolder, "temp.png"),
			)
			// TEMP таймер
			fmt.Println("Извлекаем информацию о ячейках\\t", time.Since(t))
			t = time.Now()

			// Извлекаем текст из страницы .pdf и распределяем его по ячейкам
			ExtractText(
				pageInfo,
				pdfFilePath,
				pageNumber,
				path.Join(specialImagesFolder, "temp.png"),
			)
			// TEMP таймер
			fmt.Println("Извлекаем текст из страницы\\t", time.Since(t))
			t = time.Now()

			// Подготавливаем картинки с подписями над маленькими версиями расписания
			pageInfo.tempData.init(
				pageInfo.BuildingNum,
				pageInfo.Date,
				pageInfo.DateFontSize,
				pageInfo.Cells[0][0].Border[0],
			)
			// TEMP таймер
			fmt.Println("Подготавливаем картинки с подписями\\t", time.Since(t))
			t = time.Now()

			// Создаём изображения для преподавателей
			CreateTeachersImages(
				pageInfo,
				specialImagesFolder,
			)
			// TEMP таймер
			fmt.Println("Создаём изображения для преподавателей\\t", time.Since(t))
			t = time.Now()

			// Создаём изображения для групп
			CreateGroupsImages(
				pageInfo,
				specialImagesFolder,
			)
			// TEMP таймер
			fmt.Println("Создаём изображения для групп\\t", time.Since(t))
			t = time.Now()
		}

		break
	}

	// TEMP таймер
	fmt.Println("ВСЕГО\\t", time.Since(start))
}

func main() {
	const (
		countBuildings = 2
	)

	var (
		scheduleFilesURL = [countBuildings]string{
			"https://rasp.vksit.ru/spo.pdf",
			"https://rasp.vksit.ru/npo.pdf",
		}

		scheduleFilesNames = [2]string{
			"spo.pdf",
			"npo.pdf",
		}

		downloadedFiles = [2][]byte{}
		currentFiles    = [2][]byte{}

		err error
	)

	// Чтение текущих файлов с диска при старте
	{
		for i := range countBuildings {
			currentFiles[i], err = os.ReadFile(scheduleFilesNames[i])

			if err != nil {
				log.Fatalln("Не удалось прочитать файл", scheduleFilesNames[i], ", при старте. Ошибка:", err.Error())
			}
		}
	}

	{
		// Сравнение текущих файлов с версией из URL
		for i := range countBuildings {

			// Скачивание файлов
			{
				resp, err := http.Get(scheduleFilesURL[i])
				if err != nil {
					log.Println("[ERROR] Не удалось скачать новую версию файла", scheduleFilesNames[i], "с сайта. Ошибка:", err.Error())
					continue
				}
				downloadedFiles[i], err = io.ReadAll(resp.Body)
				if err != nil {
					log.Println("[ERROR]", err.Error())
					continue
				}
			}

			// Сравнение
			fmt.Println(slices.Compare(downloadedFiles[i], currentFiles[i]))
		}
	}
}
