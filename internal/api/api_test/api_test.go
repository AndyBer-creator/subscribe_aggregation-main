package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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

// TestListSubscriptions - тестирование эндпоинта получения списка подписок.
func TestListSubscriptions(t *testing.T) {
	mockStore := new(MockStorage)
	handler := &api.Handler{
		Storage: mockStore,
	}

	tests := []struct {
		name           string
		queryParams    map[string]string
		mockResp       []models.Subscription
		mockErr        error
		expectedStatus int
		expectedBody   []models.Subscription
	}{
		{
			name: "success",
			queryParams: map[string]string{
				"page":  "1",
				"limit": "2",
			},
			mockResp: []models.Subscription{
				{ID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), ServiceName: "svc1", Price: 100},
				{ID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"), ServiceName: "svc2", Price: 200},
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedBody: []models.Subscription{
				{ID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), ServiceName: "svc1", Price: 100},
				{ID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"), ServiceName: "svc2", Price: 200},
			},
		},
		{
			name: "invalid_page",
			queryParams: map[string]string{
				"page":  "invalid",
				"limit": "2",
			},
			mockResp: []models.Subscription{
				{ID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), ServiceName: "svc1", Price: 100},
				{ID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"), ServiceName: "svc2", Price: 200},
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedBody: []models.Subscription{
				{ID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), ServiceName: "svc1", Price: 100},
				{ID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"), ServiceName: "svc2", Price: 200},
			},
		},
		{
			name: "invalid_limit",
			queryParams: map[string]string{
				"page":  "1",
				"limit": "invalid",
			},
			mockResp: []models.Subscription{
				{ID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), ServiceName: "svc1", Price: 100},
				{ID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"), ServiceName: "svc2", Price: 200},
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedBody: []models.Subscription{
				{ID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), ServiceName: "svc1", Price: 100},
				{ID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"), ServiceName: "svc2", Price: 200},
			},
		},
		{
			name: "storage_error",
			queryParams: map[string]string{
				"page":  "1",
				"limit": "2",
			},
			mockResp:       nil,
			mockErr:        assert.AnError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем мок - не ограничиваем количество вызовов для гибкости
			mockStore.On("ListSubscriptions", mock.Anything, mock.Anything, mock.Anything).
				Return(tt.mockResp, tt.mockErr)

			req := httptest.NewRequest("GET", "/subscriptions", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			rr := httptest.NewRecorder()
			handler.ListSubscriptions(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response []models.Subscription
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			} else if tt.expectedStatus >= 400 {
				// При ошибках можно проверить тело ответа с сообщением
				assert.Contains(t, rr.Body.String(), "internal server error")
			}

			mockStore.AssertExpectations(t)
			mockStore.ExpectedCalls = nil
			mockStore.Calls = nil
		})
	}
}

func TestSumSubscriptionsCostHandler(t *testing.T) {
	mockStore := new(MockStorage)
	handler := &api.Handler{
		Storage: mockStore,
	}

	tests := []struct {
		name           string
		queryString    string
		sumReturn      int64
		sumErr         error
		wantStatusCode int
		wantBody       string
		setupMock      func()
	}{
		{
			name:           "Valid request returns sum",
			queryString:    "?user_id=123&service_name=svc1&start_date=01-2024&end_date=04-2024",
			sumReturn:      700,
			sumErr:         nil,
			wantStatusCode: http.StatusOK,
			wantBody:       `{"total_price":700}`,
			setupMock: func() {
				mockStore.On("SumSubscriptionsCost", mock.Anything, "123", "svc1", mock.Anything, mock.Anything).
					Return(int64(700), nil).Once()
			},
		},
		{
			name:           "Invalid start_date format",
			queryString:    "?start_date=2024-01",
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "invalid start_date format, expected MM-YYYY\n",
			setupMock: func() {
				// мок не вызывается в этом случае, ничего не нужно
			},
		},
		{
			name:           "Storage error",
			queryString:    "?user_id=123",
			sumReturn:      0,
			sumErr:         errors.New("DB failure"),
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "DB failure\n",
			setupMock: func() {
				mockStore.On("SumSubscriptionsCost", mock.Anything, "123", "", mock.Anything, mock.Anything).
					Return(int64(0), errors.New("DB failure")).Once()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockStore.ExpectedCalls = nil // сброс предыдущих ожиданий
			tc.setupMock()

			req := httptest.NewRequest("GET", "/subscriptions/sum"+tc.queryString, nil)
			w := httptest.NewRecorder()

			handler.SumSubscriptionsCostHandler(w, req)
			res := w.Result()

			if res.StatusCode != tc.wantStatusCode {
				t.Errorf("expected status %d, got %d", tc.wantStatusCode, res.StatusCode)
			}

			bodyBytes := w.Body.Bytes()
			if string(bodyBytes) != tc.wantBody {
				if strings.HasPrefix(tc.wantBody, "{") {
					var wantMap, gotMap map[string]int64
					_ = json.Unmarshal([]byte(tc.wantBody), &wantMap)
					_ = json.Unmarshal(bodyBytes, &gotMap)
					if !equalMaps(wantMap, gotMap) {
						t.Errorf("expected body %s, got %s", tc.wantBody, string(bodyBytes))
					}
				} else {
					t.Errorf("expected body %s, got %s", tc.wantBody, string(bodyBytes))
				}
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func equalMaps(a, b map[string]int64) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func TestUpdateSubscriptionHandler(t *testing.T) {
	mockStore := new(MockStorage)
	handler := &api.Handler{
		Storage: mockStore,
	}

	validUUID := uuid.New()
	validSub := models.Subscription{
		ID:          validUUID,
		UserID:      uuid.New(),
		ServiceName: "svc1",
		Price:       123,
		StartDate:   models.DataOnly(time.Now()),
	}
	validSubJSON, _ := json.Marshal(validSub)

	tests := []struct {
		name           string
		method         string
		url            string
		body           string
		setupMock      func()
		wantStatusCode int
		wantBody       string
	}{
		{
			name:   "Valid update",
			method: http.MethodPut,
			url:    "/subscriptions/" + validUUID.String(),
			body:   string(validSubJSON),
			setupMock: func() {
				mockStore.On("UpdateSubscription", mock.Anything, mock.MatchedBy(func(sub *models.Subscription) bool {
					return sub.ID == validUUID
				})).Return(nil).Once()
			},
			wantStatusCode: http.StatusOK,
			wantBody:       string(validSubJSON) + "\n",
		},
		{
			name:           "Invalid UUID",
			method:         http.MethodPut,
			url:            "/subscriptions/invalid-uuid",
			body:           string(validSubJSON),
			setupMock:      func() {}, // мок не вызывается
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "invalid UUID\n",
		},
		{
			name:           "Invalid JSON body",
			method:         http.MethodPut,
			url:            "/subscriptions/" + validUUID.String(),
			body:           `{"invalid_json":`,
			setupMock:      func() {}, // мок не вызывается
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "unexpected end of JSON input\n",
		},
		{
			name:   "Storage returns sql.ErrNoRows",
			method: http.MethodPut,
			url:    "/subscriptions/" + validUUID.String(),
			body:   string(validSubJSON),
			setupMock: func() {
				mockStore.On("UpdateSubscription", mock.Anything, mock.Anything).Return(sql.ErrNoRows).Once()
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       "subscription not found\n",
		},
		{
			name:   "Storage returns error",
			method: http.MethodPut,
			url:    "/subscriptions/" + validUUID.String(),
			body:   string(validSubJSON),
			setupMock: func() {
				mockStore.On("UpdateSubscription", mock.Anything, mock.Anything).Return(errors.New("DB error")).Once()
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "DB error\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockStore.ExpectedCalls = nil
			tc.setupMock()

			req := httptest.NewRequest(tc.method, tc.url, strings.NewReader(tc.body))
			w := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Put("/subscriptions/{id}", handler.UpdateSubscription)
			r.ServeHTTP(w, req)

			res := w.Result()
			bodyBytes := w.Body.Bytes()

			if res.StatusCode != tc.wantStatusCode {
				t.Errorf("expected status %d, got %d", tc.wantStatusCode, res.StatusCode)
			}

			switch tc.name {
			case "Invalid JSON body":
				if !strings.Contains(string(bodyBytes), "unexpected EOF") {
					t.Errorf("expected body to contain unexpected EOF, got %s", string(bodyBytes))
				}
			default:
				if string(bodyBytes) != tc.wantBody {
					t.Errorf("expected body %q, got %q", tc.wantBody, string(bodyBytes))
				}
			}

			mockStore.AssertExpectations(t)
		})
	}
}
