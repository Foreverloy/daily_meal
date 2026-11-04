package service

import (
	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/nutrition"
)

type Comparison struct {
	Target    *float64         `json:"target"`
	Intake    nutrition.Amount `json:"intake"`
	Remaining *float64         `json:"remaining"`
}

type Summary struct {
	Date            model.Date                  `json:"date"`
	Goal            ResolvedGoal                `json:"goal"`
	Meals           map[string]nutrition.Totals `json:"meals"`
	NutritionTotals nutrition.Totals            `json:"nutrition_totals"`
	Comparison      map[string]Comparison       `json:"comparison"`
}

func (s *Service) DailySummary(rawDate string) (*Summary, error) {
	date, err := ParseDate(rawDate)
	if err != nil {
		return nil, err
	}
	result := &Summary{Date: date, Meals: map[string]nutrition.Totals{}, Comparison: map[string]Comparison{}}
	// TODO：目标和记录分开读取，暂不保证并发修改时来自同一数据库快照。
	result.Goal, err = s.resolveGoal(date)
	if err != nil {
		return nil, err
	}
	entries, err := s.store.Entries(date, "")
	if err != nil {
		return nil, err
	}
	all := make([]nutrition.Totals, 0, len(entries))
	byMeal := map[string][]nutrition.Totals{"breakfast": {}, "lunch": {}, "dinner": {}}
	for _, entry := range entries {
		all = append(all, entry.NutritionTotals)
		byMeal[entry.MealType] = append(byMeal[entry.MealType], entry.NutritionTotals)
	}
	for meal, parts := range byMeal {
		result.Meals[meal] = nutrition.Sum(parts...)
	}
	// 直接汇总记录，避免空餐的零把全天未知营养变成已知零。
	result.NutritionTotals = nutrition.Sum(all...)
	for _, key := range []string{"energy_kcal", "protein_g", "carbohydrate_g", "fat_g"} {
		intake := result.NutritionTotals[key]
		comparison := Comparison{Intake: intake}
		if result.Goal.Goal != nil {
			target := result.Goal.Goal.Targets[key]
			comparison.Target = &target
			if intake.Complete && intake.KnownTotal != nil {
				comparison.Remaining = nutrition.Number(target - *intake.KnownTotal)
			}
		}
		result.Comparison[key] = comparison
	}
	return result, nil
}
