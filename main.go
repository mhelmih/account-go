package main

import (
	"fmt"
	"log"
	"os"

	"account/handlers"
	"account/models"
	"account/routes"
	"account/utils"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("error loading .env file, %v", err)
	}

	db, err := gorm.Open(postgres.Open(fmt.Sprintf(
		"host=localhost user=%s password=%s dbname=%s port=5432 sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(&models.Nasabah{}, &models.Transaksi{}, &models.Counter{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	client := handlers.NewClient(db)

	e := echo.New()
	e.Validator = utils.NewValidator()
	e.Use(middleware.Recover())

	routes.RegisterRoutes(e, client)

	e.Logger.Fatal(e.Start(":1323"))
}
