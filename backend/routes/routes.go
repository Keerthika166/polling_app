package routes

import (
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
) {
	api := router.Group("/api")

	// Health check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "PulsePoll API",
			"version": "1.0.0",
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
