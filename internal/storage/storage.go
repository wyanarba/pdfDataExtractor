package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	models2 "github.com/wyanarba/pdfDataExtractor/models"

	"aaa/internal/storage/models"
)

var (
	client     *mongo.Client
	collection *mongo.Collection
)

func Connect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Создаем клиент и подключаемся
	{
		var err error
		uri := "mongodb://localhost:27017"

		client, err = mongo.Connect(options.Client().ApplyURI(uri))
		if err != nil {
			log.Fatalf("Ошибка подключения: %v", err)
		}
	}

	// Проверяем соединение пингом
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB недоступна: %v", err)
	}
	fmt.Println("Успешное подключение к MongoDB!")

	collection = client.Database("pdfDETest").Collection("daily_pages")
}

func Disconnect() {
	if err := client.Disconnect(context.Background()); err != nil {
		log.Fatalf("Ошибка отключения: %v", err)
	}
}

func LoadDayInfo(dayId int, buildingIndex int) (dailyPage models2.DailyPage, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dailyPageInternal := models.DailyPage{}
	filter := bson.M{"day": dayId, "building": buildingIndex}

	err = collection.FindOne(ctx, filter).Decode(&dailyPageInternal)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("Документ не найден")
			return
		}
		log.Fatalf("Ошибка поиска: %v", err)
	}

	dailyPage.Day = dailyPageInternal.Day
	dailyPage.Hash = dailyPageInternal.Hash
	dailyPage.Page = dailyPageInternal.Page

	return
}

func SaveDayInfo(dailyPage models.DailyPage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"day": dailyPage.Day, "building": dailyPage.Building}

	// Включаем Upsert: если не найдено — вставить
	opts := options.Replace().SetUpsert(true)

	_, err := collection.ReplaceOne(ctx, filter, dailyPage, opts)

	return err
}
