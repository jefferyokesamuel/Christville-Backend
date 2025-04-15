package main

import (
	"christville/controllers"
	"christville/db"
	"christville/middleware"
	"christville/routes"
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func fetchAndLogBibleVerse() {
	log.Println("Starting the cron job to fetch daily Bible verse...")
	bibleVerse, err := controllers.FetchRandomBibleVerse()
	if err != nil {
		log.Printf("Error fetching daily Bible verse: %v\n", err)
		return
	}
	log.Printf("Successfully fetched and saved Bible verse: %s - %s\n", bibleVerse.Reference, bibleVerse.Text)
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database connection
	_ = db.GetClient()

	// Connect to MongoDB
	client, err := db.ConnectMongoDB()
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v\n", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting MongoDB: %v\n", err)
		}
	}()

	// Define allowed hosts
	allowedHosts := []string{"localhost:8080", "vivablockchainconsulting.xyz", "3.83.246.115", "3.83.246.115:8080", "172.31.95.36"}

	// Initialize Gin router
	router := gin.Default()

	// Apply the AllowedHostsMiddleware
	router.Use(middleware.AllowedHostsMiddleware(allowedHosts))

	// Setup routes
	routes.SetupRoutes(router)

	// Initialize the cron job scheduler
	c := cron.New(cron.WithSeconds())
	_, err = c.AddFunc("@daily", fetchAndLogBibleVerse) // Runs every day at midnight
	if err != nil {
		log.Fatalf("Failed to schedule cron job: %v\n", err)
	}
	c.Start()
	defer c.Stop()

	// Start the HTTP server in a goroutine
	go func() {
		log.Println("Christville server started on port 8080...")
		if err := http.ListenAndServe(":8080", router); err != nil {
			log.Fatalf("Error starting server: %v\n", err)
		}
	}()

	// Test the cron job immediately for debugging purposes
	fetchAndLogBibleVerse()

	// Block forever to keep server and cron running
	select {}
}
