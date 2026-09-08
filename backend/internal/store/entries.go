package store

import (
	"context"
	"dailymeal/backend/internal/model"
	"gorm.io/gorm"
)

func (s *Store) Entry(ctx context.Context, id int64, forUpdate bool) (*model.MealEntry, error) {
	var entry model.MealEntry
	err := lock(s.db.WithContext(ctx), forUpdate).First(&entry, id).Error
	return &entry, err
}

func (s *Store) Entries(ctx context.Context, date model.Date, meal string) ([]model.MealEntry, error) {
	items := []model.MealEntry{}
	query := s.db.WithContext(ctx).Where("date = ?", date)
	if meal != "" {
		query = query.Where("meal_type = ?", meal)
	}
	err := query.Order("id ASC").Find(&items).Error
	return items, err
}

func (s *Store) CreateEntry(ctx context.Context, entry *model.MealEntry) error {
	return s.db.WithContext(ctx).Create(entry).Error
}
func (s *Store) SaveEntry(ctx context.Context, entry *model.MealEntry) error {
	return s.db.WithContext(ctx).Save(entry).Error
}

func (s *Store) DeleteEntry(ctx context.Context, id int64) error {
	result := s.db.WithContext(ctx).Delete(&model.MealEntry{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
