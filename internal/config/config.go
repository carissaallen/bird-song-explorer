package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	Environment        string
	DatabaseURL        string
	YotoClientID       string
	YotoClientSecret   string
	YotoAccessToken    string
	YotoRefreshToken   string
	YotoCardID         string
	YotoAPIBaseURL     string
	EBirdAPIKey        string
	XenoCantoAPIKey    string
	ElevenLabsAPIKey   string
	OpenAIAPIKey       string
	SchedulerToken     string
	CacheTTLHours      int
	BirdOfDayResetHour int
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Port:               getEnv("PORT", "8080"),
		Environment:        getEnv("ENV", "development"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		YotoClientID:       getEnv("YOTO_CLIENT_ID", ""),
		YotoClientSecret:   getEnv("YOTO_CLIENT_SECRET", ""),
		YotoAccessToken:    getEnv("YOTO_ACCESS_TOKEN", ""),
		YotoRefreshToken:   getEnv("YOTO_REFRESH_TOKEN", ""),
		YotoCardID:         getEnv("YOTO_CARD_ID", ""),
		YotoAPIBaseURL:     getEnv("YOTO_API_BASE_URL", "https://api.yotoplay.com"),
		EBirdAPIKey:        getEnv("EBIRD_API_KEY", ""),
		XenoCantoAPIKey:    getEnv("XENOCANTO_API_KEY", ""),
		ElevenLabsAPIKey:   getEnv("ELEVENLABS_API_KEY", ""),
		OpenAIAPIKey:       getEnv("OPENAI_API_KEY", ""),
		SchedulerToken:     getEnv("SCHEDULER_TOKEN", ""),
		CacheTTLHours:      24,
		BirdOfDayResetHour: 6,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
