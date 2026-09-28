package pdfDataExtractor

import (
	"github.com/wyanarba/pdfDataExtractor/basicLogger"
	"github.com/wyanarba/pdfDataExtractor/models"
)

// LoadFileInfo получает информацию о нужной странице и файле из бд.
type LoadFileInfo func(dayId int, buildingIndex int) (dailyPage models.DailyPage, err error)

type SettingsS struct {
	// Базовые пути
	BaseImagesPath  string
	BasePdfPath     string
	BaseStoragePath string

	MaxWorkers int

	LoadFileInfoFunc LoadFileInfo

	Logger interface {
		Debug(args ...interface{})
		Info(args ...interface{})
		Warn(args ...interface{})
		Error(args ...interface{})
		Panic(args ...interface{})
	}
}

var (
	Settings = SettingsS{

		Logger: basicLogger.BasicLogger{},
	}
)
