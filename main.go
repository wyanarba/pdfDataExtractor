package main

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/wyanarba/pdfDataExtractor"

	"aaa/internal/storage"
	"aaa/internal/storage/models"
)

func main() {
	var (
		pdfFileName   = "9_spo.pdf"
		buildingIndex = 0
	)

	storage.Connect()

	pdfDataExtractor.Settings = pdfDataExtractor.SettingsS{
		BaseImagesPath:  "files/rImages",
		BasePdfPath:     "files/lastPdfs",
		BaseStoragePath: "files/storage",

		MaxWorkers: 8,

		LoadFileInfoFunc: storage.LoadDayInfo,

		Logger: pdfDataExtractor.Settings.Logger,
	}

	pdfDataExtractor.ProcessPdf(
		pdfFileName,
		buildingIndex,
	)

	//updateFileStorage(fileInfo, buildingIndex, pdfFileName)

	storage.Disconnect()
}

// updateFileStorage обновляет папку storage и записи в бд, поле обработки нового файла.
func updateFileStorage(fileInfo *pdfDataExtractor.FileData, buildingIndex int, pdfFileName string) {
	var (
		newStorageFolder = path.Join("files/storage", fileInfo.PdfFileHash)
	)

	os.Mkdir(newStorageFolder, 0777)

	copyFile(
		"files/lastPdfs/"+pdfFileName,
		path.Join(newStorageFolder, "file.pdf"),
	)

	copyFile(
		"files/lastPdfs/"+fmt.Sprintf("%d_Data.json", buildingIndex),
		path.Join(newStorageFolder, "file.json"),
	)

	// Обновление записей в бд
	for pageIdx := range fileInfo.PagesCount {
		page := &fileInfo.PagesInfo[pageIdx]

		dailyPage := models.DailyPage{
			Hash:     fileInfo.PdfFileHash,
			Page:     pageIdx,
			Day:      page.DateDayNumber,
			Building: buildingIndex,
		}

		storage.SaveDayInfo(dailyPage)
	}
}

func copyFile(src, dst string) error {
	// Открываем исходный файл
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Создаем новый файл для записи
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Копируем данные
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Синхронизируем запись на диск
	return destFile.Sync()
}
