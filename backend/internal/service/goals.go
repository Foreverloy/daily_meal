package service

import (
	"errors"
	"math"

	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/nutrition"
	"gorm.io/gorm"
)

type GoalInput struct {
	EnergyKcal          *float64 `json:"energy_kcal"`
	ProteinPercent      *float64 `json:"protein_percent"`
	CarbohydratePercent *float64 `json:"carbohydrate_percent"`
	FatPercent          *float64 `json:"fat_percent"`
}

func (in GoalInput) values() (model.GoalValues, error) {
	for k, v := range map[string]*float64{"energy_kcal": in.EnergyKcal, "protein_percent": in.ProteinPercent, "carbohydrate_percent": in.CarbohydratePercent, "fat_percent": in.FatPercent} {
		if v == nil || !nutrition.Finite(*v) {
			return model.GoalValues{}, Invalid(k, "必填且须为有限数值")
		}
		if k == "energy_kcal" {
			if *v <= 0 {
				return model.GoalValues{}, Invalid(k, "热量必须大于 0")
			}
		} else if *v < 0 || *v > 100 {
			return model.GoalValues{}, Invalid(k, "百分比须在 0～100 之间")
		}
	}
	if math.Abs(*in.ProteinPercent+*in.CarbohydratePercent+*in.FatPercent-100) > 1e-8 {
		return model.GoalValues{}, Invalid("percentages", "供能百分比合计必须为 100")
	}
	return model.GoalValues{EnergyKcal: *in.EnergyKcal, ProteinPercent: *in.ProteinPercent, CarbohydratePercent: *in.CarbohydratePercent, FatPercent: *in.FatPercent}, nil
}

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
	values, err := in.values()
	if err != nil {
		return nil, err
	}
	goal := model.GoalSetting{EffectiveOn: s.Today(), GoalValues: values}
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
	var result ResolvedGoal
	err = s.transaction(func(tx *Service) error { var err error; result, err = tx.resolveGoal(date); return err })
	return result, err
}

func (s *Service) PutDailyGoal(rawDate string, in GoalInput) (ResolvedGoal, error) {
	date, err := ParseDate(rawDate)
	if err != nil {
		return ResolvedGoal{}, err
	}
	values, err := in.values()
	if err != nil {
		return ResolvedGoal{}, err
	}
	goal := model.DailyGoal{Date: date, GoalValues: values}
	if err := s.store.PutDailyGoal(&goal); err != nil {
		return ResolvedGoal{}, err
	}
	return ResolvedGoal{Date: date, Source: "override", Goal: goalView(values, nil)}, nil
}

func (s *Service) DeleteDailyGoal(rawDate string) error {
	date, err := ParseDate(rawDate)
	if err != nil {
		return err
	}
	return s.store.DeleteDailyGoal(date)
}
