package store

import (
	"dailymeal/backend/internal/model"
	"gorm.io/gorm"
)

func (s *Store) Entry(id int64, forUpdate bool) (*model.MealEntry, error) {
	var entry model.MealEntry
	err := lock(s.db, forUpdate).First(&entry, id).Error
	return &entry, err
}

func (s *Store) Entries(date model.Date, meal string) ([]model.MealEntry, error) {
	items := []model.MealEntry{}
	query := s.db.Where("date = ?", date)
	if meal != "" {
		query = query.Where("meal_type = ?", meal)
	}
	err := query.Order("id ASC").Find(&items).Error
	return items, err
}

func (s *Store) CreateEntry(entry *model.MealEntry) error {
	return s.db.Create(entry).Error
}
func (s *Store) SaveEntry(entry *model.MealEntry) error {
	return s.db.Save(entry).Error
}

func (s *Store) DeleteEntry(id int64) error {
	result := s.db.Delete(&model.MealEntry{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
