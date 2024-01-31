package helpers

import (
	"context"
	"fmt"
	"log"
	"oiynlike/database"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var userCollection *mongo.Collection = database.OpenCollection(database.ConnectToMongoDB(), "users")

var SECRET_KEY []byte = []byte(os.Getenv("SECRET_KEY"))

// GenerateAllTokens генерирует токен и refresh token
func GenerateAllTokens(email, firstName, lastName, userType, uid string) (string, string, error) {
	// Создание токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":     email,
		"firstName": firstName,
		"lastName":  lastName,
		"userType":  userType,
		"uid":       uid,
		"exp":       time.Now().Add(time.Hour * 1).Unix(), // Токен действителен в течение 1 часа
	})

	// Подпись токена
	tokenString, err := token.SignedString(SECRET_KEY)
	if err != nil {
		return "", "", err
	}

	// Создание refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": uid,
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(), // Refresh token действителен в течение 7 дней
	})

	// Подпись refresh token
	refreshTokenString, err := refreshToken.SignedString(SECRET_KEY)
	if err != nil {
		return "", "", err
	}

	return tokenString, refreshTokenString, nil
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var updateObj primitive.D

	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj, bson.E{Key: "refreshToken", Value: signedRefreshToken})

	Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: Updated_at})

	upsert := true
	filter := bson.M{"user_id": userId}

	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{Key: "$set", Value: updateObj},
		},
		&opt,
	)

	if err != nil {
		log.Panic(err)
		return
	}
	return

}

func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return SECRET_KEY, nil
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to parse token: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("Invalid token")
	}

	// Проверка времени истечения (exp)
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("Failed to extract claims from token")
	}

	expirationTime := time.Unix(int64(claims["exp"].(float64)), 0)
	if time.Now().After(expirationTime) {
		return nil, fmt.Errorf("Token has expired")
	}

	return claims, nil
}
