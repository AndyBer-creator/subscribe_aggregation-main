package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"subscribe_aggregation-main/pkg/logging"
	"time"
)

// SumSubscriptionsCostHandler godoc
// @Summary Calculate total subscription cost filtered by user, service and period
// @Tags subscription
// @Accept json
// @Produce json
// @Param user_id query string false "User ID UUID"
// @Param service_name query string false "Service Name"
// @Param start_date query string false "Start month-year MM-YYYY"
// @Param end_date query string false "End month-year MM-YYYY"
// @Success 200 {object} map[string]int64
// @Failure 400 {string} string "Invalid parameter"
// @Failure 500 {string} string "Server error"
// @Router /subscriptions/sum [get]
func (h *Handler) SumSubscriptionsCostHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()

	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")
	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse("01-2006", startStr)
		if err != nil {
			logger.Error("SumSubscriptionsCostHandler: invalid start_date format", slog.String("error", err.Error()))
			http.Error(w, "invalid start_date format, expected MM-YYYY", http.StatusBadRequest)
			return
		}
	} else {
		start = time.Time{}
	}

	if endStr != "" {
		end, err = time.Parse("01-2006", endStr)
		if err != nil {
			logger.Error("SumSubscriptionsCostHandler: invalid end_date format", slog.String("error", err.Error()))
			http.Error(w, "invalid end_date format, expected MM-YYYY", http.StatusBadRequest)
			return
		}
	} else {
		end = time.Now()
	}

	total, err := h.Storage.SumSubscriptionsCost(r.Context(), userID, serviceName, start, end)
	if err != nil {
		logger.Error("SumSubscriptionsCostHandler: failed to sum subscriptions cost", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info("SumSubscriptionsCostHandler: total price calculated",
		slog.String("user_id", userID), slog.String("service_name", serviceName), slog.Int64("total_price", total))

	json.NewEncoder(w).Encode(map[string]int64{"total_price": total})
}
