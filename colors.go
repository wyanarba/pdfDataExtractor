package main

import "image/color"

var (
	// Базовые контрастные цвета
	ColorNN1 = color.NRGBA{0xFF, 0x33, 0x33, 150} // Ярко-красный
	ColorNN2 = color.NRGBA{0x33, 0xCC, 0x33, 150} // Насыщенный зеленый
	ColorNN3 = color.NRGBA{0x33, 0x66, 0xFF, 150} // Ярко-синий
	ColorNN4 = color.NRGBA{0xFF, 0xCC, 0x00, 150} // Желтый / Золотой
	ColorNN5 = color.NRGBA{0xFF, 0x33, 0xCC, 150} // Пурпурный / Маджента
	ColorNN6 = color.NRGBA{0x00, 0xCC, 0xFF, 150} // Голубой / Циан

	// Дополнительные контрастные оттенки
	ColorNN7  = color.NRGBA{0xFF, 0x66, 0x00, 150} // Оранжевый
	ColorNN8  = color.NRGBA{0x99, 0x33, 0xFF, 150} // Фиолетовый
	ColorNN9  = color.NRGBA{0x99, 0xFF, 0x00, 150} // Лаймовый
	ColorNN10 = color.NRGBA{0xFF, 0x99, 0xCC, 150} // Розовый
	ColorNN11 = color.NRGBA{0x00, 0x99, 0x99, 150} // Темно-бирюзовый / Teal
	ColorNN12 = color.NRGBA{0x99, 0x66, 0x33, 150} // Коричневый
	ColorNN13 = color.NRGBA{0x66, 0xFF, 0x99, 100} // Мятный
	ColorNN14 = color.NRGBA{0x00, 0x33, 0x99, 150} // Темно-синий / Navy
	ColorNN15 = color.NRGBA{0xFF, 0x99, 0x66, 150} // Коралловый

	// Специфичные, но читаемые оттенки
	ColorNN16 = color.NRGBA{0x99, 0x99, 0x00, 150} // Оливковый
	ColorNN17 = color.NRGBA{0xCC, 0x99, 0xFF, 150} // Лавандовый
	ColorNN18 = color.NRGBA{0x99, 0x00, 0x33, 150} // Бордовый
	ColorNN19 = color.NRGBA{0x99, 0xCC, 0xFF, 150} // Светло-голубой
	ColorNN20 = color.NRGBA{0x80, 0x80, 0x80, 150} // Нейтральный серый

	// Цвет для выделения ячеек (для преподавателя / группы) в больших версиях расписания
	ColorHighlightMain = color.NRGBA{0, 255, 0, 64} // Нейтральный серый
)
