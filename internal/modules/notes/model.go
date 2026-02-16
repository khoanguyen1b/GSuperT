package notes

import (
	"time"
)

type Note struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Content    string    `gorm:"not null" json:"content"`
	CustomerID string    `gorm:"type:uuid;not null" json:"customer_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
