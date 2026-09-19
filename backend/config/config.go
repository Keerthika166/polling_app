package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	MongoURI    string
	DBName      string
	RedisURL    string
	JWTSecret   string
	FrontendURL string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "pulsepoll_db"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "super-secret-pulsepoll-jwt-key-2026"
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	return &Config{
		Port:        port,
		MongoURI:    mongoURI,
		DBName:      dbName,
		RedisURL:    redisURL,
		JWTSecret:   jwtSecret,
		FrontendURL: frontendURL,
	}
}
