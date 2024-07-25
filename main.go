package main

import (
	"flag"
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
	host := flag.String("host", "0.0.0.0", "Service host")
	port := flag.String("port", "1323", "Service port")
	flag.Parse()

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("error loading .env file, %v", err)
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	db, err := gorm.Open(postgres.Open(fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		dbHost,
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(&models.Nasabah{}, &models.Transaksi{}, &models.Counter{}, &models.Saldo{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	client := handlers.NewClient(db)

	e := echo.New()
	e.Validator = utils.NewValidator()
	e.Use(middleware.Recover())

	routes.RegisterRoutes(e, client)

	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", *host, *port)))
}
