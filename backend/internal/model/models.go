package model

import (
	"database/sql/driver"
	"fmt"
	"time"

	"dailymeal/backend/internal/nutrition"
)

type Date string

func (d *Date) Scan(value any) error {
	switch v := value.(type) {
	case time.Time:
		*d = Date(v.Format(time.DateOnly))
	case string:
		*d = Date(v)
	case []byte:
		*d = Date(string(v))
	default:
		return fmt.Errorf("invalid database date: %T", value)
	}
	return nil
}
func (d Date) Value() (driver.Value, error) { return string(d), nil }

type Base struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GoalValues struct {
	EnergyKcal          float64 `json:"energy_kcal"`
	ProteinPercent      float64 `json:"protein_percent"`
	CarbohydratePercent float64 `json:"carbohydrate_percent"`
	FatPercent          float64 `json:"fat_percent"`
}

func (g GoalValues) Targets() map[string]float64 {
	return map[string]float64{
		"energy_kcal":    g.EnergyKcal,
		"protein_g":      g.EnergyKcal * (g.ProteinPercent / 100) / 4,
		"carbohydrate_g": g.EnergyKcal * (g.CarbohydratePercent / 100) / 4,
		"fat_g":          g.EnergyKcal * (g.FatPercent / 100) / 9,
	}
}

type GoalSetting struct {
	Base
	EffectiveOn Date `json:"effective_on" gorm:"type:date"`
	GoalValues  `gorm:"embedded"`
}

type DailyGoal struct {
	Base
	Date       Date `json:"date" gorm:"type:date"`
	GoalValues `gorm:"embedded"`
}

type Food struct {
	Base
	Name            string           `json:"name"`
	BasisUnit       string           `json:"basis_unit"`
	NutritionPer100 nutrition.Values `json:"nutrition_per_100" gorm:"column:nutrition_per_100;serializer:json;type:jsonb"`
}

type Recipe struct {
	Base
	Name string `json:"name"`
}

type RecipeIngredient struct {
	Base
	RecipeID int64   `json:"recipe_id"`
	FoodID   int64   `json:"food_id"`
	Quantity float64 `json:"quantity"`
	Food     Food    `json:"-" gorm:"foreignKey:FoodID"`
}

type IngredientSnapshot struct {
	FoodID          int64            `json:"food_id"`
	Name            string           `json:"name"`
	BasisUnit       string           `json:"basis_unit"`
	Quantity        float64          `json:"quantity"`
	NutritionPer100 nutrition.Values `json:"nutrition_per_100"`
}

type Snapshot struct {
	Name            string               `json:"name"`
	BasisUnit       string               `json:"basis_unit,omitempty"`
	NutritionPer100 nutrition.Values     `json:"nutrition_per_100,omitempty"`
	Ingredients     []IngredientSnapshot `json:"ingredients,omitempty"`
}

type MealEntry struct {
	Base
	Date            Date             `json:"date" gorm:"type:date"`
	MealType        string           `json:"meal_type"`
	ItemType        string           `json:"item_type"`
	FoodID          *int64           `json:"food_id"`
	RecipeID        *int64           `json:"recipe_id"`
	Quantity        *float64         `json:"quantity"`
	Snapshot        Snapshot         `json:"snapshot" gorm:"serializer:json;type:jsonb"`
	NutritionTotals nutrition.Totals `json:"nutrition_totals" gorm:"serializer:json;type:jsonb"`
}
