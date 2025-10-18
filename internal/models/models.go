package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type DataOnly time.Time

func (d *DataOnly) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*d = DataOnly(t)
	return nil
}
func (d DataOnly) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	return []byte(`"` + t.Format("2006-01-02") + `"`), nil
}
func (d DataOnly) ToTime() time.Time {
	return time.Time(d).Truncate(time.Second)
}

type Subscription struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ServiceName string    `json:"service_name" db:"service_name"`
	Price       int       `json:"price" db:"price"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	StartDate   DataOnly  `json:"start_date" db:"start_date"`
	EndDate     *DataOnly `json:"end_date,omitempty" db:"end_date"`
	CreatedAt   DataOnly  `json:"created_at" db:"created_at"`
	UpdatedAt   DataOnly  `json:"updated_at" db:"updated_at"`
}
