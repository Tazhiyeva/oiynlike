package controllers

import (
	"context"
	"fmt"
	"net/http"
	database "oiynlike/database"
	"oiynlike/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var chatsCollection *mongo.Collection = database.OpenCollection("chats")

func GetUserChatsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		// Извлекаем user_id из JWT
		userID, _ := c.Get("uid")
		userIDString := fmt.Sprintf("%v", userID)

		// Получаем чаты пользователя из базы данных
		cursor, err := chatsCollection.Find(ctx, bson.M{"members.user_id": userIDString})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user chats"})
			return
		}
		defer cursor.Close(ctx) // Close the cursor once done

		// Iterate through the cursor and decode each chat
		var chats []models.Chat
		if err := cursor.All(ctx, &chats); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding user chats"})
			return
		}

		// Отправляем список чатов пользователю
		c.JSON(http.StatusOK, gin.H{"chats": chats})
	}
}

func CreateChatIfNeeded(c *gin.Context, gameCard *models.GameCard) error {
	// Создаем чат, если число присоединенных пользователей равно максимальному числу игроков
	if len(gameCard.MatchedPlayers) == gameCard.MaxPlayers {
		chat := models.Chat{
			Title:      gameCard.Title,
			GameCardID: gameCard.ID,
			Members:    make([]models.Sender, 0),
		}
		// Добавляем хост-пользователя в список участников чата
		hostUser := models.Sender{
			FirstName: gameCard.HostUser.FirstName,
			LastName:  gameCard.HostUser.LastName,
			UserID:    gameCard.HostUser.UserID,
			PhotoURL:  gameCard.HostUser.PhotoURL,
		}
		chat.Members = append(chat.Members, hostUser)
		// Добавляем присоединенных пользователей в список участников чата
		for _, player := range gameCard.MatchedPlayers {
			matchedUser := models.Sender{
				FirstName: player.FirstName,
				LastName:  player.LastName,
				UserID:    player.UserID,
				PhotoURL:  player.PhotoURL,
			}
			chat.Members = append(chat.Members, matchedUser)
		}
		// Сохраняем чат в базе данных
		err := createChat(c, &chat)
		if err != nil {
			return fmt.Errorf("Error creating chat: %v", err)
		}
	}
	return nil
}

func createChat(c *gin.Context, chat *models.Chat) error {
	// Определяем контекст для выполнения операции с базой данных

	ctx := context.Background()

	// Вставляем чат в коллекцию
	result, err := chatsCollection.InsertOne(ctx, chat)
	if err != nil {
		return fmt.Errorf("Error inserting chat into database: %v", err)
	}

	// Получаем вставленный ID чата
	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return fmt.Errorf("Failed to get inserted chat ID")
	}

	// Устанавливаем ID чата в структуре
	chat.ID = insertedID

	return nil
}

func LeaveChatHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		// Извлекаем user_id из JWT
		userID, _ := c.Get("uid")
		userIDString := fmt.Sprintf("%v", userID)

		// Получаем идентификатор чата, который пользователь покидает
		chatID := c.Param("chat_id")

		fmt.Println("Received chat_id:", chatID)

		objectID, err := primitive.ObjectIDFromHex(chatID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error with chat_id"})
			return
		}

		// Получаем чат из базы данных
		var chat models.Chat
		err = chatsCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&chat)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving chat data"})
			return
		}

		// Удаляем пользователя из списка участников чата
		var updatedMembers []models.Sender
		for _, member := range chat.Members {

			if member.UserID != userIDString {
				updatedMembers = append(updatedMembers, member)
			}
		}

		// Обновляем список участников чата
		_, err = chatsCollection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": bson.M{"members": updatedMembers}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating chat members"})
			return
		}

		// Отправляем успешный ответ пользователю
		c.JSON(http.StatusOK, gin.H{"msg": "User left the chat successfully"})
	}
}
