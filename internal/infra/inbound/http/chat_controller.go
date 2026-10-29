package http

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/application/port/inbound"
	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/repository"
)

type ChatController struct {
	getMessageHistory inbound.GetMessageHistoryUseCase
}

func NewChatController(getMessageHistory inbound.GetMessageHistoryUseCase) *ChatController {
	return &ChatController{getMessageHistory: getMessageHistory}
}

func (c *ChatController) Routes(router *gin.Engine) {
	router.GET("/api/v1/chats/messages", c.GetMessageHistory)
}

func (c *ChatController) GetMessageHistory(ctx *gin.Context) {
	taskID := ctx.Query("taskId")
	if taskID == "" {
		ctx.JSON(400, gin.H{"error": "taskId is required"})
		return
	}

	messages, err := c.getMessageHistory.GetMessageHistory(ctx.Request.Context(), taskID)
	if err != nil {
		status := 500
		if errors.Is(err, repository.ErrChatNotFound) {
			status = 404
		}
		ctx.JSON(status, gin.H{"error": "request failed"})
		return
	}

	ctx.JSON(200, messages)
}
