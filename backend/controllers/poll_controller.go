package controllers

import (
	"net/http"

	"live-polling-backend/models"
	"live-polling-backend/services"
	"live-polling-backend/websocket"

	"github.com/gin-gonic/gin"
)

type PollController struct {
	pollService *services.PollService
	hub         *websocket.Hub
}

func NewPollController(pollService *services.PollService, hub *websocket.Hub) *PollController {
	return &PollController{
		pollService: pollService,
		hub:         hub,
	}
}

func (ctrl *PollController) CreatePoll(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.CreatePollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	poll, err := ctrl.pollService.CreatePoll(c.Request.Context(), userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Poll created successfully",
		"poll":    poll,
	})
}

func (ctrl *PollController) GetUserPolls(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	polls, err := ctrl.pollService.GetUserPolls(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"polls": polls})
}

func (ctrl *PollController) GetPollByID(c *gin.Context) {
	pollID := c.Param("id")
	poll, err := ctrl.pollService.GetPollByID(c.Request.Context(), pollID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"poll": poll})
}

func (ctrl *PollController) GetPollByShareCode(c *gin.Context) {
	shareCode := c.Param("shareCode")
	poll, err := ctrl.pollService.GetPollByShareCode(c.Request.Context(), shareCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"poll": poll})
}

func (ctrl *PollController) Vote(c *gin.Context) {
	pollID := c.Param("id")

	var req models.VoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Capture fingerprint from request body or header
	fingerprint := req.VoterFingerprint
	if fingerprint == "" {
		fingerprint = c.GetHeader("X-Voter-Fingerprint")
	}
	req.VoterFingerprint = fingerprint

	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	update, err := ctrl.pollService.Vote(c.Request.Context(), pollID, &req, ip, ua)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Vote recorded successfully",
		"update":  update,
	})
}

func (ctrl *PollController) GetResults(c *gin.Context) {
	pollID := c.Param("id")
	poll, err := ctrl.pollService.GetPollByID(c.Request.Context(), pollID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"poll": poll})
}

func (ctrl *PollController) ClosePoll(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	pollID := c.Param("id")
	err := ctrl.pollService.ClosePoll(c.Request.Context(), pollID, userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Poll closed successfully"})
}

func (ctrl *PollController) DeletePoll(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	pollID := c.Param("id")
	err := ctrl.pollService.DeletePoll(c.Request.Context(), pollID, userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Poll deleted successfully"})
}

// HandleWebSocket upgrades HTTP connection and registers client to poll room
func (ctrl *PollController) HandleWebSocket(c *gin.Context) {
	pollID := c.Param("id")
	if pollID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Poll ID required"})
		return
	}

	// Fetch current initial snapshot to send immediately to the newly connected browser
	poll, err := ctrl.pollService.GetPollByID(c.Request.Context(), pollID)
	var initialSnapshot interface{}
	if err == nil && poll != nil {
		initialSnapshot = ctrl.pollService.BuildInitialSnapshot(c.Request.Context(), poll)
	}

	websocket.ServeWs(ctrl.hub, pollID, c.Writer, c.Request, initialSnapshot)
}
