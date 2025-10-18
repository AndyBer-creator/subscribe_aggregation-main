package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"subscribe_aggregation-main/pkg/logging"
)

// ListSubscriptions godoc
// @Summary      List all subscriptions
// @Tags         subscriptions
// @Produce      json
// @Success      200  {array}   models.Subscription
// @Failure      500  {string}  string "Internal server error"
// @Router       /subscriptions [get]
func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	query := r.URL.Query()

	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}

	subs, err := h.Storage.ListSubscriptions(r.Context(), page, limit)
	if err != nil {
		logger.Error("ListSubscriptions: failed to list subscriptions", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("ListSubscriptions: retrieved subscriptions", slog.Int("count", len(subs)))
	json.NewEncoder(w).Encode(subs)
}
