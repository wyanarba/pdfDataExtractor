package pdfDataExtractor

import (
	"image"
	"image/color"
	"image/draw"
	"strconv"

	"github.com/wyanarba/pdfDataExtractor/imageUtils"
	im "github.com/wyanarba/pdfDataExtractor/imageUtils"
)

type PageData struct {
	BuildingIndex int // Индекс корпуса (0, 1)
	PdfPageNumber int // Номер страницы в .pdf файле
	DPI           int // DPI для страницы

	// Информация о типе изменений
	ChangesInfo ChangesData

	// Дата
	Date           string  // Дата "13.06.2026г. (суббота)"
	DateShort      string  // Короткая дата "13.06.2026"
	DateFontSize   float64 // Размер шрифта надписи с датой, в пунктах
	DateDayNumber  int     // Особый формат даты (уникальный номер дня: time.Time.Unix() / 86400)
	DateWeekNumber int     // Особый формат даты (уникальный номер недели: y, w := time.Time.ISOWeek(); y*100 + w)

	// Размеры изображений
	boundsBeforeCut image.Rectangle // До обрезки
	boundsAfterCut  image.Rectangle // После обрезки
	Crop            CropInfo        // Информация об обрезке исходной фотографии

	// Ячейки
	Cells            [][]*CellData
	dividedPartCells []*CellData
	mergedCell       []*CellData

	// Размерность таблицы с расписанием
	CountCols, CountRows int

	// Найденные группы и преподаватели
	FoundGroups, FoundTeachers []string

	// Изменения (добавлю позже)
	ChangedGroups, ChangedTeachers map[string]struct{}

	// Для обнаружения пропажи в изменениях (был преподаватель и пропал)
	GroupCells, TeacherCells       map[string]([][2]int)
	oldGroupCells, oldTeacherCells map[string]([][2]int)

	doFindChanges bool // Была ли эта страница обработана ранее (особый формат изменений)

	// Структура с временными данными (нужными для создания изображений), не доступными из вне и не сохраняющимися куда либо
	tempData processingData
}

// Типы изменений
type ChangesData string

var (
	NewPage                ChangesData = "fullNew"
	NewPageDimensionsTable ChangesData = "fullNewTableDimensionsChanged"
	Changed                ChangesData = "changed"
	NonChanged             ChangesData = "nonChanged"
)

type CropInfo struct {
	CutOut image.Rectangle // Сколько удаленно со сторон
	Margin IndentInfo      // Оставленный отступ по краям, от содержимого (текста или таблицы)
}

type processingData struct {
	// Сама основная картинка и она же с подписанными изменениями
	img, imgCh *image.RGBA

	// Картинка с текстом "1 КОРПУС"
	buildNumLine image.Image

	// Картинка с текстом "13.06.2026г. (суббота)"
	dateLine image.Image

	// Линия для отступа между мини таблицей и датой
	spacerLine *image.RGBA

	changedImg, changedImg2 image.Image // Картинки с надписями изм.
}

// init подготавливает временные картинки для обработки расписания, по заданным аргументам.
// fBorder это ширина первой левой линии рамки таблицы.
func (pd *processingData) init(buildingNum int, date string, DPI int, dateFontSize float64, fBorder int) {
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

	// Картинки с надписями изм.
	pd.changedImg = imageUtils.RenderTextWithoutBackground(
		"изм.",
		timesFont,
		dateFontSize,
		float64(DPI),
		1,
		color.RGBA{255, 0, 0, 255},
	)

	pd.changedImg2 = imageUtils.RenderTextWithoutBackground(
		"изм.",
		timesFont,
		dateFontSize,
		float64(DPI),
		1,
		color.RGBA{255, 0, 255, 255},
	)
}

// X1, Y1, X2, Y2
type IndentInfo [4]int
