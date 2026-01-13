package scheduler

import (
	"context"
	"log"
	"time"

	"reminder-service/internal/kafka"
	"reminder-service/internal/service"
)

type Scheduler struct {
	service  *service.ReminderService
	producer *kafka.Producer
	ticker   *time.Ticker
	done     chan bool
}

func NewScheduler(svc *service.ReminderService, producer *kafka.Producer) *Scheduler {
	return &Scheduler{
		service:  svc,
		producer: producer,
		done:     make(chan bool),
	}
}

func (s *Scheduler) Start() {
	s.ticker = time.NewTicker(30 * time.Second) // Check every 30 seconds
	log.Println("Reminder scheduler started")

	for {
		select {
		case <-s.ticker.C:
			s.checkPendingReminders()
		case <-s.done:
			s.ticker.Stop()
			return
		}
	}
}

func (s *Scheduler) Stop() {
	s.done <- true
}

func (s *Scheduler) checkPendingReminders() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reminders, err := s.service.GetPendingReminders(ctx, time.Now())
	if err != nil {
		log.Printf("Error fetching pending reminders: %v", err)
		return
	}

	for _, reminder := range reminders {
		if err := s.service.TriggerReminder(ctx, reminder); err != nil {
			log.Printf("Error triggering reminder %s: %v", reminder.ID.Hex(), err)
			continue
		}
		log.Printf("Triggered reminder: %s for user %s", reminder.ID.Hex(), reminder.UserID)
	}
}
