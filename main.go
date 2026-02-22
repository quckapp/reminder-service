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

	// Get database reference for new services
	db := repo.Database()

	// Initialize Kafka producer
	producer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to connect to Kafka: %v", err)
	}
	defer producer.Close()

	// ── Initialize Core Service ──
	reminderService := service.NewReminderService(repo, producer)

	// ── Initialize Extended Services ──
	tagService := service.NewTagService(db)
	templateService := service.NewTemplateService(db)
	notificationService := service.NewNotificationService(db)
	sharingService := service.NewSharingService(db)
	noteService := service.NewNoteService(db)
	activityService := service.NewActivityService(db)
	analyticsService := service.NewAnalyticsService(repo, db)
	searchService := service.NewSearchService(db)
	exportService := service.NewExportService(repo, db)
	priorityService := service.NewPriorityService(db)
	subtaskService := service.NewSubtaskService(db)
	calendarService := service.NewCalendarService(db)
	escalationService := service.NewEscalationService(db)
	categoryService := service.NewCategoryService(db)
	delegationService := service.NewDelegationService(db)
	recurringService := service.NewRecurringService(db)
	timezoneService := service.NewTimezoneService(db)
	habitService := service.NewHabitService(db)
	extended2Service := service.NewExtended2Service(db)

	// ── Initialize Scheduler ──
	reminderScheduler := scheduler.NewScheduler(reminderService, producer)
	go reminderScheduler.Start()

	// ── Initialize Kafka Consumer ──
	consumer, err := kafka.NewConsumer(cfg.KafkaBrokers, "reminder-service", reminderService)
	if err != nil {
		log.Printf("Warning: Failed to connect Kafka consumer: %v", err)
	} else {
		go consumer.Start()
		defer consumer.Close()
	}

	// ── Initialize Handlers ──
	tagHandler := api.NewTagHandler(tagService)
	templateHandler := api.NewTemplateHandler(templateService, reminderService)
	notifHandler := api.NewNotificationHandler(notificationService)
	sharingHandler := api.NewSharingHandler(sharingService)
	noteHandler := api.NewNoteHandler(noteService)
	analyticsHandler := api.NewAnalyticsHandler(analyticsService, searchService, activityService, exportService, reminderService)
	priorityHandler := api.NewPriorityHandler(priorityService)
	subtaskHandler := api.NewSubtaskHandler(subtaskService)
	calendarHandler := api.NewCalendarHandler(calendarService)
	escalationHandler := api.NewEscalationHandler(escalationService)
	categoryHandler := api.NewCategoryHandler(categoryService)
	delegationHandler := api.NewDelegationHandler(delegationService)
	recurringHandler := api.NewRecurringHandler(recurringService)
	timezoneHandler := api.NewTimezoneHandler(timezoneService)
	habitHandler := api.NewHabitHandler(habitService)
	ext2Handler := api.NewExtended2Handler(extended2Service)

	// ── Setup HTTP Server ──
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	api.RegisterRoutes(
		router,
		reminderService,
		tagHandler,
		templateHandler,
		notifHandler,
		sharingHandler,
		noteHandler,
		analyticsHandler,
		priorityHandler,
		subtaskHandler,
		calendarHandler,
		escalationHandler,
		categoryHandler,
		delegationHandler,
		recurringHandler,
		timezoneHandler,
		habitHandler,
		ext2Handler,
	)

	port := cfg.Port
	log.Printf("Reminder service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
