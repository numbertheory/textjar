package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Paste represents a saved snippet of text
type Paste struct {
	gorm.Model
	Content string `gorm:"type:text"`
}

var db *gorm.DB

func initDB() {
	dbPath := "db/textjar.db"

	// Ensure the db directory exists
	err := os.MkdirAll(filepath.Dir(dbPath), 0755)
	if err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	// Initialize SQLite database
	var openErr error
	db, openErr = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if openErr != nil {
		log.Fatal("Failed to connect to database:", openErr)
	}

	// Automatically migrate the schema
	db.AutoMigrate(&Paste{})
}

func main() {
	// Initialize the database
	initDB()

	r := gin.Default()

	// Set trusted proxies to nil to resolve the security warning.
	r.SetTrustedProxies(nil)

	// Serve static files
	r.Static("/static", "./static")

	// Routes
	r.GET("/", func(c *gin.Context) {
		render(c, "index.html", gin.H{
			"Title": "New Paste",
		})
	})

	r.POST("/save", func(c *gin.Context) {
		content := c.PostForm("content")
		if content == "" {
			c.Redirect(http.StatusSeeOther, "/")
			return
		}

		paste := Paste{Content: content}
		result := db.Create(&paste)
		if result.Error != nil {
			c.String(http.StatusInternalServerError, "Error saving paste")
			return
		}

		c.Redirect(http.StatusSeeOther, "/view/"+strconv.FormatUint(uint64(paste.ID), 10))
	})

	r.GET("/recent", func(c *gin.Context) {
		var pastes []Paste
		db.Order("created_at desc").Limit(20).Find(&pastes)

		render(c, "recent.html", gin.H{
			"Title":  "Recent Saves",
			"Pastes": pastes,
		})
	})

	r.GET("/view/:id", func(c *gin.Context) {
		id := c.Param("id")
		var paste Paste
		if err := db.First(&paste, id).Error; err != nil {
			c.String(http.StatusNotFound, "Paste not found")
			return
		}

		render(c, "view.html", gin.H{
			"Title": "View Paste",
			"Paste": paste,
		})
	})

	log.Println("Server starting on :8080")
	r.Run(":8080")
}

func render(c *gin.Context, name string, data gin.H) {
	tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/"+name))
	err := tmpl.Execute(c.Writer, data)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}
