package helpers

import (
	"context"
	"fmt"
	"oiynlike/models"
	"time"

	"github.com/robfig/cron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Scheduler структура для работы с шедулером
type Scheduler struct {
	c *cron.Cron
}

var gameCardCollection *mongo.Collection // Импортируйте вашу коллекцию MongoDB здесь

// Инициализация шедулера
func InitScheduler() {
	scheduler := &Scheduler{
		c: cron.New(),
	}

	// Регистрация функции обновления статуса gameCard
	scheduler.c.AddFunc("@every 1m", scheduler.updateGameCardStatus)

	// Запуск шедулера
	scheduler.c.Start()
}

// Функция для обновления статуса gameCard
func (s *Scheduler) updateGameCardStatus() {
	// Получение текущей даты и времени
	currentTime := time.Now()

	// Формирование фильтра для поиска gameCard со статусом "active" и scheduled_time в прошлом
	filter := bson.M{
		"status":         "active",
		"scheduled_time": bson.M{"$lt": currentTime},
	}

	// Получение активных gameCard, у которых scheduled_time в прошлом
	cursor, err := gameCardCollection.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Error querying database:", err)
		return
	}
	defer cursor.Close(context.TODO())

	// Обновление статуса для найденных gameCard
	for cursor.Next(context.TODO()) {
		var gameCard models.GameCard
		if err := cursor.Decode(&gameCard); err != nil {
			fmt.Println("Error decoding results:", err)
			continue
		}

		// Обновление статуса на "inactive" и сохранение изменений в базе данных
		gameCard.Status = "inactive"
		update := bson.M{"$set": bson.M{"status": "inactive"}}
		if _, err := gameCardCollection.UpdateOne(context.TODO(), bson.M{"_id": gameCard.ID}, update); err != nil {
			fmt.Println("Error updating gameCard status:", err)
		}
	}

	fmt.Println("GameCard statuses updated.")
}
