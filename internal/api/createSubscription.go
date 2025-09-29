package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"subscribe_aggregation-main/internal/models"
	"subscribe_aggregation-main/pkg/logging"

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
	logger := logging.GetLogger()
	var sub models.Subscription

	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		//logger.Error("CreateSubscription: invalid request payload", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sub.ID = uuid.New()

	if err := h.storage.CreateSubscription(r.Context(), &sub); err != nil {
		logger.Error("CreateSubscription: failed to create subscription", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info("CreateSubscription: subscription created", slog.String("subscription_id", sub.ID.String()))
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sub)
}
