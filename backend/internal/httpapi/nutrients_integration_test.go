package httpapi

import (
	"encoding/json"
	"reflect"
	"testing"

	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/nutrition"
	"dailymeal/backend/internal/service"
	"dailymeal/backend/migrations"
	"github.com/pressly/goose/v3"
)

// nutrientCatalog 通过真实接口读取目录，避免测试依赖 Go 中的固定指标列表。
func (f *fixture) nutrientCatalog(t *testing.T) model.Nutrients {
	t.Helper()
	return response[struct {
		Items model.Nutrients `json:"items"`
	}](t, f, "GET", "/api/v1/nutrients", nil, 200).Items
}

// TestIntegrationNutrientRelations 验证目录、关联持久化、零值、清空及省略字段的完整路径。
func TestIntegrationNutrientRelations(t *testing.T) {
	f := setup(t)
	catalog := f.nutrientCatalog(t)
	if len(catalog) != 38 {
		t.Fatalf("got %d seeded nutrients; want 38", len(catalog))
	}
	byCode := map[string]model.Nutrient{}
	for _, n := range catalog {
		byCode[n.Code] = n
	}
	for code, unit := range map[string]string{
		"energy": "kcal", "dietary_fiber": "g", "saturated_fat": "g", "cholesterol": "mg",
		"potassium": "mg", "iron": "mg", "selenium": "ug", "vitamin_b12": "ug",
		"vitamin_a_rae": "ug", "vitamin_e_alpha": "mg", "folate_total": "ug",
	} {
		if n := byCode[code]; n.Unit != unit || n.Name == "" {
			t.Fatalf("missing or incorrect nutrient %s: %+v", code, n)
		}
	}
	food := f.food(t, "燕麦", "g", nutrition.Values{
		"energy_kcal": nutrition.Number(350), "dietary_fiber_g": nutrition.Number(8),
		"potassium_mg": nutrition.Number(200), "selenium_ug": nutrition.Number(0),
	})
	if f.db.Migrator().HasColumn("foods", "nutrition_per_100") {
		t.Fatal("food nutrition is still persisted in the old JSONB column")
	}
	var rows []model.FoodNutrient
	if err := f.db.Where("food_id = ?", food.ID).Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != len(catalog) {
		t.Fatalf("got %d food nutrient rows; want %d", len(rows), len(catalog))
	}
	for _, row := range rows {
		switch row.NutrientID {
		case byCode["dietary_fiber"].ID:
			if row.AmountPer100 == nil || *row.AmountPer100 != 8 {
				t.Fatal("fiber was not persisted in relation")
			}
		case byCode["selenium"].ID:
			if row.AmountPer100 == nil || *row.AmountPer100 != 0 {
				t.Fatal("zero became unknown")
			}
		case byCode["iron"].ID:
			if row.AmountPer100 != nil {
				t.Fatal("unknown iron became known")
			}
		}
	}
	patched := response[model.Food](t, f, "PATCH", path("foods", food.ID), map[string]any{
		"nutrition_per_100": nutrition.Values{"dietary_fiber_g": nutrition.Number(0), "potassium_mg": nil},
	}, 200)
	stored := response[model.Food](t, f, "GET", path("foods", food.ID), nil, 200)
	if !reflect.DeepEqual(patched.NutritionPer100, stored.NutritionPer100) ||
		stored.NutritionPer100["potassium_mg"] != nil || *stored.NutritionPer100["dietary_fiber_g"] != 0 ||
		*stored.NutritionPer100["selenium_ug"] != 0 || *stored.NutritionPer100["energy_kcal"] != 350 {
		t.Fatalf("patch did not preserve null, zero or omitted values: %+v", stored.NutritionPer100)
	}
	list := response[service.List[model.Food]](t, f, "GET", "/api/v1/foods", nil, 200)
	if len(list.Items) != 1 || !reflect.DeepEqual(list.Items[0].NutritionPer100, stored.NutritionPer100) {
		t.Fatal("food list did not load nutrient relations")
	}
}

// TestIntegrationExtendedNutritionSnapshots 验证扩展指标进入菜品及记录，目录变化不改写历史快照。
func TestIntegrationExtendedNutritionSnapshots(t *testing.T) {
	f := setup(t)
	first := f.food(t, "燕麦", "g", nutrition.Values{"energy_kcal": nutrition.Number(350), "dietary_fiber_g": nutrition.Number(8), "potassium_mg": nutrition.Number(200)})
	second := f.food(t, "牛奶", "ml", nutrition.Values{"energy_kcal": nutrition.Number(60), "dietary_fiber_g": nutrition.Number(0)})
	recipe := response[service.RecipeView](t, f, "POST", "/api/v1/recipes", service.RecipeInput{
		Name: "燕麦牛奶", Ingredients: []service.IngredientInput{
			{FoodID: first.ID, Quantity: 0.05, Unit: "kg"}, {FoodID: second.ID, Quantity: 0.2, Unit: "L"},
		},
	}, 201)
	amount(t, recipe.NutritionTotals["dietary_fiber_g"], nutrition.Number(4), true)
	amount(t, recipe.NutritionTotals["potassium_mg"], nutrition.Number(100), false)
	entry := response[model.MealEntry](t, f, "POST", "/api/v1/meal-entries", map[string]any{
		"date": "2026-11-09", "meal_type": "breakfast", "item_type": "recipe", "recipe_id": recipe.ID,
	}, 201)
	foodEntry := response[model.MealEntry](t, f, "POST", "/api/v1/meal-entries", map[string]any{
		"date": "2026-11-09", "meal_type": "lunch", "item_type": "food", "food_id": first.ID, "quantity": 100, "unit": "g",
	}, 201)
	// 通过数据库增加一项定义，证明服务没有另一份写死的计算目录。
	n := model.Nutrient{Code: "beta_carotene", Name: "β-胡萝卜素", Unit: "ug"}
	if err := f.db.Create(&n).Error; err != nil {
		t.Fatal(err)
	}
	withNew := response[model.Food](t, f, "GET", path("foods", first.ID), nil, 200)
	if value, exists := withNew.NutritionPer100[n.Key()]; !exists || value != nil {
		t.Fatal("new catalog nutrient must appear as unknown on existing food")
	}
	f.request(t, "PATCH", path("foods", first.ID), map[string]any{"nutrition_per_100": nutrition.Values{
		"dietary_fiber_g": nutrition.Number(20), n.Key(): nutrition.Number(50),
	}}, 200)
	current := response[service.RecipeView](t, f, "GET", path("recipes", recipe.ID), nil, 200)
	amount(t, current.NutritionTotals["dietary_fiber_g"], nutrition.Number(10), true)
	amount(t, current.NutritionTotals[n.Key()], nutrition.Number(25), false)
	unchanged := response[model.MealEntry](t, f, "GET", path("meal-entries", entry.ID), nil, 200)
	if !reflect.DeepEqual(entry.Snapshot, unchanged.Snapshot) || !reflect.DeepEqual(entry.NutritionTotals, unchanged.NutritionTotals) {
		t.Fatal("food or catalog update rewrote historical recipe entry")
	}
	resized := response[model.MealEntry](t, f, "PATCH", path("meal-entries", foodEntry.ID), map[string]any{"quantity": 200, "unit": "g"}, 200)
	amount(t, resized.NutritionTotals["dietary_fiber_g"], nutrition.Number(16), true)
	if _, exists := resized.Snapshot.NutritionPer100[n.Key()]; exists {
		t.Fatal("quantity update filled a new nutrient into the old snapshot")
	}
	summary := response[service.Summary](t, f, "GET", "/api/v1/daily-summary?date=2026-11-09", nil, 200)
	amount(t, summary.NutritionTotals["dietary_fiber_g"], nutrition.Number(20), true)
	amount(t, summary.NutritionTotals[n.Key()], nil, false)
	empty := response[service.Summary](t, f, "GET", "/api/v1/daily-summary?date=2026-11-10", nil, 200)
	amount(t, empty.NutritionTotals[n.Key()], nutrition.Number(0), true)
}

// TestIntegrationNutrientMigration 验证旧 JSONB 正反向迁移保留数值、未知和历史快照。
func TestIntegrationNutrientMigration(t *testing.T) {
	f := setup(t)
	db, err := f.db.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := goose.DownTo(db, ".", 4); err != nil {
		t.Fatal(err)
	}
	var foodID int64
	if err := f.db.Raw(`INSERT INTO foods (name, basis_unit, nutrition_per_100, favorite)
		VALUES ('旧牛奶', 'ml', '{"energy_kcal":60,"protein_g":3.2,"fat_g":0,"sodium_mg":null}', true) RETURNING id`).Scan(&foodID).Error; err != nil {
		t.Fatal(err)
	}
	snapshot := model.Snapshot{Name: "旧牛奶", BasisUnit: "ml", NutritionPer100: nutrition.Values{
		"energy_kcal": nutrition.Number(60), "protein_g": nutrition.Number(3.2), "fat_g": nutrition.Number(0), "sodium_mg": nil,
	}}
	entry := model.MealEntry{Date: "2026-11-08", MealType: "breakfast", ItemType: "food", FoodID: &foodID,
		Quantity: nutrition.Number(100), Snapshot: snapshot, NutritionTotals: nutrition.Scale(snapshot.NutritionPer100, 100)}
	if err := f.db.Create(&entry).Error; err != nil {
		t.Fatal(err)
	}
	if err := migrations.Run(db, "up"); err != nil {
		t.Fatal(err)
	}
	food := response[model.Food](t, f, "GET", path("foods", foodID), nil, 200)
	if !food.Favorite || food.Name != "旧牛奶" || food.BasisUnit != "ml" {
		t.Fatalf("migration changed food metadata: %+v", food)
	}
	for key, value := range snapshot.NutritionPer100 {
		if !reflect.DeepEqual(food.NutritionPer100[key], value) {
			t.Fatalf("migration changed %s", key)
		}
	}
	if food.NutritionPer100["dietary_fiber_g"] != nil || len(food.NutritionPer100) != 38 {
		t.Fatal("new nutrients must be unknown on migrated foods")
	}
	var historical model.MealEntry
	if err := f.db.First(&historical, entry.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(historical.Snapshot, snapshot) || !reflect.DeepEqual(historical.NutritionTotals, entry.NutritionTotals) {
		t.Fatal("migration rewrote historical nutrition")
	}
	f.request(t, "PATCH", path("foods", foodID), map[string]any{"nutrition_per_100": nutrition.Values{"dietary_fiber_g": nutrition.Number(2)}}, 200)
	if err := goose.DownTo(db, ".", 4); err != nil {
		t.Fatal(err)
	}
	var raw string
	if err := f.db.Raw("SELECT nutrition_per_100::text FROM foods WHERE id = ?", foodID).Scan(&raw).Error; err != nil {
		t.Fatal(err)
	}
	var restored nutrition.Values
	if err := json.Unmarshal([]byte(raw), &restored); err != nil {
		t.Fatal(err)
	}
	if *restored["energy_kcal"] != 60 || *restored["fat_g"] != 0 || restored["sodium_mg"] != nil || *restored["dietary_fiber_g"] != 2 {
		t.Fatalf("rollback lost values: %s", raw)
	}
	if err := migrations.Run(db, "up"); err != nil {
		t.Fatal(err)
	}
}
