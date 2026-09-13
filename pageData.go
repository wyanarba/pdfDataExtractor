package main

import (
	"image"
	"image/color"
	"image/draw"
	im "pdfDataExtractor/imageUtils"
	"strconv"
)

type PageData struct {
	// Индекс корпуса (0, 1)
	BuildingIndex int

	// Номер страницы в .pdf файле
	PdfPageNumber int

	// Дата
	Date         string  // Дата "13.06.2026г. (суббота)"
	DateShort    string  // Короткая дата "13.06.2026"
	DateFontSize float64 // Размер шрифта надписи с датой, в пунктах

	// Размеры изображений
	boundsBeforeCut image.Rectangle // До обрезки
	boundsAfterCut  image.Rectangle // После обрезки
	Crop            CropInfo        // Информация об обрезке исходной фотографии

	// Ячейки
	Cells            [][]*CellData
	DividedPartCells []*CellData
	MergedCells      []*CellData

	// Размерность таблицы с расписанием
	CountCols, CountRows int

	// Структура с временными данными (нужными для создания изображений), не доступными из вне и не сохраняющимися куда либо
	tempData processingData
}

type CropInfo struct {
	CutOut image.Rectangle // Сколько удаленно со сторон
	Margin IndentInfo      // Оставленный отступ по краям, от содержимого (текста или таблицы)
}

type processingData struct {
	// Сама основная картинка
	img *image.RGBA

	// Картинка с текстом "1 КОРПУС"
	buildNumLine image.Image

	// Картинка с текстом "13.06.2026г. (суббота)"
	dateLine image.Image

	// Линия для отступа между мини таблицей и датой
	spacerLine *image.RGBA
}

// init подготавливает временные картинки для обработки расписания, по заданным аргументам.
// fBorder это ширина первой левой линии рамки таблицы.
func (pd *processingData) init(buildingNum int, date string, dateFontSize float64, fBorder int) {
	// Картинка с текстом "1 КОРПУС"
	pd.buildNumLine = im.RenderTextRectangle(
		strconv.FormatInt(int64(buildingNum+1), 10)+" корпус",
		timesBoldFont,
		dateFontSize,
		float64(DPI),
		0,
	)

	// Картинка с текстом "13.06.2026г. (суббота)"
	pd.dateLine = im.RenderTextRectangle(
		date,
		timesFont,
		dateFontSize,
		float64(DPI),
		int(float32(fBorder)*0.3),
	)

	// Линия для отступа между мини таблицей и датой
	pd.spacerLine = image.NewRGBA(image.Rect(0, 0, 5, int(float32(fBorder)*3.6)))
	draw.Draw(pd.spacerLine, pd.spacerLine.Bounds(), &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.Point{}, draw.Src)
}
