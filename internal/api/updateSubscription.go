package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"subscribe_aggregation-main/internal/models"
	"subscribe_aggregation-main/pkg/logging"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// UpdateSubscription godoc
// @Summary      Update subscription by ID
// @Description  Update subscription record by UUID with new data
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id    path      string             true  "Subscription ID (UUID)"
// @Param        sub   body      models.Subscription true  "Subscription object"
// @Success      200   {object}  models.Subscription
// @Failure      400   {string}  string "Invalid input or UUID"
// @Failure      404   {string}  string "Subscription not found"
// @Failure      500   {string}  string "Internal server error"
// @Router       /subscriptions/{id} [put]

func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Error("UpdateSubscription: invalid UUID", slog.String("uuid", idStr), slog.String("error", err.Error()))
		http.Error(w, "invalid UUID", http.StatusBadRequest)
		return
	}

	var sub models.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		logger.Error("UpdateSubscription: invalid request body", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sub.ID = id

	err = h.Storage.UpdateSubscription(r.Context(), &sub)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info("UpdateSubscription: subscription not found", slog.String("subscription_id", id.String()))
			http.Error(w, "subscription not found", http.StatusNotFound)
			return
		}
		logger.Error("UpdateSubscription: failed to update subscription", slog.String("subscription_id", id.String()), slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info("UpdateSubscription: subscription updated", slog.String("subscription_id", id.String()))
	json.NewEncoder(w).Encode(sub)
}
