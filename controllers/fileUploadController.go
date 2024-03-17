package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func UploadPhoto() gin.HandlerFunc {
	return func(c *gin.Context) {

		file, err := c.FormFile("photo")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error retrieving the file"})
			return
		}

		maxFileSize := 20 << 20 // 20 MB в байтах
		if file.Size > int64(maxFileSize) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds the maximum allowed size (20 MB)"})
			return
		}

		// Проверяем расширение файла
		allowedExtensions := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true}
		ext := filepath.Ext(file.Filename)
		if !allowedExtensions[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type"})
			return
		}

		// Сохраняем файл
		if err := c.SaveUploadedFile(file, "./uploads/"+file.Filename); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving the file"})
			return
		}

		// Формируем URL для загруженного файла
		fileURL := fmt.Sprintf("http://%s/uploads/%s", c.Request.Host, file.Filename)

		// Возвращаем успешный ответ с URL загруженного файла
		c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "photo_url": fileURL})
	}
}
