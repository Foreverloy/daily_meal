package service

import (
	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/nutrition"
)

// Comparison 对照一个营养指标的每日目标与实际摄入，单位由所在 map 的指标名确定。
type Comparison struct {
	// Target 是目标值；当天未设置目标时为 nil。
	Target *float64 `json:"target"`
	// Intake 是摄入的已知合计及完整性，例如蛋白质已知 80g 但有一项食物含量未知。
	Intake nutrition.Amount `json:"intake"`
	// Remaining 为目标减摄入，可为负数表示超出；无目标或摄入不完整时为 nil。
	Remaining *float64 `json:"remaining"`
}

// Summary 汇总一天的目标、三餐摄入和全天营养，使用饮食记录里保存的营养快照结果。
type Summary struct {
	// Date 是汇总的业务日期，如 "2026-11-09"。
	Date model.Date `json:"date"`
	// Goal 是当天最终生效的目标及其来源。
	Goal ResolvedGoal `json:"goal"`
	// Meals 按 breakfast、lunch、dinner 分别保存三餐营养合计。
	Meals map[string]nutrition.Totals `json:"meals"`
	// NutritionTotals 是全天所有饮食记录的合计，包含热量、三大营养素及微量营养素。
	NutritionTotals nutrition.Totals `json:"nutrition_totals"`
	// Comparison 仅对热量、蛋白质、碳水和脂肪做目标对比，钠、维生素等不设目标。
	Comparison map[string]Comparison `json:"comparison"`
}

// DailySummary 按当前营养素目录汇总历史记录，旧快照缺少的指标按未知处理。
func (s *Service) DailySummary(rawDate string) (*Summary, error) {
	date, err := ParseDate(rawDate)
	if err != nil {
		return nil, err
	}
	catalog, err := s.store.Nutrients()
	if err != nil {
		return nil, err
	}
	keys := catalog.Keys()
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
		result.Meals[meal] = nutrition.Sum(keys, parts...)
	}
	// 直接汇总记录，避免空餐的零把全天未知营养变成已知零。
	result.NutritionTotals = nutrition.Sum(keys, all...)
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
