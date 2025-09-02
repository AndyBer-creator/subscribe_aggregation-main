package models

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey" db:"id"`
	ServiceName string     `json:"service_name" gorm:"not null" db:"service_name"`
	Price       int        `json:"price" gorm:"not null" db:"price"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null" db:"user_id"`
	StartDate   time.Time  `json:"start_date" gorm:"not null" db:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date"`
}
