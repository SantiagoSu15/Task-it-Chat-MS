package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/application/port/inbound"
	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/repository"
)

type ChatController struct {
	getMessageHistory inbound.GetMessageHistoryUseCase
}

func NewChatController(getMessageHistory inbound.GetMessageHistoryUseCase) *ChatController {
	return &ChatController{getMessageHistory: getMessageHistory}
}

func (c *ChatController) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/api/v1/chats/messages", c.GetMessageHistory)
	return r
}

func (c *ChatController) GetMessageHistory(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("taskId")
	if taskID == "" {
		http.Error(w, "taskId is required", http.StatusBadRequest)
		return
	}

	messages, err := c.getMessageHistory.GetMessageHistory(r.Context(), taskID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, repository.ErrChatNotFound) {
			status = http.StatusNotFound
		}
		http.Error(w, http.StatusText(status), status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(messages); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
