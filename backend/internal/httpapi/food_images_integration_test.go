package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dailymeal/backend/internal/agent"
	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/nutrition"
)

func TestFoodImageCreatesAndUpdatesPersistedFood(t *testing.T) {
	f := setup(t)
	reading := agent.FoodRecognition{Name: "截图牛奶", BasisQuantity: 250, BasisUnit: "ml",
		Nutrients: []agent.NutrientReading{
			{Key: "energy", Value: 627.6, Unit: "kJ"},
			{Key: "protein", Value: 8, Unit: "g"},
			{Key: "fat", Value: 0, Unit: "g"},
		}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			ImageDataURL  string         `json:"image_data_url"`
			ModelSettings map[string]any `json:"model_settings"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Error(err)
		}
		if r.URL.Path != "/api/v1/food-recognitions" || input.ImageDataURL != "data:image/png;base64,test" || input.ModelSettings["api_key"] != "test-secret" {
			t.Error("image/configuration not forwarded to recognition endpoint")
		}
		_ = json.NewEncoder(w).Encode(reading)
	}))
	defer server.Close()
	chatConfig(t, server.URL)
	input := map[string]any{"image_data_url": "data:image/png;base64,test", "name": "我的牛奶"}
	food := response[model.Food](t, f, "POST", "/api/v1/foods/from-image", input, 201)
	stored := response[model.Food](t, f, "GET", path("foods", food.ID), nil, 200)
	if stored.Name != "我的牛奶" || *stored.NutritionPer100["protein_g"] != 3.2 || stored.NutritionPer100["calcium_mg"] != nil || *stored.NutritionPer100["fat_g"] != 0 {
		t.Fatalf("incorrect saved food: %+v", stored)
	}
	response[model.Food](t, f, "PATCH", path("foods", food.ID), map[string]any{
		"nutrition_per_100": nutrition.Values{"calcium_mg": nutrition.Number(120)}, "favorite": true,
	}, 200)
	reading.Nutrients = []agent.NutrientReading{{Key: "protein", Value: 10, Unit: "g"}}
	reading.Name = ""
	delete(input, "name")
	updated := response[model.Food](t, f, "PATCH", path("foods", food.ID)+"/from-image", input, 200)
	if updated.Name != stored.Name || !updated.Favorite || *updated.NutritionPer100["calcium_mg"] != 120 || *updated.NutritionPer100["protein_g"] != 4 || *updated.NutritionPer100["energy_kcal"] != *stored.NutritionPer100["energy_kcal"] {
		t.Fatalf("update lost existing fields: %+v", updated)
	}
	reading.BasisUnit = "g"
	f.request(t, "PATCH", path("foods", food.ID)+"/from-image", input, 422)
	reading.BasisUnit = "ml"
	f.request(t, "POST", "/api/v1/foods/from-image", input, 422) // 名称和热量缺失。
	reading.BasisQuantity = 0
	f.request(t, "PATCH", path("foods", food.ID)+"/from-image", input, 422)
	after := response[model.Food](t, f, "GET", path("foods", food.ID), nil, 200)
	if *after.NutritionPer100["protein_g"] != 4 {
		t.Fatal("incomplete recognition changed food")
	}
	var count int64
	f.db.Model(&model.Food{}).Count(&count)
	if count != 1 {
		t.Fatal("incomplete recognition created another food")
	}
}

func TestFoodImageProviderFailureDoesNotWrite(t *testing.T) {
	f := setup(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(502)
		_, _ = w.Write([]byte(`{"error":"test-secret"}`))
	}))
	defer server.Close()
	chatConfig(t, server.URL)
	body := f.request(t, "POST", "/api/v1/foods/from-image", map[string]any{"image_data_url": "data:image/png;base64,test"}, 502)
	if strings.Contains(string(body), "test-secret") {
		t.Fatal("provider secret leaked")
	}
	var count int64
	f.db.Model(&model.Food{}).Count(&count)
	if count != 0 {
		t.Fatal("failed recognition created food")
	}
}
