package models

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ServiceName string     `json:"service_name" gorm:"not null"`
	Price       int        `json:"price" gorm:"not null"` // рубли
	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	StartDate   time.Time  `json:"start_date" gorm:"not null"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}
