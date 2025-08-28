package models

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
<<<<<<< HEAD
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ServiceName string     `json:"service_name" gorm:"not null"`
	Price       int        `json:"price" gorm:"not null"` // рубли
	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	StartDate   time.Time  `json:"start_date" gorm:"not null"`
	EndDate     *time.Time `json:"end_date,omitempty"`
=======
	ID          uuid.UUID  `json:"id" db:"id"`
	ServiceName string     `json:"service_name" db:"service_name"`
	Price       int        `json:"price" db:"price"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	StartDate   time.Time  `json:"start_date" db:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
>>>>>>> 78bf63b (updated)
}
