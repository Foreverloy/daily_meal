package httpapi

import (
	"encoding/json"
	"net/http"
	"testing"

	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/nutrition"
	"dailymeal/backend/internal/service"
	"dailymeal/backend/migrations"
	"github.com/pressly/goose/v3"
)

// 重复请求现在独立创建，明确验证已取消的去重保障，避免把重复记餐误认为重试成功。
func TestIntegrationRepeatedEntryCreation(t *testing.T) {
	f := setup(t)
	food := f.food(t, "米饭", "g", nutrition.Values{"energy_kcal": nutrition.Number(100)})
	body := map[string]any{"date": "2026-10-12", "meal_type": "lunch", "item_type": "food", "food_id": food.ID, "quantity": 200, "unit": "g"}
	var firstID int64
	for range 2 {
		data := f.request(t, "POST", "/api/v1/meal-entries", body, 201, http.Header{"Idempotency-Key": {"lunch-1"}})
		var entry model.MealEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			t.Fatal(err)
		}
		if entry.ID == 0 || entry.ID == firstID {
			t.Fatalf("request did not create a new entry: %+v", entry)
		}
		firstID = entry.ID
	}
	summary := response[service.Summary](t, f, "GET", "/api/v1/daily-summary?date=2026-10-12", nil, 200)
	amount(t, summary.NutritionTotals["energy_kcal"], nutrition.Number(400), true)
	var requests int64
	if err := f.db.Table("meal_entry_requests").Count(&requests).Error; err != nil || requests != 0 {
		t.Fatalf("legacy request table was written: count=%d err=%v", requests, err)
	}
}

func TestIntegrationIgnoresUnknownFields(t *testing.T) {
	f := setup(t)
	food := response[model.Food](t, f, "POST", "/api/v1/foods", map[string]any{
		"name": "牛奶", "basis_unit": "ml", "nutrition_per_100": map[string]any{"energy_kcal": 60},
		"favorite": true, "unused": "ignored",
	}, 201)
	if food.Favorite || food.Name != "牛奶" {
		t.Fatalf("unknown create fields affected food: %+v", food)
	}
	food = response[model.Food](t, f, "PATCH", path("foods", food.ID), map[string]any{
		"basis_unit": "g", "favorite": true,
	}, 200)
	if food.BasisUnit != "ml" || !food.Favorite {
		t.Fatalf("unknown patch field changed basis: %+v", food)
	}
}

// 历史迁移仍可正常回滚和恢复；当前服务不再使用该表，也不迁移或删除已有业务记录。
func TestIntegrationLegacyRequestMigration(t *testing.T) {
	f := setup(t)
	food := f.food(t, "米饭", "g", nutrition.Values{"energy_kcal": nutrition.Number(100)})
	response[model.MealEntry](t, f, "POST", "/api/v1/meal-entries", map[string]any{
		"meal_type": "lunch", "item_type": "food", "food_id": food.ID, "quantity": 200, "unit": "g",
	}, 201)
	db, err := f.db.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := goose.DownTo(db, ".", 2); err != nil {
		t.Fatal(err)
	}
	if f.db.Migrator().HasTable("meal_entry_requests") {
		t.Fatal("down did not remove the legacy table")
	}
	if err := migrations.Run(db, "up"); err != nil {
		t.Fatal(err)
	}
	if !f.db.Migrator().HasTable("meal_entry_requests") {
		t.Fatal("up did not restore the legacy table")
	}
	var entries int64
	if err := f.db.Model(&model.MealEntry{}).Count(&entries).Error; err != nil || entries != 1 {
		t.Fatalf("migration changed meal entries: count=%d err=%v", entries, err)
	}
}
