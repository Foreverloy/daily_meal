package store

import "dailymeal/backend/internal/model"

func (s *Store) CreateChat(request *model.ChatRequest) error {
	return s.db.Create(request).Error
}

func (s *Store) SaveChat(request *model.ChatRequest) error {
	return s.db.Save(request).Error
}
