package service

import (
	"time"

	"dailymeal/backend/internal/config"
	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/store"
)

type Service struct {
	store *store.Store
}

func New(db *store.Store) *Service {
	return &Service{store: db}
}

func (s *Service) Today() model.Date {
	return model.Date(time.Now().In(config.Location).Format(time.DateOnly))
}

func ParseDate(date string) (model.Date, error) {
	_, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return "", err
	}
	return model.Date(date), nil
}

type Page struct {
	// Favorite 仅用于食物查询：true 只查已收藏，nil 或 false 查询全部。
	Favorite *bool
	Query    string
	Number   int
	Size     int
}

type List[T any] struct {
	Items    []T   `json:"items"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}
