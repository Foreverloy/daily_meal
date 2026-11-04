package service

import (
	"errors"

	"dailymeal/backend/internal/model"
	"gorm.io/gorm"
)

// TODO：暂未在接口层校验热量与百分比，调用方提交完整目标。
type GoalInput = model.GoalValues

type GoalView struct {
	model.GoalValues
	Targets     map[string]float64 `json:"targets"`
	EffectiveOn *model.Date        `json:"effective_on,omitempty"`
}

type ResolvedGoal struct {
	Date   model.Date `json:"date"`
	Source string     `json:"source"`
	Goal   *GoalView  `json:"goal"`
}

func goalView(values model.GoalValues, effectiveOn *model.Date) *GoalView {
	return &GoalView{GoalValues: values, Targets: values.Targets(), EffectiveOn: effectiveOn}
}

func (s *Service) DefaultGoal() (*GoalView, error) {
	goal, err := s.store.DefaultGoal(s.Today())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return goalView(goal.GoalValues, &goal.EffectiveOn), nil
}

func (s *Service) PutDefaultGoal(in GoalInput) (*GoalView, error) {
	goal := model.GoalSetting{EffectiveOn: s.Today(), GoalValues: in}
	if err := s.store.PutDefaultGoal(&goal); err != nil {
		return nil, err
	}
	return goalView(goal.GoalValues, &goal.EffectiveOn), nil
}

func (s *Service) resolveGoal(date model.Date) (ResolvedGoal, error) {
	result := ResolvedGoal{Date: date, Source: "none"}
	override, err := s.store.DailyGoal(date)
	if err == nil {
		result.Source = "override"
		result.Goal = goalView(override.GoalValues, nil)
		return result, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return result, err
	}
	def, err := s.store.DefaultGoal(date)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	result.Source = "default"
	result.Goal = goalView(def.GoalValues, &def.EffectiveOn)
	return result, nil
}

func (s *Service) DailyGoal(rawDate string) (ResolvedGoal, error) {
	date, err := ParseDate(rawDate)
	if err != nil {
		return ResolvedGoal{}, err
	}
	return s.resolveGoal(date)
}

func (s *Service) PutDailyGoal(rawDate string, in GoalInput) (ResolvedGoal, error) {
	date, err := ParseDate(rawDate)
	if err != nil {
		return ResolvedGoal{}, err
	}
	goal := model.DailyGoal{Date: date, GoalValues: in}
	if err := s.store.PutDailyGoal(&goal); err != nil {
		return ResolvedGoal{}, err
	}
	return ResolvedGoal{Date: date, Source: "override", Goal: goalView(in, nil)}, nil
}

func (s *Service) DeleteDailyGoal(rawDate string) error {
	date, err := ParseDate(rawDate)
	if err != nil {
		return err
	}
	return s.store.DeleteDailyGoal(date)
}
