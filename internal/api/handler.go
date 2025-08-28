package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"subscribe_aggregation-main/internal/models"
	"subscribe_aggregation-main/internal/storage"
)

type Handler struct {
	storage *storage.Storage
}

func NewHandler(storage *storage.Storage) *Handler {
	return &Handler{storage: storage}
}

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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sub.ID = uuid.New()

	if err := h.storage.CreateSubscription(r.Context(), &sub); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sub)
}

func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	subs, err := h.storage.ListSubscriptions(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(subs)
}

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
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid UUID", http.StatusBadRequest)
		return
	}

	sub, err := h.storage.GetSubscriptionByID(r.Context(), id)
	if err != nil || sub == nil {
		http.Error(w, "subscription not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(sub)
}

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
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid UUID", http.StatusBadRequest)
		return
	}

	var sub models.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sub.ID = id

	if err := h.storage.UpdateSubscription(r.Context(), &sub); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(sub)
}

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
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid UUID", http.StatusBadRequest)
		return
	}

	if err := h.storage.DeleteSubscription(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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
	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")
	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse("01-2006", startStr)
		if err != nil {
			http.Error(w, "invalid start_date format, expected MM-YYYY", http.StatusBadRequest)
			return
		}
	} else {
		start = time.Time{}
	}

	if endStr != "" {
		end, err = time.Parse("01-2006", endStr)
		if err != nil {
			http.Error(w, "invalid end_date format, expected MM-YYYY", http.StatusBadRequest)
			return
		}
	} else {
		end = time.Now()
	}

	total, err := h.storage.SumSubscriptionsCost(r.Context(), userID, serviceName, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int64{"total_price": total})
}
