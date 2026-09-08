package store

import (
	"context"
	"dailymeal/backend/internal/model"
	"gorm.io/gorm/clause"
)

func (s *Store) DefaultGoal(ctx context.Context, date model.Date) (*model.GoalSetting, error) {
	var goal model.GoalSetting
	err := s.db.WithContext(ctx).Where("effective_on <= ?", date).Order("effective_on DESC").First(&goal).Error
	return &goal, err
}

func (s *Store) DailyGoal(ctx context.Context, date model.Date) (*model.DailyGoal, error) {
	var goal model.DailyGoal
	err := s.db.WithContext(ctx).Where("date = ?", date).First(&goal).Error
	return &goal, err
}

var goalColumns = []string{"energy_kcal", "protein_percent", "carbohydrate_percent", "fat_percent", "updated_at"}

func (s *Store) PutDefaultGoal(ctx context.Context, goal *model.GoalSetting) error {
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "effective_on"}}, DoUpdates: clause.AssignmentColumns(goalColumns)}, clause.Returning{}).Create(goal).Error
}

func (s *Store) PutDailyGoal(ctx context.Context, goal *model.DailyGoal) error {
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "date"}}, DoUpdates: clause.AssignmentColumns(goalColumns)}, clause.Returning{}).Create(goal).Error
}

func (s *Store) DeleteDailyGoal(ctx context.Context, date model.Date) error {
	return s.db.WithContext(ctx).Where("date = ?", date).Delete(&model.DailyGoal{}).Error
}
