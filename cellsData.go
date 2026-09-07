package main

import "image"

type CellData struct {
	Rect        image.Rectangle
	Border      IndentInfo
	CompareRect image.Rectangle

	I, J int

	CellType CellType

	mergedCell        *CellData
	partsOfMergedCell []*CellData
	dividedParts      [2]*CellData

	AllText string
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
