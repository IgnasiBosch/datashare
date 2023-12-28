package main

import (
	"github.com/labstack/echo/v4/middleware"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	appPort := os.Getenv("APP_PORT")

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// TODO: This middleware is for debug purpose only. Revisit this later.
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.Logger.Fatal(e.Start("localhost:" + appPort))

}
