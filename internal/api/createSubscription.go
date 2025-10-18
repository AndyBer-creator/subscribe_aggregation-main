package api

import (
	"encoding/json"
	"net/http"
	"subscribe_aggregation-main/internal/models"

	"github.com/google/uuid"
)

// CreateSubscription godoc
// @Summary      Create a new subscription
// @Description  Создает новую подписку с уникальным UUID
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        subscription  body      models.Subscription  true  "Subscription data"
// @Success      201  {object}  models.Subscription
// @Failure      400  {string}  string "Invalid request payload"
// @Failure      500  {string}  string "Internal server error"
// @Router       /subscriptions [post]
func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var sub models.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Генерация ID
	sub.ID = uuid.New()

	// Валидация обязательных полей
	if sub.ServiceName == "" || sub.Price <= 0 || sub.UserID == uuid.Nil {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	// Сохранение подписки
	if err := h.Storage.CreateSubscription(r.Context(), &sub); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sub)
}
