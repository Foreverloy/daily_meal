package store

import "dailymeal/backend/internal/model"

func (s *Store) Food(id int64) (*model.Food, error) {
	var food model.Food
	err := s.db.First(&food, id).Error
	return &food, err
}

// Foods 将名称和收藏条件同时用于总数与分页查询，保持原有 id DESC 排序。
func (s *Store) Foods(q string, favorite *bool, page, size int) ([]model.Food, int64, error) {
	items := []model.Food{}
	var total int64
	query := named(s.db.Model(&model.Food{}), q)

	if favorite != nil && *favorite {
		query = query.Where("favorite = ?", true)
	}
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
