package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all v1 API routes
func RegisterRoutes(rg *gin.RouterGroup) {
	api := rg.Group("/v1/api")
	{
		api.GET("/health", healthCheck)
	}
}

// healthCheck returns a 200 OK status
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
