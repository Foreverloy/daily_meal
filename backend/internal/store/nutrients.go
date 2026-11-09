package store

import "dailymeal/backend/internal/model"

// Nutrients 读取营养素目录，返回稳定的初始化顺序，供接口、保存和计算共用。
func (s *Store) Nutrients() (model.Nutrients, error) {
	items := model.Nutrients{}
	err := s.db.Order("id ASC").Find(&items).Error
	return items, err
}
