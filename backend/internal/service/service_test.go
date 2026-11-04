package service

import (
	"encoding/json"
	"testing"
)

func TestGoalTargets(t *testing.T) {
	goal := GoalInput{EnergyKcal: 2000, ProteinPercent: 30, CarbohydratePercent: 40, FatPercent: 30}
	if goal.Targets()["protein_g"] != 150 || goal.Targets()["carbohydrate_g"] != 200 || goal.Targets()["fat_g"] != 200.0/3 {
		t.Fatalf("wrong targets: %v", goal.Targets())
	}
}

func TestParseDate(t *testing.T) {
	for _, value := range []string{"", "not-a-date", "2026-02-29", "2026-09-08T00:00:00Z"} {
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
	if input.Name.Set || !input.NutritionPer100.Set {
		t.Fatal("wrong field presence")
	}
	if value, ok := input.NutritionPer100.Value["protein_g"]; !ok || value != nil {
		t.Fatal("lost explicit null")
	}
	if value := input.NutritionPer100.Value["fat_g"]; value == nil || *value != 0 {
		t.Fatal("lost known zero")
	}
}

func TestFoodPatchFavoritePresence(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		want       Field[bool]
		invalid    bool
	}{
		{"omitted", `{}`, Field[bool]{}, false},
		{"favorite", `{"favorite":true}`, Field[bool]{Set: true, Value: true}, false},
		{"unfavorite", `{"favorite":false}`, Field[bool]{Set: true}, false},
		{"string", `{"favorite":"true"}`, Field[bool]{}, true},
		{"number", `{"favorite":1}`, Field[bool]{}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var in FoodPatch
			err := json.Unmarshal([]byte(tc.body), &in)
			if tc.invalid {
				if err == nil {
					t.Fatal("accepted a non-boolean favorite")
				}
				return
			}
			if err != nil || in.Favorite != tc.want {
				t.Fatalf("favorite=%+v want=%+v err=%v", in.Favorite, tc.want, err)
			}
		})
	}
}
