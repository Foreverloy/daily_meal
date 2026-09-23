package store

import "dailymeal/backend/internal/model"

func (s *Store) Food(id int64, forUpdate bool) (*model.Food, error) {
	var food model.Food
	err := lock(s.db, forUpdate).First(&food, id).Error
	return &food, err
}

func (s *Store) Foods(q string, page, size int) ([]model.Food, int64, error) {
	items := []model.Food{}
	var total int64
	query := named(s.db.Model(&model.Food{}), q)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&items).Error
	return items, total, err
}

func (s *Store) CreateFood(food *model.Food) error {
	return s.db.Create(food).Error
}
func (s *Store) SaveFood(food *model.Food) error {
	return s.db.Save(food).Error
}
