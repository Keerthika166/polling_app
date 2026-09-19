package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"live-polling-backend/config"
	"live-polling-backend/controllers"
	"live-polling-backend/middleware"
	"live-polling-backend/models"
	"live-polling-backend/repository"
	"live-polling-backend/routes"
	"live-polling-backend/services"
	"live-polling-backend/websocket"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	log.Println("==================================================")
	log.Println("   PulsePoll — Real-Time Live Polling Engine")
	log.Println("   GUVI Developer Internship Project")
	log.Println("==================================================")

	cfg := config.LoadConfig()

	// 1. Connect MongoDB
	dbInstance, err := config.ConnectMongoDB(cfg)
	if err != nil {
		log.Printf("[MongoDB] ERROR: Could not connect to MongoDB at %s: %v", cfg.MongoURI, err)
		log.Println("[MongoDB] Please ensure MongoDB is running or run via `docker compose up`")
	}

	// 2. Connect Redis
	redisClient, err := config.ConnectRedis(cfg)
	if err != nil {
		log.Printf("[Redis] ERROR: Could not connect to Redis at %s: %v", cfg.RedisURL, err)
		log.Println("[Redis] Please ensure Redis is running or run via `docker compose up`")
	}

	// 3. Initialize Repositories (with automatic in-memory fallback if db is offline)
	var db *mongo.Database
	if dbInstance != nil {
		db = dbInstance.DB
		log.Println("[MongoDB] SUCCESS: Live MongoDB connected.")
	} else {
		log.Println("[MongoDB] NOTICE: Running with in-memory persistence fallback.")
	}

	if redisClient != nil {
		log.Printf("[Redis] SUCCESS: Live Redis connected at %s (Hash tallies & Pub/Sub active).", cfg.RedisURL)
	} else {
		log.Println("[Redis] NOTICE: Running with in-memory pub/sub & counters fallback.")
	}

	userRepo := repository.NewUserRepository(db)
	pollRepo := repository.NewPollRepository(db)
	voteRepo := repository.NewVoteRepository(db)

	// 4. Initialize Services
	redisService := services.NewRedisService(redisClient)
	authService := services.NewAuthService(userRepo, cfg)
	pollService := services.NewPollService(pollRepo, voteRepo, redisService)

	// Ensure default demo organizer exists for 1-click login
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = authService.Signup(ctx, &models.SignupRequest{
			Name:     "Demo Organizer",
			Email:    "organizer@example.com",
			Password: "secret123",
		})
	}()

	// 5. Initialize & start WebSocket Hub
	hub := websocket.NewHub(redisService)
	go hub.Run()

	// 6. Initialize Controllers
	authCtrl := controllers.NewAuthController(authService)
	pollCtrl := controllers.NewPollController(pollService, hub)

	// 7. Setup Gin router
	router := gin.Default()
	router.Use(middleware.CORSMiddleware(cfg.FrontendURL))

	// Setup all API and WebSocket routes
	routes.SetupRoutes(router, authCtrl, pollCtrl, authService, redisService, dbInstance)

	// 8. HTTP Server with graceful shutdown
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	go func() {
		log.Printf("[PulsePoll Server] Listening on http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[PulsePoll Server] Listen error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[PulsePoll Server] Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[PulsePoll Server] Server forced to shutdown: %v", err)
	}

	if dbInstance != nil && dbInstance.Client != nil {
		_ = dbInstance.Client.Disconnect(ctx)
	}
	if redisClient != nil {
		_ = redisClient.Close()
	}

	log.Println("[PulsePoll Server] Server exiting cleanly.")
}
