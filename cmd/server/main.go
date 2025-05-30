package main

import (
	"log"
	"os"

	"order-matching-system/internal/api"
	"order-matching-system/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Loading configuration from system environment variables")
	}

	dbConfig := database.Config{
		Host:     getRequiredEnv("DB_HOST"),
		Port:     getRequiredEnv("DB_PORT"),
		User:     getRequiredEnv("DB_USER"),
		Password: getRequiredEnv("DB_PASSWORD"),
		Database: getRequiredEnv("DB_NAME"),
	}

	serverPort := getRequiredEnv("SERVER_PORT")

	log.Println("Connecting to database...")
	if err := database.Initialize(dbConfig); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()

	log.Println("Database connected successfully")

	router := api.SetupRouter(database.DB)

	log.Printf("Starting server on port %s...", serverPort)
	if err := router.Run(":" + serverPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func getRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s environment variable is required", key)
	}
	return value
}
