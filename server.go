package main

import (
	"dataShare/core"
	"dataShare/db"
	"dataShare/document"
	"dataShare/service"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type Template struct {
	templates map[string]*template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		err := errors.New("Template not found -> " + name)
		return err
	}
	return tmpl.ExecuteTemplate(w, "base.html", data)
}

// NewDocumentHandler is a helper function that handles the creation of the document handler with required dependencies.
func NewDocumentHandler(c echo.Context) (*document.Handler, error) {
	DB, ok := c.Get("db").(*gorm.DB)
	if !ok {
		return nil, errors.New("failed to get DB from context")
	}
	encryption, ok := c.Get("encryption").(*service.Encryption)
	if !ok {
		return nil, errors.New("failed to get encryption service from context")
	}

	return document.NewHandler(c, DB, encryption), nil
}

func indexHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "home.html", map[string]interface{}{
		"csrf": c.Get("csrf"),
	})
}

func uploadDocument(c echo.Context) error {
	h, err := NewDocumentHandler(c)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "error.html", map[string]interface{}{
			"errorMsg": "Something went wrong",
		})
	}
	form, err := c.MultipartForm()
	if err != nil {
		return c.Render(http.StatusInternalServerError, "error.html", map[string]interface{}{
			"errorMsg": "Something went wrong",
		})
	}

	idKey, err := h.Encrypt(form)
	if err != nil {
		return c.Render(http.StatusUnprocessableEntity, "error.html", map[string]interface{}{
			"errorMsg": err.Error(),
		})
	}

	return c.Render(http.StatusOK, "upload_response.html", map[string]interface{}{
		"link": os.Getenv("BASE_URL") + "/" + idKey.ID,
		"key":  idKey.Key,
	})
}

func checkDocument(c echo.Context) error {
	h, err := NewDocumentHandler(c)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "error.html", map[string]interface{}{
			"errorMsg": "Something went wrong",
		})
	}
	ID := c.Param("id")
	err = h.Check(ID)
	if err != nil {
		return c.Render(http.StatusUnprocessableEntity, "error.html", map[string]interface{}{
			"errorMsg": err.Error(),
		})
	}

	return c.Render(http.StatusOK, "get_document.html", map[string]interface{}{
		"csrf": c.Get("csrf"),
		"id":   ID,
	})
}

func downloadDocument(c echo.Context) error {
	h, err := NewDocumentHandler(c)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "error.html", map[string]interface{}{
			"errorMsg": "Something went wrong",
		})
	}
	ID := c.Param("id")
	key := c.FormValue("key")
	content, d, err := h.Decrypt(core.NewIDKey(ID, key))
	if err != nil {
		return c.Render(http.StatusUnprocessableEntity, "error.html", map[string]interface{}{
			"errorMsg": err.Error(),
		})
	}

	c.Response().Header().Set("Content-Disposition", "attachment; filename="+d.Filename)
	return c.Blob(http.StatusOK, d.FileContentType, content)
}

func ContextEncryption(e *service.Encryption) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("encryption", e)
			return next(c)
		}
	}
}

func initCleaningTask(stopChan chan os.Signal, db *gorm.DB) {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-stopChan:
			ticker.Stop()
			fmt.Println("Cleaning task stopped.")
			return
		case t := <-ticker.C:
			fmt.Println("Cleaning task triggered at: ", t)
			document.CleanUp(db)
		}
	}
}

// main is the entry point of the application.
// It initializes the environment variables, establishes a database connection, performs database migration,
// creates a new Echo instance, defines a route to handle the root path, and starts the server.
// It also uses a middleware for logging purposes.
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	appPort := os.Getenv("APP_PORT")
	dbConn := getDatabaseConnection()
	dbMigrate(dbConn)

	// interrupt signal handling
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	go initCleaningTask(stopChan, dbConn)
	go func() {
		<-stopChan
		// Print a message and exit the application
		fmt.Println("Caught stop signal. Cleaning up and terminating the application...")
		os.Exit(0)
	}()

	e := echo.New()

	e.IPExtractor = echo.ExtractIPDirect()
	e.Use(db.ContextDB(dbConn))
	e.Use(ContextEncryption(getEncryption()))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:_csrf",
	}))

	templates := make(map[string]*template.Template)
	templates["home.html"] = template.Must(template.ParseFiles("view/home.html", "view/base.html"))
	templates["upload_response.html"] = template.Must(template.ParseFiles("view/upload_response.html", "view/base.html"))
	templates["get_document.html"] = template.Must(template.ParseFiles("view/get_document.html", "view/base.html"))
	templates["error.html"] = template.Must(template.ParseFiles("view/error.html", "view/base.html"))
	e.Renderer = &Template{
		templates: templates,
	}

	e.Static("/static", "static")

	e.GET("/", indexHandler)
	e.POST("/", uploadDocument)
	e.GET("/:id", checkDocument)
	e.POST("/:id", downloadDocument)

	e.Logger.Fatal(e.Start("localhost:" + appPort))
}

// getDatabaseConnection returns the database connection object.
func getDatabaseConnection() *gorm.DB {
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatalf("Failed to convert DB_PORT to int")
	}

	dbConn, err := db.DatabaseConnection(dbHost, dbName, dbUser, dbPassword, dbPort)
	if err != nil {
		log.Fatalf("Failed to connect to database")
	}
	return dbConn
}

func getEncryption() *service.Encryption {
	iterations, err := strconv.Atoi(os.Getenv("ENCRYPTION_ITERATIONS"))
	if err != nil {
		log.Fatalf("Failed to convert ENCRYPTION_ITERATIONS to int")
	}

	bockSize, err := strconv.Atoi(os.Getenv("ENCRYPTION_BLOCK_SIZE_LENGTH"))
	if err != nil {
		log.Fatalf("Failed to convert ENCRYPTION_BLOCK_SIZE_LENGTH to int")
	}

	saltLength, err := strconv.Atoi(os.Getenv("ENCRYPTION_SALT_LENGTH"))
	if err != nil {
		log.Fatalf("Failed to convert ENCRYPTION_SALT_LENGTH to int")
	}

	return service.NewEncryption(
		iterations,
		bockSize,
		saltLength,
		os.Getenv("ENCRYPTION_HASH_SALT"),
	)
}

func dbMigrate(db *gorm.DB) {
	err := db.AutoMigrate(&document.Document{})
	if err != nil {
		log.Fatal(err)
	}
}
