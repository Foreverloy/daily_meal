package service

import "dailymeal/backend/internal/model"

// Nutrients 返回支持的营养素及标准单位，调用方可据此展示和提交营养数据。
func (s *Service) Nutrients() (model.Nutrients, error) {
	return s.store.Nutrients()
}
