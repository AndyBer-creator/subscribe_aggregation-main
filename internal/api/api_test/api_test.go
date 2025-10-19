package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"subscribe_aggregation-main/internal/api"
	"subscribe_aggregation-main/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockStorage - мок для тестирования хранилища подписок.
type MockStorage struct {
	mock.Mock
}

// CreateSubscription - мок метода создания подписки.
func (m *MockStorage) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

// GetSubscriptionByID - мок метода получения подписки по ID.
func (m *MockStorage) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Subscription), args.Error(1)
}

// ListSubscriptions - мок метода получения списка подписок.
func (m *MockStorage) ListSubscriptions(ctx context.Context, page, limit int) ([]models.Subscription, error) {
	args := m.Called(ctx, page, limit)
	return args.Get(0).([]models.Subscription), args.Error(1)
}

// UpdateSubscription - мок метода обновления подписки.
func (m *MockStorage) UpdateSubscription(ctx context.Context, sub *models.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

// DeleteSubscription - мок метода удаления подписки.
func (m *MockStorage) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// SumSubscriptionsCost - мок метода подсчёта стоимости подписок.
func (m *MockStorage) SumSubscriptionsCost(ctx context.Context, userID, serviceName string, filterStart, filterEnd time.Time) (int64, error) {
	args := m.Called(ctx, userID, serviceName, filterStart, filterEnd)
	return args.Get(0).(int64), args.Error(1)
}

// TestCreateSubscription - тестирование эндпоинта создания подписки.
func TestCreateSubscription(t *testing.T) {
	mockStore := new(MockStorage)
	handler := &api.Handler{
		Storage: mockStore,
	}

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		mockResp       error
		expectedStatus int
	}{
		{
			name: "success",
			requestBody: map[string]interface{}{
				"user_id":      uuid.New().String(),
				"service_name": "svc1",
				"price":        100.0,
				"start_date":   time.Now().Format("2006-01-02"),
			},
			mockResp:       nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid_json",
			requestBody: map[string]interface{}{
				"invalid": "data",
			},
			mockResp:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "storage_error",
			requestBody: map[string]interface{}{
				"user_id":      uuid.New().String(),
				"service_name": "svc1",
				"price":        100.0,
				"start_date":   time.Now().Format("2006-01-02"),
			},
			mockResp:       assert.AnError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockResp == nil && tt.name != "invalid_json" {
				mockStore.On("CreateSubscription", mock.Anything, mock.MatchedBy(func(sub *models.Subscription) bool {
					return sub.ServiceName == tt.requestBody["service_name"].(string) &&
						sub.Price == int(tt.requestBody["price"].(float64)) &&
						sub.UserID.String() == tt.requestBody["user_id"].(string)
				})).Return(tt.mockResp).Once()
			} else if tt.name == "storage_error" {
				mockStore.On("CreateSubscription", mock.Anything, mock.MatchedBy(func(sub *models.Subscription) bool {
					return sub.ServiceName == tt.requestBody["service_name"].(string) &&
						sub.Price == int(tt.requestBody["price"].(float64)) &&
						sub.UserID.String() == tt.requestBody["user_id"].(string)
				})).Return(tt.mockResp).Once()
			}

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/subscriptions", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.CreateSubscription(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response models.Subscription
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, response.ID)
			}

			mockStore.AssertExpectations(t)
		})
	}
}

// TestDeleteSubscription - тестирование эндпоинта удаления подписки.
func TestDeleteSubscription(t *testing.T) {
	mockStore := new(MockStorage)
	handler := &api.Handler{
		Storage: mockStore,
	}

	tests := []struct {
		name           string
		id             string
		mockResp       error
		expectedStatus int
	}{
		{
			name:           "success",
			id:             "123e4567-e89b-12d3-a456-426614174000",
			mockResp:       nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "not_found",
			id:             "00000000-0000-0000-0000-000000000000",
			mockResp:       sql.ErrNoRows,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "internal_error",
			id:             "11111111-1111-1111-1111-111111111111",
			mockResp:       assert.AnError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockID, _ := uuid.Parse(tt.id)
			mockStore.On("DeleteSubscription", mock.Anything, mockID).Return(tt.mockResp).Once()

			req, _ := http.NewRequest("DELETE", "/subscriptions/"+tt.id, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.DeleteSubscription(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockStore.AssertExpectations(t)
		})
	}
}

// TestGetSubscription - тестирование эндпоинта получения подписки по ID.
func TestGetSubscription(t *testing.T) {
	mockStore := new(MockStorage)
	handler := &api.Handler{
		Storage: mockStore,
	}

	tests := []struct {
		name           string
		id             string
		mockResp       *models.Subscription
		mockErr        error
		expectedStatus int
	}{
		{
			name: "success",
			id:   "123e4567-e89b-12d3-a456-426614174000",
			mockResp: &models.Subscription{
				ID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				UserID:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				ServiceName: "svc1",
				Price:       100,
				StartDate:   models.DataOnly(time.Now()),
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid_uuid",
			id:             "invalid-uuid",
			mockResp:       nil,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "internal_error",
			id:             "123e4567-e89b-12d3-a456-426614174000",
			mockResp:       nil,
			mockErr:        assert.AnError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockResp != nil || tt.mockErr != nil {
				mockID, _ := uuid.Parse(tt.id)
				mockStore.On("GetSubscriptionByID", mock.Anything, mockID).Return(tt.mockResp, tt.mockErr).Once()
			}

			req, _ := http.NewRequest("GET", "/subscriptions/"+tt.id, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.GetSubscription(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockStore.AssertExpectations(t)

			if tt.expectedStatus == http.StatusOK {
				var response models.Subscription
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResp.ID, response.ID)
				assert.Equal(t, tt.mockResp.ServiceName, response.ServiceName)
				assert.Equal(t, tt.mockResp.Price, response.Price)
			}
		})
	}
}

// MockLogger - мок логгера для тестирования middleware.
type MockLogger struct {
	logs []string
}

func (m *MockLogger) Enabled(context.Context, slog.Level) bool {
	return true
}

func (m *MockLogger) Handle(ctx context.Context, r slog.Record) error {
	// Собираем сообщение из Record вручную
	var attrs []string
	r.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr.Key+"="+attr.Value.String())
		return true
	})

	msg := r.Level.String() + " " + r.Message + " [" + strings.Join(attrs, " ") + "]"
	m.logs = append(m.logs, msg)
	return nil
}

func (m *MockLogger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return m
}

func (m *MockLogger) WithGroup(name string) slog.Handler {
	return m
}

// type mockResponseWriter struct {
// 	io.Writer
// 	statusCode int // <- Поле для хранения статус кода
// }

// func (m *mockResponseWriter) Header() http.Header { return http.Header{} }

// func (m *mockResponseWriter) WriteHeader(code int) {
// 	m.statusCode = code // <- Обновляем статус код
// }

// func (m *mockResponseWriter) Write(p []byte) (int, error) {
// 	return m.Writer.Write(p)
// }

// TestLoggingMiddleware - тестирование middleware логирования запросов.
func TestLoggingMiddleware(t *testing.T) {
	logger := &MockLogger{}
	slogLogger := slog.New(logger)
	api.SetLogger(slogLogger)

	middleware := api.LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	require.GreaterOrEqual(t, len(logger.logs), 2)
	assert.Contains(t, logger.logs[0], "INFO Request started")
	assert.Contains(t, logger.logs[1], "INFO Request completed")
}
