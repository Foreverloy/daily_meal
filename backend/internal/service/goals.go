package service

import (
	"errors"

	"dailymeal/backend/internal/model"
	"gorm.io/gorm"
)

// GoalInput 是 GoalValues 的类型别名，接口直接复用热量与供能占比字段。
// TODO：暂未在接口层校验热量与百分比，调用方提交完整目标。
type GoalInput = model.GoalValues

// GoalView 在原始目标上补充换算结果，方便前端直接展示各营养素的克数目标。
type GoalView struct {
	// GoalValues 嵌入输入的总热量和三大营养素供能占比。
	model.GoalValues
	// Targets 是热量及三大营养素目标，如 {"energy_kcal":2000,"protein_g":100,...}。
	Targets map[string]float64 `json:"targets"`
	// EffectiveOn 是默认目标的生效日；单日覆盖目标为 nil，JSON 中省略此字段。
	EffectiveOn *model.Date `json:"effective_on,omitempty"`
}

// ResolvedGoal 是某一天最终使用的目标，优先取单日覆盖，其次取当天有效的默认目标。
type ResolvedGoal struct {
	// Date 是正在查询的日期，如 "2026-11-09"。
	Date model.Date `json:"date"`
	// Source 标明来源：override（单日覆盖）、default（默认）、none（未设置）。
	Source string `json:"source"`
	// Goal 是最终目标内容；Source="none" 时为 nil，返回 JSON null。
	Goal *GoalView `json:"goal"`
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
