package pdfDataExtractor

import (
	"image"
	"image/color"
	"image/draw"
)

type CellData struct {
	Rect        image.Rectangle
	Border      IndentInfo
	CompareRect image.Rectangle

	I, J int

	CellType CellType

	MergedCell        *CellData
	dividedParentCell *CellData
	partsOfMergedCell []*CellData
	DividedParts      [2]*CellData

	AllText string

	IsChanged        bool
	IsChangedEarlier bool
}

func MakeCellData(i, j int, rect image.Rectangle, border IndentInfo) (cd CellData) {
	cd.I = i
	cd.J = j
	cd.Rect = rect
	cd.Border = border

	cd.CompareRect = image.Rect(
		rect.Min.X-border[0],
		rect.Min.Y-border[1],
		rect.Max.X,
		rect.Max.Y,
	)

	return
}

// drawChangedMark отрисовывает надпись о изменении ячейки.
func (cell CellData) drawChangedMark(pageData *PageData, img *image.RGBA) {
	// Картинка с надписью
	var textImg image.Image
	if cell.IsChanged {
		textImg = pageData.tempData.changedImg
	} else if cell.IsChangedEarlier {
		textImg = pageData.tempData.changedImg2
	} else {
		return
	}

	// Отрисовка
	mask := image.NewUniform(color.Alpha{A: 160})
	r := image.Rectangle{
		Min: image.Point{cell.Rect.Max.X - textImg.Bounds().Dx(), cell.Rect.Min.Y},
		Max: image.Point{cell.Rect.Max.X, cell.Rect.Max.Y},
	}
	draw.DrawMask(img, r, textImg, textImg.Bounds().Min, mask, image.Point{}, draw.Over)
}

type CellType string

var (
	CellType_Basic         CellType = "basic"         // Обычная ячейка таблицы
	CellType_Group         CellType = "group"         // Тонкая ячейка с названием группы
	CellType_Service       CellType = "service"       // Ячейка входящая в шапку таблицы (сверху и слева)
	CellType_PartOfMerged  CellType = "partOfMerged"  // Ячейка являющаяся частью объединенной ячейки
	CellType_PartOfDivided CellType = "partOfDivided" // Ячейка являющаяся частью разделённой ячейки, на 2 подгруппы
	CellType_Divided       CellType = "divided"       // Ячейка которая была разделена на 2 подгруппы
	CellType_Merged        CellType = "merged"        // Большая ячейка, (несколько объединённых ячеек) объединение происходит только сверху и/или снизу.
)
