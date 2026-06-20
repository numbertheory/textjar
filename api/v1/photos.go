package v1

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func uploadPhotos(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Get form err: %s", err.Error())})
		return
	}
	files := form.File["photos"]

	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No photos found in request"})
		return
	}

	var uploadedFiles []string
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		// Add timestamp to prevent name collisions
		newFilename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filename)
		if err := c.SaveUploadedFile(file, filepath.Join("photos", newFilename)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Upload file err: %s", err.Error())})
			return
		}
		uploadedFiles = append(uploadedFiles, newFilename)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Files uploaded successfully",
		"files":   uploadedFiles,
	})
}
