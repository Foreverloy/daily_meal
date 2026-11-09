package service

import (
	"math"
	"testing"

	"dailymeal/backend/internal/agent"
	"dailymeal/backend/internal/model"
)

// imageCatalog 模拟数据库中的营养素定义，包含原有和新增指标。
var imageCatalog = model.Nutrients{
	{Code: "energy", Unit: "kcal"}, {Code: "protein", Unit: "g"}, {Code: "fat", Unit: "g"},
	{Code: "sodium", Unit: "mg"}, {Code: "vitamin_d", Unit: "ug"},
	{Code: "dietary_fiber", Unit: "g"}, {Code: "selenium", Unit: "ug"},
}

// TestFoodImageConvertsOriginalUnits 验证原有及新增指标使用目录指定的单位。
func TestFoodImageConvertsOriginalUnits(t *testing.T) {
	food, err := foodFromRecognition(agent.FoodRecognition{
		Name: " 牛奶 ", BasisQuantity: 250, BasisUnit: "ml",
		Nutrients: []agent.NutrientReading{
			{Key: "energy", Value: 627.6, Unit: "kJ"},
			{Key: "protein", Value: 8000, Unit: "mg"},
			{Key: "fat", Value: 0, Unit: "g"},
			{Key: "sodium", Value: 0.125, Unit: "g"},
			{Key: "vitamin_d", Value: 0.005, Unit: "mg"},
			{Key: "dietary_fiber", Value: 2500, Unit: "mg"},
			{Key: "selenium", Value: 0.025, Unit: "mg"},
		},
	}, imageCatalog)
	if err != nil {
		t.Fatal(err)
	}
	if food.Name != "牛奶" || food.BasisUnit != "ml" {
		t.Fatalf("unexpected food: %+v", food)
	}
	for key, want := range map[string]float64{"energy_kcal": 60, "protein_g": 3.2, "fat_g": 0, "sodium_mg": 50, "vitamin_d_ug": 2, "dietary_fiber_g": 1, "selenium_ug": 10} {
		got := food.NutritionPer100[key]
		if got == nil || math.Abs(*got-want) > 1e-8 {
			t.Errorf("%s: got %v want %v", key, got, want)
		}
	}
	if _, exists := food.NutritionPer100["calcium_mg"]; exists {
		t.Fatal("unread nutrient must be omitted, not submitted as null")
	}
}

// TestFoodImageBasisConversion 验证截图份量由 kg/L 换算为每 100g/ml。
func TestFoodImageBasisConversion(t *testing.T) {
	for _, unit := range []string{"kg", "L"} {
		food, err := foodFromRecognition(agent.FoodRecognition{
			BasisQuantity: 0.25, BasisUnit: unit,
			Nutrients: []agent.NutrientReading{{Key: "energy", Value: 150, Unit: "kcal"}},
		}, imageCatalog)
		if err != nil || *food.NutritionPer100["energy_kcal"] != 60 {
			t.Fatalf("%s conversion failed: %+v %v", unit, food, err)
		}
	}
}

// TestFoodImageMissingInformation 验证识别信息缺失时直接返回错误。
func TestFoodImageMissingInformation(t *testing.T) {
	for _, reading := range []agent.FoodRecognition{
		{BasisUnit: "g"},
		{BasisQuantity: 100},
		{BasisQuantity: 100, BasisUnit: "g"},
	} {
		if _, err := foodFromRecognition(reading, imageCatalog); err == nil {
			t.Fatalf("missing information accepted: %+v", reading)
		}
	}
}
