package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"subscribe_aggregation-main/pkg/logging"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// GetSubscription godoc
// @Summary      Get subscription by ID
// @Description  Получить подписку по её уникальному UUID
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Subscription ID (UUID)"
// @Success      200  {object}  models.Subscription
// @Failure      400  {string}  string "Invalid UUID"
// @Failure      404  {string}  string "Subscription not found"
// @Failure      500  {string}  string "Internal server error"
// @Router       /subscriptions/{id} [get]
func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Error("GetSubscription: invalid UUID", slog.String("uuid", idStr), slog.String("error", err.Error()))
		http.Error(w, "invalid UUID", http.StatusBadRequest)
		return
	}

	sub, err := h.storage.GetSubscriptionByID(r.Context(), id)
	if err != nil {
		logger.Error("GetSubscription: internal error", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if sub == nil {
		logger.Info("GetSubscription: subscription not found", slog.String("subscription_id", id.String()))
		http.Error(w, "subscription not found", http.StatusNotFound)
		return
	}

	logger.Info("GetSubscription: subscription retrieved", slog.String("subscription_id", id.String()))
	json.NewEncoder(w).Encode(sub)
}
