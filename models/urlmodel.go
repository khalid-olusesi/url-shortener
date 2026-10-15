package models

import "time"

type URL struct {
	ID          uint      `json:"id"`
	ShortCode   string    `json:"short_code" gorm:"uniqueIndex"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}
