package main

import (
	"dataShare/db"
	"dataShare/document"
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
	"strconv"
)

type Template struct {
	//templates *template.Template
	templates map[string]*template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	//return t.templates.ExecuteTemplate(w, name, data)

	tmpl, ok := t.templates[name]
	if !ok {
		err := errors.New("Template not found -> " + name)
		return err
	}
	return tmpl.ExecuteTemplate(w, "base.html", data)
}

func upload(c echo.Context) error {
	// Read form fields
	name := c.FormValue("name")
	email := c.FormValue("email")

	//------------
	// Read files
	//------------

	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	files := form.File["files"]

	for _, file := range files {
		// Source
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		// Destination
		dst, err := os.Create(file.Filename)
		if err != nil {
			return err
		}
		defer dst.Close()

		// Copy
		if _, err = io.Copy(dst, src); err != nil {
			return err
		}

	}

	return c.HTML(http.StatusOK, fmt.Sprintf("<p>Uploaded successfully %d files with fields name=%s and email=%s.</p>", len(files), name, email))
}

func indexHandler(c echo.Context) error {

	return c.Render(http.StatusOK, "home.html", map[string]interface{}{
		"name": "About",
		"msg":  "All about Boatswain!",
		"csrf": c.Get("csrf"),
	})

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

	e := echo.New()

	templates := make(map[string]*template.Template)
	templates["home.html"] = template.Must(template.ParseFiles("view/home.html", "view/base.html"))
	e.Renderer = &Template{
		templates: templates,
	}
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:csrf",
	}))

	e.Static("/static", "static")

	e.GET("/", indexHandler)
	e.POST("/upload", upload)

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

func dbMigrate(db *gorm.DB) {
	err := db.AutoMigrate(&document.Document{})
	if err != nil {
		log.Fatal(err)
	}
}
