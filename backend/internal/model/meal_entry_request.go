package model

import "time"

// MealEntryRequest retains the original creation result even after the entry is edited or deleted.
type MealEntryRequest struct {
	Key         string `gorm:"primaryKey"`
	RequestHash string
	Response    MealEntry `gorm:"serializer:json;type:jsonb"`
	CreatedAt   time.Time
}
