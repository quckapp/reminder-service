package config

import "os"

type Config struct {
	Port         string
	MongoDBURL   string
	DatabaseName string
	KafkaBrokers []string
	RedisURL     string
}

func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "4010"),
		MongoDBURL:   getEnv("MONGODB_URL", "mongodb://localhost:27017"),
		DatabaseName: getEnv("DATABASE_NAME", "quickchat_reminders"),
		KafkaBrokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		RedisURL:     getEnv("REDIS_URL", "localhost:6379"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
