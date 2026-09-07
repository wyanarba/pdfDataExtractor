package main

// getUpperGroupCell по ключевой ячейке (keyCell) возвращает gCell (ссылку на данные о ячейке с типом CellType_Group, находящеюся прямо над ней) или nil.
func getUpperGroupCell(fileInfo *PageData, keyCell *CellData) (gCell *CellData) {
	i := keyCell.I

	for j := keyCell.J; j >= 0; j-- {
		if fileInfo.Cells[i][j].CellType == CellType_Group {
			gCell = fileInfo.Cells[i][j]
			return
		}
	}

	return nil
}

// getLastCellBeforeUnderGroupCell по ключевой ячейке (keyCell).
// Возвращает foundCell, ссылку на данные о ячейке.
// Которая идёт перед более нижней ячейкой с типом CellType_Group.
// Или последней ячейкой таблицы, в колонне, если такой нету.
func getLastCellBeforeUnderGroupCell(fileInfo *PageData, keyCell *CellData) (foundCell *CellData) {
	i := keyCell.I

	for j := keyCell.J; j < fileInfo.CountRows; j++ {
		if fileInfo.Cells[i][j].CellType == CellType_Group {
			return fileInfo.Cells[i][j-1]
		}
	}

	// Возвращаем последнюю ячейку в колонне, потому что не нашли другой
	return fileInfo.Cells[i][fileInfo.CountRows-1]
}

// getDistanceToGroupCell получает расстояние до ближайшей GroupCell.
// GroupCell находится в этой же колонне таблицы и выше, чем целевая ячейка (keyCell).
func getDistanceToGroupCell(fileInfo *PageData, keyCell *CellData) (distance int) {
	targetCell := getUpperGroupCell(fileInfo, keyCell)

	if targetCell == nil {
		return -1
	}

	return keyCell.J - targetCell.J
}
