package store

import (
	"errors"

	"dailymeal/backend/internal/model"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrMealEntryRequestExists = errors.New("meal entry request already exists")

func (s *Store) MealEntryRequest(key string) (*model.MealEntryRequest, error) {
	var request model.MealEntryRequest
	err := s.db.Where("key = ?", key).First(&request).Error
	return &request, err
}

func (s *Store) CreateMealEntryRequest(request *model.MealEntryRequest) error {
	err := s.db.Create(request).Error
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "meal_entry_requests_pkey" {
		return ErrMealEntryRequestExists
	}
	return err
}
