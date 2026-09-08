package service

import (
	"encoding/json"
	"testing"
	"time"

	"dailymeal/backend/internal/nutrition"
)

func TestBusinessDateAndGoalTargets(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	svc := New(nil, location, func() time.Time { return time.Date(2026, 9, 8, 16, 1, 0, 0, time.UTC) })
	if svc.Today() != "2026-09-09" {
		t.Fatalf("wrong business date: %s", svc.Today())
	}
	input := GoalInput{nutrition.Number(2000), nutrition.Number(30), nutrition.Number(40), nutrition.Number(30)}
	goal, err := input.values()
	if err != nil {
		t.Fatal(err)
	}
	if goal.Targets()["protein_g"] != 150 || goal.Targets()["carbohydrate_g"] != 200 || goal.Targets()["fat_g"] != 200.0/3 {
		t.Fatalf("wrong targets: %v", goal.Targets())
	}
	for _, invalid := range []GoalInput{
		{}, {nutrition.Number(2000), nil, nutrition.Number(100), nutrition.Number(0)},
		{nutrition.Number(0), nutrition.Number(30), nutrition.Number(40), nutrition.Number(30)},
		{nutrition.Number(2000), nutrition.Number(30), nutrition.Number(40), nutrition.Number(20)},
		{nutrition.Number(2000), nutrition.Number(-10), nutrition.Number(100), nutrition.Number(10)},
	} {
		if _, err := invalid.values(); err == nil {
			t.Fatalf("accepted %+v", invalid)
		}
	}
}

func TestParseDate(t *testing.T) {
	for _, value := range []string{"", "2026-2-03", "2026-02-29", "0000-01-01", "2026-09-08T00:00:00Z"} {
		if _, err := ParseDate(value); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
	if _, err := ParseDate("2024-02-29"); err != nil {
		t.Fatal(err)
	}
}

func TestPatchPresence(t *testing.T) {
	var input FoodPatch
	if err := json.Unmarshal([]byte(`{"nutrition_per_100":{"protein_g":null,"fat_g":0}}`), &input); err != nil {
		t.Fatal(err)
	}
	if input.Name.Set || !input.NutritionPer100.Set || input.NutritionPer100.Null {
		t.Fatal("wrong field presence")
	}
	if value, ok := input.NutritionPer100.Value["protein_g"]; !ok || value != nil {
		t.Fatal("lost explicit null")
	}
	if value := input.NutritionPer100.Value["fat_g"]; value == nil || *value != 0 {
		t.Fatal("lost known zero")
	}
}
