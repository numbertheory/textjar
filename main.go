package main

import (
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"math/rand"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Paste represents a group of versions identified by a slug
type Paste struct {
	gorm.Model
	Slug     string         `gorm:"uniqueIndex"`
	Versions []PasteVersion `gorm:"foreignKey:PasteID;constraint:OnDelete:CASCADE;"`
}

// PasteVersion represents a specific version of a paste's content
type PasteVersion struct {
	gorm.Model
	PasteID uint
	Content string `gorm:"type:text"`
	Number  int    // Version number (1, 2, 3...)
}

var db *gorm.DB
var wordList []string

func initDB() {
	dbPath := "db/textjar.db"
	photosPath := "photos"

	// Ensure the db directory exists
	err := os.MkdirAll(filepath.Dir(dbPath), 0755)
	if err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	// Ensure the photos directory exists
	err = os.MkdirAll(photosPath, 0755)
	if err != nil {
		log.Fatal("Failed to create photos directory:", err)
	}

	// Initialize SQLite database
	var openErr error
	db, openErr = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if openErr != nil {
		log.Fatal("Failed to connect to database:", openErr)
	}

	// Automatically migrate the schema
	db.AutoMigrate(&Paste{}, &PasteVersion{})

	// Load words for slug generation
	loadWords()
}

func loadWords() {
	content, err := os.ReadFile("templates/words.txt")
	if err != nil {
		log.Println("Warning: templates/words.txt not found, falling back to simple slugs")
		wordList = []string{"paste", "note", "text"}
		return
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		word := strings.ToLower(strings.TrimSpace(line))
		if word != "" {
			wordList = append(wordList, word)
		}
	}
}

func generateUniqueSlug() string {
	for {
		var parts []string
		for i := 0; i < 3; i++ {
			parts = append(parts, wordList[rand.Intn(len(wordList))])
		}
		slug := strings.Join(parts, "-")

		var count int64
		db.Model(&Paste{}).Where("slug = ?", slug).Count(&count)
		if count == 0 {
			return slug
		}
	}
}

func main() {
	// Initialize the database
	initDB()

	r := gin.Default()

	// Set trusted proxies to nil to resolve the security warning.
	r.SetTrustedProxies(nil)

	// Serve static files
	r.Static("/static", "./static")
	r.Static("/photos-data", "./photos")

	// Routes
	r.GET("/", func(c *gin.Context) {
		render(c, "index.html", gin.H{
			"Title": "Text",
		})
	})

	r.GET("/photos", func(c *gin.Context) {
		files, _ := os.ReadDir("photos")
		var photoNames []string
		for _, file := range files {
			if !file.IsDir() {
				photoNames = append(photoNames, file.Name())
			}
		}

		// Sort photos by name (timestamp prefix) in descending order to show newest first
		sort.Slice(photoNames, func(i, j int) bool {
			return photoNames[i] > photoNames[j]
		})

		// Pagination logic
		pageSize := 25
		pageStr := c.DefaultQuery("page", "1")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		totalPhotos := len(photoNames)
		totalPages := (totalPhotos + pageSize - 1) / pageSize

		start := (page - 1) * pageSize
		end := start + pageSize

		if start > totalPhotos {
			start = totalPhotos
		}
		if end > totalPhotos {
			end = totalPhotos
		}

		paginatedPhotos := photoNames[start:end]

		render(c, "photos.html", gin.H{
			"Title":       "Photos",
			"Photos":      paginatedPhotos,
			"CurrentPage": page,
			"TotalPages":  totalPages,
			"NextPage":    page + 1,
			"PrevPage":    page - 1,
			"HasNext":     page < totalPages,
			"HasPrev":     page > 1,
		})
	})

	r.POST("/photos/upload", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.String(http.StatusBadRequest, "Get form err: %s", err.Error())
			return
		}
		files := form.File["photos"]

		for _, file := range files {
			filename := filepath.Base(file.Filename)
			// Add timestamp to prevent name collisions
			newFilename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filename)
			if err := c.SaveUploadedFile(file, filepath.Join("photos", newFilename)); err != nil {
				c.String(http.StatusBadRequest, "Upload file err: %s", err.Error())
				return
			}
		}

		c.Redirect(http.StatusSeeOther, "/photos")
	})

	r.POST("/save", func(c *gin.Context) {
		content := c.PostForm("content")
		slug := c.PostForm("slug") // For edits

		if content == "" {
			c.Redirect(http.StatusSeeOther, "/")
			return
		}

		if slug != "" {
			// Update existing paste (New Version)
			var paste Paste
			if err := db.Where("slug = ?", slug).First(&paste).Error; err != nil {
				c.String(http.StatusNotFound, "Paste not found")
				return
			}

			var lastVersion PasteVersion
			db.Where("paste_id = ?", paste.ID).Order("number desc").First(&lastVersion)

			newVersion := PasteVersion{
				PasteID: paste.ID,
				Content: content,
				Number:  lastVersion.Number + 1,
			}
			db.Create(&newVersion)

			// Update the parent Paste's UpdatedAt timestamp using time.Now()
			db.Model(&paste).Update("updated_at", time.Now())

			c.Redirect(http.StatusSeeOther, "/view/"+slug)
		} else {
			// New paste
			slug = generateUniqueSlug()
			paste := Paste{Slug: slug}
			db.Create(&paste)

			version := PasteVersion{
				PasteID: paste.ID,
				Content: content,
				Number:  1,
			}
			db.Create(&version)
			c.Redirect(http.StatusSeeOther, "/view/"+slug)
		}
	})

	r.GET("/recent", func(c *gin.Context) {
		const pageSize = 25
		pageStr := c.DefaultQuery("page", "1")
		var page int
		fmt.Sscanf(pageStr, "%d", &page)
		if page < 1 {
			page = 1
		}

		var pastes []Paste
		var totalCount int64
		db.Model(&Paste{}).Count(&totalCount)

		offset := (page - 1) * pageSize
		// Join with versions to get the latest content for preview
		db.Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("paste_versions.number DESC")
		}).Order("pastes.updated_at DESC").Offset(offset).Limit(pageSize).Find(&pastes)

		render(c, "recent.html", gin.H{
			"Title":       "Recent Saves",
			"Pastes":      pastes,
			"CurrentPage": page,
			"HasMore":     int64(offset+pageSize) < totalCount,
		})
	})

	r.GET("/view/:slug", func(c *gin.Context) {
		slug := c.Param("slug")
		versionStr := c.Query("v")

		var paste Paste
		if err := db.Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("number desc")
		}).Where("slug = ?", slug).First(&paste).Error; err != nil {
			c.String(http.StatusNotFound, "Paste not found")
			return
		}

		var selectedVersion PasteVersion
		if versionStr != "" {
			vNum, _ := strconv.Atoi(versionStr)
			for _, v := range paste.Versions {
				if v.Number == vNum {
					selectedVersion = v
					break
				}
			}
		}

		// Default to latest if not found or not specified
		if selectedVersion.ID == 0 && len(paste.Versions) > 0 {
			selectedVersion = paste.Versions[0]
		}

		render(c, "view.html", gin.H{
			"Title":           "View Paste",
			"Paste":           paste,
			"SelectedVersion": selectedVersion,
			"Content":         template.HTML(selectedVersion.Content),
		})
	})

	r.GET("/edit/:slug", func(c *gin.Context) {
		slug := c.Param("slug")
		var paste Paste
		if err := db.Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("number desc")
		}).Where("slug = ?", slug).First(&paste).Error; err != nil {
			c.String(http.StatusNotFound, "Paste not found")
			return
		}

		render(c, "index.html", gin.H{
			"Title": "Edit Paste: " + slug,
			"Paste": paste,
			// Load latest version content
			"InitialContent": template.HTML(paste.Versions[0].Content),
		})
	})

	r.POST("/delete/:slug", func(c *gin.Context) {
		slug := c.Param("slug")
		// GORM will handle deletion of versions due to OnDelete:CASCADE
		result := db.Where("slug = ?", slug).Delete(&Paste{})
		if result.Error != nil {
			c.String(http.StatusInternalServerError, "Error deleting paste")
			return
		}
		c.Redirect(http.StatusSeeOther, "/recent")
	})

	log.Println("Server starting on :3636")
	r.Run(":3636")
}

func render(c *gin.Context, name string, data gin.H) {
	// Parse the layout and the requested template
	tmpl := template.Must(template.New("base.html").Funcs(template.FuncMap{
		"stripHTML": stripHTML,
		"add": func(a, b int) int {
			return a + b
		},
	}).ParseFiles("templates/base.html", "templates/"+name))

	// Execute the base template, which will use the content block from the requested template
	err := tmpl.Execute(c.Writer, data)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func stripHTML(s string) string {
	// First strip HTML tags
	re := regexp.MustCompile("<[^>]*>")
	plainText := re.ReplaceAllString(s, " ")

	// Decode HTML entities (e.g., &nbsp; -> space, &amp; -> &)
	decoded := html.UnescapeString(plainText)

	// Replace non-breaking spaces with normal spaces explicitly if UnescapeString leaves them as \u00a0
	decoded = strings.ReplaceAll(decoded, "\u00a0", " ")

	return strings.TrimSpace(decoded)
}
