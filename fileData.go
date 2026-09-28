package pdfDataExtractor

type FileData struct {
	// Исходник pdf
	PdfFileName string // Имя файла .pdf
	PdfFileHash string // md5 хеш для проверки

	// Номер корпуса
	buildingIndex int

	// Информация о страницах файла
	PagesCount int
	PagesInfo  []PageData
}
