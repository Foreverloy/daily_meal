package store

import (
	"context"
	"dailymeal/backend/internal/model"
)

func (s *Store) Food(ctx context.Context, id int64, forUpdate bool) (*model.Food, error) {
	var food model.Food
	err := lock(s.db.WithContext(ctx), forUpdate).First(&food, id).Error
	return &food, err
}

func (s *Store) Foods(ctx context.Context, q string, page, size int) ([]model.Food, int64, error) {
	items := []model.Food{}
	var total int64
	query := named(s.db.WithContext(ctx).Model(&model.Food{}), q)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&items).Error
	return items, total, err
}

func (s *Store) CreateFood(ctx context.Context, food *model.Food) error {
	return s.db.WithContext(ctx).Create(food).Error
}
func (s *Store) SaveFood(ctx context.Context, food *model.Food) error {
	return s.db.WithContext(ctx).Save(food).Error
}
