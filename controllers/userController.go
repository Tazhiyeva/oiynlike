package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"oiynlike/database"
	helper "oiynlike/helpers"

	"oiynlike/models"
	"time"

	"github.com/gin-gonic/gin"
	validator "github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection("users")
var validate = validator.New()

func HashPassword(providedPassword string) string {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(providedPassword), 14)

	if err != nil {
		log.Panic(err)
	}
	return string(hashPassword)

}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := " "

	if err != nil {
		msg = fmt.Sprintf("email or password is incorrect")
		check = false
	}
	return check, msg
}

// @Summary User signup
// @Description Creates a new user account with the provided details and returns tokens and user information
// @ID user-signup
// @Tags Users
// @Accept json
// @Produce json
// @Param user body models.User true "User details for signup"
// @Success 200 {object} models.User "User created successfully"
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 409 {object} ErrorResponse "User already exists"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /users/signup [post]
func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
			return
		}

		if userCollection == nil || ctx == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
			return
		}

		// Проверка почты
		emailCount, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while checking for the email"})
			return
		}

		password := HashPassword(user.Password)
		user.Password = password

		// Проверка номера телефона
		// phoneCount, err := userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		// if err != nil {
		// 	log.Panic(err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while checking for the phone number"})
		// 	return
		// }

		// if emailCount > 0 || phoneCount > 0 {
		if emailCount > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}

		user.CreatedAt, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while parsing time"})
			return
		}

		user.UpdatedAt, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while parsing time"})
			return
		}

		user.ID = primitive.NewObjectID()
		user_id := user.ID.Hex()
		user.UserId = user_id

		fmt.Printf("Email: %s, FirstName: %s, LastName: %s, UserType: %s, UserID: %s\n", user.Email, user.FirstName, user.LastName, user.UserType, user.UserId)

		token, refreshToken, err := helper.GenerateAllTokens(user.Email, user.FirstName, user.LastName, user.UserType, user.UserId)
		if err != nil {
			log.Printf("Error generating JWT: %v", err)
			msg := fmt.Sprint("couldn't generate jwt")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		user.Token = token
		user.RefreshToken = refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := fmt.Sprint("user hasn't created ")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

// @Summary User login
// @Description Logs in a user with the provided email and password, returning the user details and tokens
// @ID user-login
// @Tags Users
// @Accept json
// @Produce json
// @Param user body models.User true "User credentials for login"
// @Success 200 {object} models.User "Successful login"
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /users/login [post]
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email or password is incorrect"})
			return
		}

		passwordIsValid, msg := VerifyPassword(user.Password, foundUser.Password)
		defer cancel()
		if passwordIsValid != true {
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		if foundUser.Email == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		}

		token, refreshToken, _ := helper.GenerateAllTokens(foundUser.Email, foundUser.FirstName, foundUser.LastName, foundUser.UserType, foundUser.UserId)
		helper.UpdateAllTokens(token, refreshToken, foundUser.UserId)
		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.UserId}).Decode(&foundUser)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, foundUser)

	}

}

func updateProfile(ctx context.Context, userID string, updatedUser models.User) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid UserID format")
	}

	// Формируем фильтр по _id
	filter := bson.M{"_id": objectID}

	updateFields := bson.D{
		{Key: "$set", Value: bson.D{}},
	}

	updateFields[0].Value = append(updateFields[0].Value.(bson.D), bson.E{Key: "first_name", Value: updatedUser.FirstName})
	updateFields[0].Value = append(updateFields[0].Value.(bson.D), bson.E{Key: "last_name", Value: updatedUser.LastName})
	updateFields[0].Value = append(updateFields[0].Value.(bson.D), bson.E{Key: "about_user", Value: updatedUser.AboutUser})
	updateFields[0].Value = append(updateFields[0].Value.(bson.D), bson.E{Key: "city", Value: updatedUser.City})
	updateFields[0].Value = append(updateFields[0].Value.(bson.D), bson.E{Key: "photo_url", Value: updatedUser.PhotoURL})
	updateFields[0].Value = append(updateFields[0].Value.(bson.D), bson.E{Key: "updated_at", Value: time.Now()})

	options := options.FindOneAndUpdate().SetReturnDocument(options.After)

	result := userCollection.FindOneAndUpdate(ctx, filter, updateFields, options)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("error updating user: %v", result.Err())
	}

	var updated models.User
	if err := result.Decode(&updated); err != nil {
		return fmt.Errorf("error decoding updated user: %v", err)
	}

	return nil
}

func UpdateProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем ID пользователя из JWT
		userID, _ := c.Get("uid")
		userIDString := fmt.Sprintf("%v", userID)

		// Извлекаем данные для обновления из тела запроса
		var updateData models.User
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		// Вызовите функцию обновления gameCard
		err := updateProfile(context.Background(), userIDString, updateData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error updating user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"msg": "user updated successfully"})
	}
}

func GetProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем ID пользователя из JWT
		userID, _ := c.Get("uid")

		// Преобразуем userID в строку
		userIDString := fmt.Sprintf("%v", userID)

		fmt.Print("hello")

		// Формируем фильтр для поиска пользователя по ID
		filter := bson.M{"user_id": userIDString}

		// Выполняем запрос к базе данных
		var user models.User
		err := userCollection.FindOne(context.Background(), filter).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user data"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": user})
	}
}

func GetUserData() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Получаем user_id из параметров запроса
		userID := c.Param("user_id")

		// Получаем информацию о пользователе из коллекции users
		userFilter := bson.M{"user_id": userID}
		var user models.User
		err := userCollection.FindOne(ctx, userFilter).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user data"})
			return
		}

		// Получаем список игровых карт, где пользователь является хостом
		hostFilter := bson.M{"host_user.user_id": userID}
		hostCursor, err := gameCardCollection.Find(ctx, hostFilter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch game cards"})
			return
		}
		var hostGameCards []models.GameCard
		err = hostCursor.All(ctx, &hostGameCards)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode game cards"})
			return
		}

		// Получаем список игровых карт, где пользователь состоит в списке matched_players
		matchFilter := bson.M{"matched_players": bson.M{"$elemMatch": bson.M{"user_id": userID}}}
		matchCursor, err := gameCardCollection.Find(ctx, matchFilter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch game cards"})
			return
		}
		var matchGameCards []models.GameCard
		err = matchCursor.All(ctx, &matchGameCards)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode game cards"})
			return
		}

		// Собираем информацию о пользователе и его игровых картах в один ответ
		userData := gin.H{
			"user": gin.H{
				"first_name": user.FirstName,
				"last_name":  user.LastName,
				"about_me":   user.AboutUser,
				"photo_url":  user.PhotoURL,
				"city":       user.City,
			},
			"games": append(hostGameCards, matchGameCards...),
		}

		// Возвращаем ответ
		c.JSON(http.StatusOK, userData)
	}
}
