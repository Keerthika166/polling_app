package routes

import (
	"time"

	"live-polling-backend/config"
	"live-polling-backend/controllers"
	"live-polling-backend/middleware"
	"live-polling-backend/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	router *gin.Engine,
	authCtrl *controllers.AuthController,
	pollCtrl *controllers.PollController,
	authService *services.AuthService,
	redisService *services.RedisService,
	dbInstance *config.Database,
) {
	api := router.Group("/api")

	// Comprehensive Health check verifying both MongoDB and Redis connectivity
	api.GET("/health", func(c *gin.Context) {
		redisConnected := false
		redisLatency := "N/A"
		if redisService != nil && redisService.HasLiveClient() {
			start := time.Now()
			pong, err := redisService.Ping(c.Request.Context())
			if err == nil && pong == "PONG" {
				redisConnected = true
				redisLatency = time.Since(start).String()
			}
		}

		mongoConnected := false
		mongoLatency := "N/A"
		if dbInstance != nil {
			start := time.Now()
			if err := dbInstance.Ping(c.Request.Context()); err == nil {
				mongoConnected = true
				mongoLatency = time.Since(start).String()
			}
		}

		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "PulsePoll API",
			"version": "1.0.0",
			"databases": gin.H{
				"redis": gin.H{
					"connected": redisConnected,
					"status":    func() string { if redisConnected { return "connected" }; return "offline" }(),
					"ping":      redisLatency,
					"role":      "Atomic HINCRBY Vote Counts & Pub/Sub Live WebSocket Broadcast",
				},
				"mongodb": gin.H{
					"connected": mongoConnected,
					"status":    func() string { if mongoConnected { return "connected" }; return "in-memory fallback" }(),
					"ping":      mongoLatency,
					"role":      "Persistent Storage (Users, Polls, Audit Votes)",
				},
			},
		})
	})

	// 1. Authentication routes (Public)
	auth := api.Group("/auth")
	{
		auth.POST("/signup", authCtrl.Signup)
		auth.POST("/login", authCtrl.Login)
		// Protected user me
		auth.GET("/me", middleware.AuthMiddleware(authService), authCtrl.GetMe)
	}

	// 2. Poll routes
	polls := api.Group("/polls")
	{
		// Public lookup and voting
		polls.GET("/share/:shareCode", pollCtrl.GetPollByShareCode)
		polls.GET("/:id", pollCtrl.GetPollByID)
		polls.POST("/:id/vote", pollCtrl.Vote)
		polls.GET("/:id/results", pollCtrl.GetResults)

		// Real-time WebSocket feed
		polls.GET("/:id/live", pollCtrl.HandleWebSocket)

		// Protected poll creation and management
		protected := polls.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			protected.POST("", pollCtrl.CreatePoll)
			protected.GET("", pollCtrl.GetUserPolls)
			protected.PATCH("/:id/close", pollCtrl.ClosePoll)
			protected.DELETE("/:id", pollCtrl.DeletePoll)
		}
	}
}
