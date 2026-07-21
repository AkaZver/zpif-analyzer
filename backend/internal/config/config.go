package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	OpenAIAPIKey  string
	OpenAIBaseURL string
	OpenAIModel   string

	WebSearchProvider string
	WebSearchAPIKey   string

	JWTSecret string

	ServerPort string
}

func Load() *Config {
	cfg := &Config{
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnvInt("DB_PORT", 5432),
		DBUser:            getEnv("DB_USER", "zpif"),
		DBPassword:        getEnv("DB_PASSWORD", "zpif"),
		DBName:            getEnv("DB_NAME", "zpif_analyzer"),
		DBSSLMode:         getEnv("DB_SSL_MODE", "disable"),
		OpenAIAPIKey:      getEnv("OPENAI_API_KEY", ""),
		OpenAIBaseURL:     getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIModel:       getEnv("OPENAI_MODEL", "gpt-4o-mini"),
		WebSearchProvider: getEnv("WEBSEARCH_PROVIDER", "serpapi"),
		WebSearchAPIKey:   getEnv("WEBSEARCH_API_KEY", ""),
		JWTSecret:         getEnv("JWT_SECRET", "change-me-in-production"),
		ServerPort:        getEnv("SERVER_PORT", "8080"),
	}
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, ok := os.LookupEnv(key); ok {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
