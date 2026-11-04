package httpapi

import (
	"net/url"
	"os"
	"reflect"
	"testing"

	"dailymeal/backend/internal/database"
	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/nutrition"
	"dailymeal/backend/internal/service"
	"dailymeal/backend/internal/store"
	"dailymeal/backend/migrations"
	"dailymeal/backend/openapi"
	"github.com/pressly/goose/v3"
)

func TestIntegrationFoodFavorites(t *testing.T) {
	f := setup(t)
	milkA := f.food(t, "牛奶 A", "ml", nutrition.Values{"energy_kcal": nutrition.Number(60)})
	milkB := f.food(t, "牛奶 B", "ml", nutrition.Values{"energy_kcal": nutrition.Number(65)})
	milkC := f.food(t, "牛奶 C", "ml", nutrition.Values{"energy_kcal": nutrition.Number(70)})
	rice := f.food(t, "米饭", "g", nutrition.Values{"energy_kcal": nutrition.Number(100)})
	if milkA.Favorite || milkB.Favorite || milkC.Favorite || rice.Favorite {
		t.Fatal("new foods must default to unfavorited")
	}
	empty := response[service.List[model.Food]](t, f, "GET", "/api/v1/foods?favorite=true", nil, 200)
	if empty.Total != 0 || empty.Items == nil || len(empty.Items) != 0 {
		t.Fatalf("expected an empty favorites list, got %+v", empty)
	}

	// 重复设置验证状态幂等；B 保持未收藏，米饭验证名称条件仍会排除它。
	for _, food := range []model.Food{milkA, milkC, rice} {
		for i := 0; i < 2; i++ {
			got := response[model.Food](t, f, "PATCH", path("foods", food.ID), map[string]any{"favorite": true}, 200)
			if !got.Favorite {
				t.Fatalf("food %d was not favorited", food.ID)
			}
		}
	}
	search := "q=" + url.QueryEscape("牛奶")
	for _, tc := range []struct {
		name, query string
		total       int64
		ids         []int64
	}{
		{"all", "", 4, []int64{rice.ID, milkC.ID, milkB.ID, milkA.ID}},
		{"false means all", "favorite=false", 4, []int64{rice.ID, milkC.ID, milkB.ID, milkA.ID}},
		{"favorites", "favorite=true", 3, []int64{rice.ID, milkC.ID, milkA.ID}},
		{"search all", search, 3, []int64{milkC.ID, milkB.ID, milkA.ID}},
		{"false preserves search", "favorite=false&" + search, 3, []int64{milkC.ID, milkB.ID, milkA.ID}},
		{"search favorites", "favorite=true&" + search, 2, []int64{milkC.ID, milkA.ID}},
		{"first page", "favorite=true&" + search + "&page_size=1", 2, []int64{milkC.ID}},
		{"second page", "favorite=true&" + search + "&page_size=1&page=2", 2, []int64{milkA.ID}},
		{"past last page", "favorite=true&" + search + "&page_size=1&page=3", 2, nil},
		{"no match", "favorite=true&q=missing", 0, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := response[service.List[model.Food]](t, f, "GET", "/api/v1/foods?"+tc.query, nil, 200)
			if got.Total != tc.total || got.Items == nil || len(got.Items) != len(tc.ids) {
				t.Fatalf("got %+v want total=%d ids=%v", got, tc.total, tc.ids)
			}
			for i, id := range tc.ids {
				if got.Items[i].ID != id || got.Items[i].Favorite != (id != milkB.ID) {
					t.Fatalf("item %d: got id=%d favorite=%v want id=%d favorite=%v", i, got.Items[i].ID, got.Items[i].Favorite, id, id != milkB.ID)
				}
			}
		})
	}

	got := response[model.Food](t, f, "PATCH", path("foods", milkA.ID), map[string]any{
		"name": "新牛奶", "nutrition_per_100": map[string]any{"energy_kcal": 80},
	}, 200)
	if !got.Favorite || got.Name != "新牛奶" || *got.NutritionPer100["energy_kcal"] != 80 {
		t.Fatal("editing name and nutrition did not preserve favorite")
	}
	for i := 0; i < 2; i++ {
		got = response[model.Food](t, f, "PATCH", path("foods", milkA.ID), map[string]any{"favorite": false}, 200)
		if got.Favorite {
			t.Fatal("repeated unfavorite did not preserve false")
		}
	}
	got = response[model.Food](t, f, "GET", path("foods", milkA.ID), nil, 200)
	if got.Favorite {
		t.Fatal("detail did not reflect the saved unfavorite")
	}
	remaining := response[service.List[model.Food]](t, f, "GET", "/api/v1/foods?favorite=true&"+search, nil, 200)
	if remaining.Total != 1 || len(remaining.Items) != 1 || remaining.Items[0].ID != milkC.ID {
		t.Fatal("unfavorited food remained in the favorites query")
	}

	// 关闭原连接并重建 Service/Router，验证状态来自 PostgreSQL 而非服务内存。
	var schema string
	if err := f.db.Raw("SELECT current_schema()").Scan(&schema).Error; err != nil {
		t.Fatal(err)
	}
	oldDB, err := f.db.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := oldDB.Close(); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	query := u.Query()
	query.Set("search_path", schema)
	u.RawQuery = query.Encode()
	db, err := database.Open(u.String())
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	repository := store.New(db)
	f.db = db
	f.router = New(service.New(repository), repository.Ping, openapi.Document)
	for _, tc := range []struct {
		id       int64
		favorite bool
	}{{milkA.ID, false}, {milkB.ID, false}, {milkC.ID, true}} {
		got := response[model.Food](t, f, "GET", path("foods", tc.id), nil, 200)
		if got.Favorite != tc.favorite {
			t.Fatalf("food %d lost favorite after service reconstruction", tc.id)
		}
	}
}

func TestIntegrationFoodFavoriteErrors(t *testing.T) {
	f := setup(t)
	food := f.food(t, "牛奶", "ml", nutrition.Values{"energy_kcal": nutrition.Number(60)})
	endpoint := path("foods", food.ID)
	f.request(t, "PATCH", endpoint, map[string]any{"favorite": true}, 200)
	before := response[model.Food](t, f, "GET", endpoint, nil, 200)
	for _, tc := range []struct {
		name   string
		body   map[string]any
		status int
	}{
		{"string", map[string]any{"favorite": "false"}, 400},
		{"number", map[string]any{"favorite": 0}, 400},
		{"database rejects name", map[string]any{"name": "", "favorite": false}, 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f.request(t, "PATCH", endpoint, tc.body, tc.status)
			got := response[model.Food](t, f, "GET", endpoint, nil, 200)
			if !reflect.DeepEqual(got, before) {
				t.Fatalf("failed patch persisted changes: got %+v want %+v", got, before)
			}
		})
	}
	for _, favorite := range []bool{true, false} {
		f.request(t, "PATCH", "/api/v1/foods/99999", map[string]any{"favorite": favorite}, 404)
	}
	for _, value := range []string{"invalid", "null", ""} {
		f.request(t, "GET", "/api/v1/foods?favorite="+value, nil, 400)
	}
	// 菜品列表沿用原规则，忽略与其无关的查询参数。
	f.request(t, "GET", "/api/v1/recipes?favorite=invalid", nil, 200)
	// 名称与收藏状态由一条 UPDATE 保存，数据库失败时直接返回错误。
	if err := f.db.Exec("ALTER TABLE foods ADD CONSTRAINT test_favorite_write_failure CHECK (favorite)").Error; err != nil {
		t.Fatal(err)
	}
	f.request(t, "PATCH", endpoint, map[string]any{"name": "不能保存", "favorite": false}, 500)
	got := response[model.Food](t, f, "GET", endpoint, nil, 200)
	if !reflect.DeepEqual(got, before) {
		t.Fatal("database write failure persisted a partial patch")
	}
}

func TestIntegrationFoodFavoritesDoNotChangeMealsOrRecipes(t *testing.T) {
	f := setup(t)
	food := f.food(t, "米饭", "g", nutrition.Values{"energy_kcal": nutrition.Number(100), "protein_g": nutrition.Number(2)})
	recipe := response[service.RecipeView](t, f, "POST", "/api/v1/recipes", service.RecipeInput{
		Name: "米饭菜品", Ingredients: []service.IngredientInput{{FoodID: food.ID, Quantity: 150, Unit: "g"}},
	}, 201)
	foodEntry := response[model.MealEntry](t, f, "POST", "/api/v1/meal-entries", map[string]any{
		"date": "2026-09-08", "meal_type": "lunch", "item_type": "food", "food_id": food.ID, "quantity": 100, "unit": "g",
	}, 201)
	recipeEntry := response[model.MealEntry](t, f, "POST", "/api/v1/meal-entries", map[string]any{
		"date": "2026-09-08", "meal_type": "dinner", "item_type": "recipe", "recipe_id": recipe.ID,
	}, 201)
	// 先重新查询已落库的数据，避免创建响应与数据库时间戳精度的差异。
	beforeFood := response[model.Food](t, f, "GET", path("foods", food.ID), nil, 200)
	beforeRecipe := response[service.RecipeView](t, f, "GET", path("recipes", recipe.ID), nil, 200)
	beforeEntries := []model.MealEntry{
		response[model.MealEntry](t, f, "GET", path("meal-entries", foodEntry.ID), nil, 200),
		response[model.MealEntry](t, f, "GET", path("meal-entries", recipeEntry.ID), nil, 200),
	}
	beforeSummary := response[service.Summary](t, f, "GET", "/api/v1/daily-summary?date=2026-09-08", nil, 200)
	for _, favorite := range []bool{true, false} {
		got := response[model.Food](t, f, "PATCH", path("foods", food.ID), map[string]any{"favorite": favorite}, 200)
		if got.Favorite != favorite || got.Name != beforeFood.Name || got.BasisUnit != beforeFood.BasisUnit || !reflect.DeepEqual(got.NutritionPer100, beforeFood.NutritionPer100) {
			t.Fatal("favorite operation changed food nutrition or identity")
		}
		current := response[service.RecipeView](t, f, "GET", path("recipes", recipe.ID), nil, 200)
		if !reflect.DeepEqual(current, beforeRecipe) {
			t.Fatal("favorite operation changed recipe composition or nutrition")
		}
		for _, before := range beforeEntries {
			current := response[model.MealEntry](t, f, "GET", path("meal-entries", before.ID), nil, 200)
			if !reflect.DeepEqual(current, before) {
				t.Fatal("favorite operation changed a historical meal entry")
			}
		}
		currentSummary := response[service.Summary](t, f, "GET", "/api/v1/daily-summary?date=2026-09-08", nil, 200)
		if !reflect.DeepEqual(currentSummary, beforeSummary) {
			t.Fatal("favorite operation changed the nutrition summary")
		}
		var count int64
		if err := f.db.Model(&model.MealEntry{}).Count(&count).Error; err != nil || count != 2 {
			t.Fatalf("favorite operation created meal entries: count=%d err=%v", count, err)
		}
	}
}

func TestIntegrationFoodFavoritesMigration(t *testing.T) {
	f := setup(t)
	db, err := f.db.DB()
	if err != nil {
		t.Fatal(err)
	}
	// 明确回到收藏功能之前的版本，后续新增迁移不会改变本测试的起点。
	if err := goose.DownTo(db, ".", 1); err != nil {
		t.Fatal(err)
	}
	if !f.db.Migrator().HasTable("foods") || f.db.Migrator().HasColumn("foods", "favorite") {
		t.Fatal("rolling back favorites must retain foods and remove only the favorite column")
	}
	legacy := model.Food{Name: "迁移前的米饭", BasisUnit: "g", NutritionPer100: nutrition.Values{"energy_kcal": nutrition.Number(100), "protein_g": nutrition.Number(2)}.Normalized()}
	if err := f.db.Omit("Favorite").Create(&legacy).Error; err != nil {
		t.Fatal(err)
	}
	var before model.Food
	// 旧版本只读取当时存在的列，避免跨 schema 变更复用 SELECT * 的缓存计划。
	if err := f.db.Select("id", "name", "basis_unit", "nutrition_per_100", "created_at", "updated_at").First(&before, legacy.ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := migrations.Run(db, "up"); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := f.db.Model(&model.Food{}).Where("id = ? AND favorite IS FALSE", legacy.ID).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("existing food did not receive database default false: count=%d err=%v", count, err)
	}
	got := response[model.Food](t, f, "GET", path("foods", legacy.ID), nil, 200)
	if got.Favorite || !reflect.DeepEqual(got, before) {
		t.Fatalf("migration changed existing food data: got %+v want %+v", got, before)
	}
	// 迁移后能正常收藏，避免只验证列存在却遗漏 ORM 列名不一致。
	got = response[model.Food](t, f, "PATCH", path("foods", legacy.ID), map[string]any{"favorite": true}, 200)
	if !got.Favorite {
		t.Fatal("migrated food could not be favorited")
	}
}
