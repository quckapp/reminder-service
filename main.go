package main

import (
	"log"
	"os"

	"reminder-service/internal/api"
	"reminder-service/internal/config"
	"reminder-service/internal/kafka"
	"reminder-service/internal/repository"
	"reminder-service/internal/scheduler"
	"reminder-service/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// Initialize MongoDB repository
	repo, err := repository.NewMongoRepository(cfg.MongoDBURL, cfg.DatabaseName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer repo.Close()

	// Initialize Kafka producer
	producer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to connect to Kafka: %v", err)
	}
	defer producer.Close()

	// Initialize service
	reminderService := service.NewReminderService(repo, producer)

	// Initialize scheduler
	reminderScheduler := scheduler.NewScheduler(reminderService, producer)
	go reminderScheduler.Start()

	// Initialize Kafka consumer
	consumer, err := kafka.NewConsumer(cfg.KafkaBrokers, "reminder-service", reminderService)
	if err != nil {
		log.Printf("Warning: Failed to connect Kafka consumer: %v", err)
	} else {
		go consumer.Start()
		defer consumer.Close()
	}

	// Setup HTTP server
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	api.RegisterRoutes(router, reminderService)

	port := cfg.Port
	log.Printf("Reminder service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
