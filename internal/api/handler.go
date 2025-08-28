package api

import (
	"encoding/json"
	"net/http"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"subscribe_aggregation-main/internal/models"
	"subscribe_aggregation-main/internal/storage"
	"subscribe_aggregation-main/pkg/logging"
)

// Handler инкапсулирует слой хранения и предоставляет методы для обработки HTTP-запросов
type Handler struct {
	storage *storage.Storage
}

// NewHandler создаёт новый экземпляр Handler с указанным хранилищем
func NewHandler(storage *storage.Storage) *Handler {
	return &Handler{storage: storage}
}

<<<<<<< HEAD
// LoggingMiddleware - middleware для логирования HTTP запросов и ответов.
=======
// LoggingMiddleware - middleware для логирования HTTP запросов и ответов
>>>>>>> 78bf63b (updated)
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logging.GetLogger()
		start := time.Now()

<<<<<<< HEAD
		// Оборачиваем ResponseWriter для мониторинга status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: 200}

		// Логируем начало запроса
=======
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: 200}
>>>>>>> 78bf63b (updated)
		logger.Info("Request started",
			slog.String("method", r.Method),
			slog.String("url", r.URL.String()),
			slog.String("remote_addr", r.RemoteAddr),
		)

<<<<<<< HEAD
		// Передаём вызов следующему обработчику
		next.ServeHTTP(lrw, r)

		// Вычисляем длительность обработки запроса
		duration := time.Since(start).Milliseconds()

		// Логируем завершение запроса с результатом
=======
		next.ServeHTTP(lrw, r)

		duration := time.Since(start).Milliseconds()

>>>>>>> 78bf63b (updated)
		logger.Info("Request completed",
			slog.String("method", r.Method),
			slog.String("url", r.URL.String()),
			slog.String("remote_addr", r.RemoteAddr),
			slog.Int("status", lrw.statusCode),
			slog.Int64("duration_ms", duration),
		)
	})
}

<<<<<<< HEAD
// loggingResponseWriter - обёртка http.ResponseWriter для отслеживания HTTP статуса
=======
>>>>>>> 78bf63b (updated)
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

<<<<<<< HEAD
// WriteHeader перехватывает установку status code и сохраняет его
=======
>>>>>>> 78bf63b (updated)
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

<<<<<<< HEAD
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

	// Парсим JSON тело запроса
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		logger.Error("CreateSubscription: invalid request payload",
			slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Генерируем уникальный UUID для подписки
	sub.ID = uuid.New()

	// Создаём подписку в базе
	if err := h.storage.CreateSubscription(r.Context(), &sub); err != nil {
		logger.Error("CreateSubscription: failed to create subscription",
			slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info("CreateSubscription: subscription created",
		slog.String("subscription_id", sub.ID.String()))

	// Отправляем ответ с созданным ресурсом
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sub)
}

// ListSubscriptions godoc
// @Summary      List all subscriptions
// @Tags         subscriptions
// @Produce      json
// @Success      200  {array}   models.Subscription
// @Failure      500  {string}  string "Internal server error"
// @Router       /subscriptions [get]
=======
// Создать подписку
func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()

	var sub models.Subscription

	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		logger.Error("CreateSubscription: invalid request payload", slog.String("error", err.Error()))
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	sub.ID = uuid.New()
	if sub.StartDate.IsZero() {
		sub.StartDate = time.Now()
	}

	if err := h.storage.CreateSubscription(r.Context(), &sub); err != nil {
		logger.Error("CreateSubscription: failed to create subscription", slog.String("error", err.Error()))
		http.Error(w, "failed to create subscription", http.StatusInternalServerError)
		return
	}

	logger.Info("CreateSubscription: subscription created", slog.String("subscription_id", sub.ID.String()))

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(sub); err != nil {
		logger.Error("CreateSubscription: failed to encode response", slog.String("error", err.Error()))
	}
}

// Получить список подписок
>>>>>>> 78bf63b (updated)
func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	subs, err := h.storage.ListSubscriptions(r.Context())
	if err != nil {
<<<<<<< HEAD
		logger.Error("ListSubscriptions: failed to list subscriptions",
			slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info("ListSubscriptions: retrieved subscriptions",
		slog.Int("count", len(subs)))

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
=======
		logger.Error("ListSubscriptions: failed to list subscriptions", slog.String("error", err.Error()))
		http.Error(w, "failed to list subscriptions", http.StatusInternalServerError)
		return
	}

	logger.Info("ListSubscriptions: retrieved subscriptions", slog.Int("count", len(subs)))

	if err := json.NewEncoder(w).Encode(subs); err != nil {
		logger.Error("ListSubscriptions: failed to encode response", slog.String("error", err.Error()))
	}
}

// Получить подписку по ID
>>>>>>> 78bf63b (updated)
func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	idStr := chi.URLParam(r, "id")

<<<<<<< HEAD
	// Парсим UUID из параметра URL
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Error("GetSubscription: invalid UUID",
			slog.String("uuid", idStr), slog.String("error", err.Error()))
=======
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Error("GetSubscription: invalid UUID", slog.String("uuid", idStr), slog.String("error", err.Error()))
>>>>>>> 78bf63b (updated)
		http.Error(w, "invalid UUID", http.StatusBadRequest)
		return
	}

<<<<<<< HEAD
	// Получаем подписку из хранилища
	sub, err := h.storage.GetSubscriptionByID(r.Context(), id)
	if err != nil || sub == nil {
		logger.Error("GetSubscription: subscription not found",
			slog.String("subscription_id", id.String()))
=======
	sub, err := h.storage.GetSubscriptionByID(r.Context(), id)
	if err != nil {
		logger.Error("GetSubscription: internal error", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if sub == nil {
		logger.Error("GetSubscription: subscription not found", slog.String("subscription_id", id.String()))
>>>>>>> 78bf63b (updated)
		http.Error(w, "subscription not found", http.StatusNotFound)
		return
	}

<<<<<<< HEAD
	logger.Info("GetSubscription: subscription retrieved",
		slog.String("subscription_id", id.String()))

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
=======
	logger.Info("GetSubscription: subscription retrieved", slog.String("subscription_id", id.String()))

	if err := json.NewEncoder(w).Encode(sub); err != nil {
		logger.Error("GetSubscription: failed to encode response", slog.String("error", err.Error()))
	}
}

// Обновить подписку по ID
>>>>>>> 78bf63b (updated)
func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	idStr := chi.URLParam(r, "id")

<<<<<<< HEAD
	// Парсим UUID из URL
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Error("UpdateSubscription: invalid UUID",
			slog.String("uuid", idStr), slog.String("error", err.Error()))
=======
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Error("UpdateSubscription: invalid UUID", slog.String("uuid", idStr), slog.String("error", err.Error()))
>>>>>>> 78bf63b (updated)
		http.Error(w, "invalid UUID", http.StatusBadRequest)
		return
	}

<<<<<<< HEAD
	// Парсим тело запроса
	var sub models.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		logger.Error("UpdateSubscription: invalid request body",
			slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
=======
	var sub models.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		logger.Error("UpdateSubscription: invalid request body", slog.String("error", err.Error()))
		http.Error(w, "invalid request body", http.StatusBadRequest)
>>>>>>> 78bf63b (updated)
		return
	}

	sub.ID = id
<<<<<<< HEAD

	// Обновляем подписку в базе
	if err := h.storage.UpdateSubscription(r.Context(), &sub); err != nil {
		logger.Error("UpdateSubscription: failed to update subscription",
			slog.String("subscription_id", id.String()), slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info("UpdateSubscription: subscription updated",
		slog.String("subscription_id", id.String()))

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
=======
	if err := h.storage.UpdateSubscription(r.Context(), &sub); err != nil {
		logger.Error("UpdateSubscription: failed to update subscription", slog.String("subscription_id", id.String()), slog.String("error", err.Error()))
		http.Error(w, "failed to update subscription", http.StatusInternalServerError)
		return
	}

	logger.Info("UpdateSubscription: subscription updated", slog.String("subscription_id", id.String()))

	if err := json.NewEncoder(w).Encode(sub); err != nil {
		logger.Error("UpdateSubscription: failed to encode response", slog.String("error", err.Error()))
	}
}

// Удалить подписку по ID
>>>>>>> 78bf63b (updated)
func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	idStr := chi.URLParam(r, "id")

<<<<<<< HEAD
	// Парсим UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Error("DeleteSubscription: invalid UUID",
			slog.String("uuid", idStr), slog.String("error", err.Error()))
=======
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Error("DeleteSubscription: invalid UUID", slog.String("uuid", idStr), slog.String("error", err.Error()))
>>>>>>> 78bf63b (updated)
		http.Error(w, "invalid UUID", http.StatusBadRequest)
		return
	}

<<<<<<< HEAD
	// Удаляем подписку
	if err := h.storage.DeleteSubscription(r.Context(), id); err != nil {
		logger.Error("DeleteSubscription: failed to delete subscription",
			slog.String("subscription_id", id.String()), slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info("DeleteSubscription: subscription deleted",
		slog.String("subscription_id", id.String()))
=======
	if err := h.storage.DeleteSubscription(r.Context(), id); err != nil {
		logger.Error("DeleteSubscription: failed to delete subscription", slog.String("subscription_id", id.String()), slog.String("error", err.Error()))
		http.Error(w, "failed to delete subscription", http.StatusInternalServerError)
		return
	}

	logger.Info("DeleteSubscription: subscription deleted", slog.String("subscription_id", id.String()))
>>>>>>> 78bf63b (updated)

	w.WriteHeader(http.StatusNoContent)
}

<<<<<<< HEAD
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
=======
// SumSubscriptionsCostHandler - обработчик подсчёта суммы стоимости подписок с фильтрами
>>>>>>> 78bf63b (updated)
func (h *Handler) SumSubscriptionsCostHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()

	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")
	startStr := r.URL.Query().Get("start_date")
<<<<<<< HEAD
	endStr := r.URL.Query().Get("end_date")

	var start, end time.Time
	var err error

	// Парсим дату начала периода
	if startStr != "" {
		start, err = time.Parse("01-2006", startStr)
		if err != nil {
			logger.Error("SumSubscriptionsCostHandler: invalid start_date format",
				slog.String("error", err.Error()))
=======

	var start time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse("01-2006", startStr)
		if err != nil {
			logger.Error("SumSubscriptionsCostHandler: invalid start_date format", slog.String("error", err.Error()))
>>>>>>> 78bf63b (updated)
			http.Error(w, "invalid start_date format, expected MM-YYYY", http.StatusBadRequest)
			return
		}
	} else {
		start = time.Time{}
	}

<<<<<<< HEAD
	// Парсим дату окончания периода
	if endStr != "" {
		end, err = time.Parse("01-2006", endStr)
		if err != nil {
			logger.Error("SumSubscriptionsCostHandler: invalid end_date format",
				slog.String("error", err.Error()))
			http.Error(w, "invalid end_date format, expected MM-YYYY", http.StatusBadRequest)
			return
		}
	} else {
		end = time.Now()
	}

	// Считаем сумму по подпискам с учётом фильтров
	total, err := h.storage.SumSubscriptionsCost(r.Context(), userID, serviceName, start, end)
	if err != nil {
		logger.Error("SumSubscriptionsCostHandler: failed to sum subscriptions cost",
			slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
=======
	total, err := h.storage.SumSubscriptionsCost(r.Context(), userID, serviceName, start)
	if err != nil {
		logger.Error("SumSubscriptionsCostHandler: failed to sum subscriptions cost", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
>>>>>>> 78bf63b (updated)
		return
	}

	logger.Info("SumSubscriptionsCostHandler: total price calculated",
		slog.String("user_id", userID), slog.String("service_name", serviceName), slog.Int64("total_price", total))

	json.NewEncoder(w).Encode(map[string]int64{"total_price": total})
}
