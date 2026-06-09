package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/yuin/goldmark"
)

func main() {
	r := gin.Default()

	// Set trusted proxies to nil to resolve the security warning.
	// In production, you would set this to the specific IP addresses of your proxies.
	r.SetTrustedProxies(nil)

	// Serve static files
	r.Static("/static", "./static")

	// Load HTML templates
	r.LoadHTMLGlob("templates/*")

	// Routes
	r.GET("/", markdownHandler("content/index.md", "Home"))
	r.GET("/about", markdownHandler("content/about.md", "About"))

	log.Println("Server starting on :8080")
	r.Run(":8080")
}

func markdownHandler(filePath string, title string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read Markdown file
		content, err := os.ReadFile(filePath)
		if err != nil {
			c.String(http.StatusNotFound, "File not found")
			return
		}

		// Convert Markdown to HTML using Goldmark
		var buf bytes.Buffer
		if err := goldmark.Convert(content, &buf); err != nil {
			c.String(http.StatusInternalServerError, "Error parsing markdown")
			return
		}

		// Render template with parsed HTML
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title":   title,
			"Content": template.HTML(buf.String()),
		})
	}
}
