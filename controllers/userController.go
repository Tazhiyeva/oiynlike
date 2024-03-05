package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"oiynlike/database"
	helper "oiynlike/helpers"

	"oiynlike/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
			msg := fmt.Sprint("User hasn't created ")
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

func UpdateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем ID пользователя из JWT
		userID, _ := c.Get("uid")

		// Извлекаем данные для обновления из тела запроса
		var updateData models.User
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Преобразуем userID в строку
		userIDString := fmt.Sprintf("%v", userID)

		// Формируем фильтр для поиска пользователя по ID
		filter := bson.M{"user_id": userIDString}

		// Формируем обновленные данные
		updateFields := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "first_name", Value: updateData.FirstName},
				{Key: "last_name", Value: updateData.LastName},
				{Key: "email", Value: updateData.Email},
				{Key: "about_user", Value: updateData.AboutUser},
				{Key: "photo_url", Value: updateData.PhotoURL},
				{Key: "city", Value: updateData.City},
				{Key: "updated_at", Value: time.Now()},
			}},
		}

		// Опции для FindOneAndUpdate
		options := options.FindOneAndUpdate().SetReturnDocument(options.After)

		// Выполняем обновление в базе данных
		result := userCollection.FindOneAndUpdate(context.Background(), filter, updateFields, options)
		if result.Err() != nil {
			if result.Err() == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user"})
			return
		}

		// Декодируем обновленные данные
		var updatedUser models.User
		if err := result.Decode(&updatedUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding updated user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"msg": "User updated successfully", "user": updatedUser})
	}
}

// Admin data
func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
		groupStage := bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "null"},
				{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "data", Value: bson.D{
					{Key: "$push", Value: "$$ROOT"},
				}},
			}},
		}

		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{
					{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}},
				}},
			}},
		}

		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})

		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing users"})
			return
		}

		var allusers []bson.M

		if err = result.All(ctx, &allusers); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allusers)

	}

}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")

		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error:": err.Error()})
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User

		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user)
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
