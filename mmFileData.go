package main

type MmFileData struct {
	// Исходник pdf
	PdfFileName string // Имя файла .pdf
	PdfFileHash string // md5 хеш для проверки

	// Номер корпуса
	buildingNumber int

	// Информация о страницах файла
	CountPages int
	PagesInfo  []PageData
}
