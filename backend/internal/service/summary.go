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
	err = s.transaction(func(tx *Service) error {
		var err error
		result.Goal, err = tx.resolveGoal(date)
		if err != nil {
			return err
		}
		entries, err := tx.store.Entries(date, "")
		if err != nil {
			return err
		}
		all := make([]nutrition.Totals, 0, len(entries))
		byMeal := map[string][]nutrition.Totals{"breakfast": {}, "lunch": {}, "dinner": {}}
		for _, entry := range entries {
			all = append(all, entry.NutritionTotals)
			byMeal[entry.MealType] = append(byMeal[entry.MealType], entry.NutritionTotals)
		}
		for meal, parts := range byMeal {
			result.Meals[meal], err = nutrition.Sum(parts...)
			if err != nil {
				return Invalid("nutrition_totals", err.Error())
			}
		}
		// Sum entries directly: an empty meal's zero must not turn an all-unknown day into a known zero.
		result.NutritionTotals, err = nutrition.Sum(all...)
		if err != nil {
			return Invalid("nutrition_totals", err.Error())
		}
		for _, k := range []string{"energy_kcal", "protein_g", "carbohydrate_g", "fat_g"} {
			a := result.NutritionTotals[k]
			comparison := Comparison{Intake: a}
			if result.Goal.Goal != nil {
				target := result.Goal.Goal.Targets[k]
				comparison.Target = &target
				if a.Complete && a.KnownTotal != nil {
					comparison.Remaining = nutrition.Number(target - *a.KnownTotal)
				}
			}
			result.Comparison[k] = comparison
		}
		return nil
	})
	return result, err
}
