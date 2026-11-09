package store

import (
	"dailymeal/backend/internal/model"
	"gorm.io/gorm"
)

// Food 读取食物及关联含量，并按营养素目录组织接口数据。
func (s *Store) Food(id int64) (*model.Food, error) {
	var food model.Food
	if err := s.db.Preload("Nutrients").First(&food, id).Error; err != nil {
		return nil, err
	}
	catalog, err := s.Nutrients()
	if err != nil {
		return nil, err
	}
	food.LoadNutrition(catalog)
	return &food, nil
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
	if err := query.Preload("Nutrients").Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	catalog, err := s.Nutrients()
	if err != nil {
		return nil, 0, err
	}
	for i := range items {
		items[i].LoadNutrition(catalog)
	}
	return items, total, nil
}

// CreateFood 由 GORM 创建食物及其含量关联，沿用框架默认写入事务。
func (s *Store) CreateFood(food *model.Food) error {
	return s.db.Create(food).Error
}

// SaveFood 同时保存食物和关联含量，FullSaveAssociations 使零值和 nil 也能更新。
func (s *Store) SaveFood(food *model.Food) error {
	return s.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(food).Error
}
