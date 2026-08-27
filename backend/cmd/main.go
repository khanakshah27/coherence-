package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"coherence.dev/backend/internal/api"
	"coherence.dev/backend/internal/config"
	"coherence.dev/backend/internal/database"
	"coherence.dev/backend/internal/drift"
	"coherence.dev/backend/internal/providers"
)

func init() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("No .env file found: %v", err)
	}
}

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg := config.LoadConfig()
	logger.WithField("config", cfg).Info("Configuration loaded")

	// Initialize database
	db, err := database.InitDB(cfg.DatabaseURL)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize database")
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		logger.WithError(err).Fatal("Failed to run migrations")
	}

	// Initialize cloud providers
	awsProvider := providers.NewAWSProvider(cfg.AWSConfig)
	gcpProvider := providers.NewGCPProvider(cfg.GCPConfig)
	azureProvider := providers.NewAzureProvider(cfg.AzureConfig)

	// Initialize drift detection service
	driftService := drift.NewService(
		db,
		awsProvider,
		gcpProvider,
		azureProvider,
		logger,
	)

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Setup API routes
	api.SetupRoutes(
		router,
		db,
		driftService,
		logger,
	)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"version": "1.0.0",
		})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.WithField("port", port).Info("Starting Coherence server")
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		logger.WithError(err).Fatal("Failed to start server")
	}
}
