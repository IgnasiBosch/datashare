package main

import (
	"dataShare/db"
	"dataShare/document"
	"fmt"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
	"io"
	"net/http"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"log"
	"os"
)

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
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/", "public")
	e.POST("/upload", upload)

	e.Logger.Fatal(e.Start("localhost:" + appPort))

	//e.GET("/", func(c echo.Context) error {
	//	return c.String(http.StatusOK, "Hello, World!")
	//})

	//// TODO: This middleware is for debug purpose only. Revisit this later.
	//e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
	//	Format: "method=${method}, uri=${uri}, status=${status}\n",
	//}))
	//e.Logger.Fatal(e.Start("localhost:" + appPort))

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
