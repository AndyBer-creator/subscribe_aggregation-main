package api

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"subscribe_aggregation-main/pkg/logging"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// DeleteSubscription godoc
// @Summary Delete subscription by ID
// @Tags subscriptions
// @Param id path string true "Subscription ID UUID"
// @Success 204 "No content"
// @Failure 400 {string} string "Invalid UUID"
// @Failure 404 {string} string "Subscription not found"
// @Failure 500 {string} string "Server error"
// @Router /subscriptions/{id} [delete]
func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Error("DeleteSubscription: invalid UUID", slog.String("uuid", idStr), slog.String("error", err.Error()))
		http.Error(w, "invalid UUID", http.StatusBadRequest)
		return
	}

	err = h.Storage.DeleteSubscription(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info("DeleteSubscription: subscription not found", slog.String("subscription_id", id.String()))
			http.Error(w, "subscription not found", http.StatusNotFound)
			return
		}
		logger.Error("DeleteSubscription: failed to delete subscription", slog.String("subscription_id", id.String()), slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("DeleteSubscription: subscription deleted", slog.String("subscription_id", id.String()))
	w.WriteHeader(http.StatusNoContent)
}
