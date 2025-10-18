package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"subscribe_aggregation-main/internal/api"
	"subscribe_aggregation-main/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockStorage) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockStorage) ListSubscriptions(ctx context.Context, page, limit int) ([]models.Subscription, error) {
	args := m.Called(ctx, page, limit)
	return args.Get(0).([]models.Subscription), args.Error(1)
}

func (m *MockStorage) UpdateSubscription(ctx context.Context, sub *models.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockStorage) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStorage) SumSubscriptionsCost(ctx context.Context, userID, serviceName string, filterStart, filterEnd time.Time) (int64, error) {
	args := m.Called(ctx, userID, serviceName, filterStart, filterEnd)
	return args.Get(0).(int64), args.Error(1)
}

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
			// Настройка мока только для успешного случая и storage_error
			if tt.mockResp == nil && tt.name != "invalid_json" {
				mockStore.On("CreateSubscription", mock.Anything, mock.MatchedBy(func(sub *models.Subscription) bool {
					return sub.ServiceName == tt.requestBody["service_name"].(string) &&
						sub.Price == int(tt.requestBody["price"].(float64)) &&
						sub.UserID.String() == tt.requestBody["user_id"].(string)
				})).Return(tt.mockResp).Once()
			} else if tt.name == "storage_error" {
				// Для storage_error также ожидаем вызов CreateSubscription
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
			// Настройка мока
			if tt.mockResp != nil || tt.mockErr != nil {
				mockID, _ := uuid.Parse(tt.id)
				mockStore.On("GetSubscriptionByID", mock.Anything, mockID).Return(tt.mockResp, tt.mockErr).Once()
			}

			// Создание запроса и установка параметров маршрута
			req, _ := http.NewRequest("GET", "/subscriptions/"+tt.id, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.GetSubscription(rr, req)

			// Проверка результата
			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockStore.AssertExpectations(t)

			// Дополнительная проверка для успешного случая
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
