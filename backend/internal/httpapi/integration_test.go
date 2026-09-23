package httpapi

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"dailymeal/backend/internal/database"
	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/nutrition"
	"dailymeal/backend/internal/service"
	"dailymeal/backend/internal/store"
	"dailymeal/backend/migrations"
	"dailymeal/backend/openapi"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type fixture struct {
	router   http.Handler
	contract routers.Router
	db       *gorm.DB
	now      time.Time
}

// Tests only accept a dedicated *_test database, then create/drop a random schema.
// No truncation, migration, or writes are performed in public or a personal schema.
func setup(t *testing.T) *fixture {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set; PostgreSQL integration test skipped")
	}
	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || !strings.HasSuffix(u.Path, "_test") {
		t.Fatal("TEST_DATABASE_URL must be a postgres URL to a dedicated database ending in _test")
	}
	admin, err := database.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	adminSQL, err := admin.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = adminSQL.Close() })
	suffix := make([]byte, 8)
	if _, err := rand.Read(suffix); err != nil {
		t.Fatal(err)
	}
	schema := "test_" + hex.EncodeToString(suffix)
	if err := admin.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := admin.Exec("DROP SCHEMA " + schema + " CASCADE").Error; err != nil {
			t.Error(err)
		}
	})
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
	if err := migrations.Run(sqlDB, "up"); err != nil {
		t.Fatal(err)
	}
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	f := &fixture{db: db, now: time.Date(2026, 9, 8, 4, 0, 0, 0, time.UTC)}
	repository := store.New(db)
	svc := service.New(repository, location, func() time.Time { return f.now })
	gin.SetMode(gin.TestMode)
	f.router = New(svc, repository.Ping, openapi.Document)
	document, err := openapi3.NewLoader().LoadFromData(openapi.Document)
	if err != nil {
		t.Fatal(err)
	}
	document.Servers = nil // Match the httptest host, independent of development server URL.
	f.contract, err = legacy.NewRouter(document)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *fixture) request(t *testing.T, method, path string, body any, status int) []byte {
	t.Helper()
	var data []byte
	if body != nil {
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	r := httptest.NewRequest(method, path, bytes.NewReader(data))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, r)
	if w.Code != status {
		t.Fatalf("%s %s: status=%d want=%d body=%s", method, path, w.Code, status, w.Body.String())
	}
	if status == 204 && w.Body.Len() != 0 {
		t.Fatal("204 response has a body")
	}
	contractRequest := httptest.NewRequest(method, path, bytes.NewReader(data))
	contractRequest.Header.Set("Content-Type", "application/json")
	route, params, err := f.contract.FindRoute(contractRequest)
	if err != nil {
		t.Fatal(err)
	}
	input := &openapi3filter.RequestValidationInput{Request: contractRequest, PathParams: params, Route: route}
	if status < 400 {
		if err := openapi3filter.ValidateRequest(context.Background(), input); err != nil {
			t.Fatalf("request violates OpenAPI: %v", err)
		}
	}
	output := &openapi3filter.ResponseValidationInput{RequestValidationInput: input, Status: status, Header: w.Header()}
	if err := openapi3filter.ValidateResponse(context.Background(), output.SetBodyBytes(w.Body.Bytes())); err != nil {
		t.Fatalf("response violates OpenAPI: %v", err)
	}
	return w.Body.Bytes()
}

func response[T any](t *testing.T, f *fixture, method, path string, body any, status int) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(f.request(t, method, path, body, status), &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func goalInput(energy float64) map[string]any {
	return map[string]any{"energy_kcal": energy, "protein_percent": 30, "carbohydrate_percent": 40, "fat_percent": 30}
}

func (f *fixture) food(t *testing.T, name, basis string, values nutrition.Values) model.Food {
	t.Helper()
	return response[model.Food](t, f, "POST", "/api/v1/foods", map[string]any{"name": name, "basis_unit": basis, "nutrition_per_100": values}, 201)
}

func path(resource string, id int64) string { return fmt.Sprintf("/api/v1/%s/%d", resource, id) }

func amount(t *testing.T, got nutrition.Amount, want *float64, complete bool) {
	t.Helper()
	if got.Complete != complete || (got.KnownTotal == nil) != (want == nil) {
		t.Fatalf("got %+v want %v/%v", got, want, complete)
	}
	if want != nil && math.Abs(*got.KnownTotal-*want) > 1e-8 {
		t.Fatalf("got %.12f want %.12f", *got.KnownTotal, *want)
	}
}

func TestIntegrationMigrations(t *testing.T) {
	f := setup(t)
	db, err := f.db.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := migrations.Run(db, "up"); err != nil {
		t.Fatal(err)
	}
	if err := migrations.Run(db, "down"); err != nil {
		t.Fatal(err)
	}
	if f.db.Migrator().HasTable(&model.Food{}) {
		t.Fatal("down did not remove business tables")
	}
	if err := migrations.Run(db, "up"); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"goal_settings", "daily_goals", "foods", "recipes", "recipe_ingredients", "meal_entries"} {
		if !f.db.Migrator().HasTable(table) {
			t.Fatalf("missing table %s", table)
		}
	}
	f.request(t, "GET", "/healthz", nil, 200)
}

func TestIntegrationGoalHistory(t *testing.T) {
	f := setup(t)
	if got := string(f.request(t, "GET", "/api/v1/goal-settings/default", nil, 200)); got != "null" {
		t.Fatalf("expected null, got %s", got)
	}
	for _, endpoint := range []string{"/api/v1/daily-goals/2026-09-07", "/api/v1/daily-goals/2026-09-08"} {
		goal := response[service.ResolvedGoal](t, f, "GET", endpoint, nil, 200)
		if goal.Source != "none" || goal.Goal != nil {
			t.Fatalf("unexpected goal: %+v", goal)
		}
	}
	f.request(t, "PUT", "/api/v1/goal-settings/default", goalInput(2000), 200)
	f.request(t, "PUT", "/api/v1/goal-settings/default", goalInput(2100), 200)
	var count int64
	if err := f.db.Model(&model.GoalSetting{}).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("same-day upsert: %d, %v", count, err)
	}
	f.request(t, "PUT", "/api/v1/daily-goals/2026-09-08", goalInput(1800), 200)
	f.now = time.Date(2026, 9, 8, 16, 1, 0, 0, time.UTC) // Already September 9 in Shanghai.
	current := response[service.GoalView](t, f, "PUT", "/api/v1/goal-settings/default", goalInput(2200), 200)
	if current.EffectiveOn == nil || *current.EffectiveOn != "2026-09-09" {
		t.Fatalf("wrong business date: %+v", current)
	}
	for _, tc := range []struct {
		date, source string
		energy       float64
	}{{"2026-09-07", "none", 0}, {"2026-09-08", "override", 1800}, {"2026-09-09", "default", 2200}, {"2027-01-01", "default", 2200}} {
		got := response[service.ResolvedGoal](t, f, "GET", "/api/v1/daily-goals/"+tc.date, nil, 200)
		if got.Source != tc.source || (got.Goal != nil && got.Goal.EnergyKcal != tc.energy) {
			t.Fatalf("date %s: %+v", tc.date, got)
		}
	}
	f.request(t, "DELETE", "/api/v1/daily-goals/2026-09-08", nil, 204)
	restored := response[service.ResolvedGoal](t, f, "GET", "/api/v1/daily-goals/2026-09-08", nil, 200)
	if restored.Source != "default" || restored.Goal.EnergyKcal != 2100 {
		t.Fatalf("wrong historical default: %+v", restored)
	}
	f.request(t, "PUT", "/api/v1/daily-goals/2026-09-01", goalInput(1600), 200)
	f.request(t, "DELETE", "/api/v1/daily-goals/2026-09-01", nil, 204)
	f.request(t, "DELETE", "/api/v1/daily-goals/2026-09-01", nil, 204)
	none := response[service.ResolvedGoal](t, f, "GET", "/api/v1/daily-goals/2026-09-01", nil, 200)
	if none.Goal != nil {
		t.Fatal("future default leaked into past")
	}
	invalid := goalInput(2000)
	invalid["fat_percent"] = 20
	f.request(t, "PUT", "/api/v1/goal-settings/default", invalid, 422)
	f.request(t, "GET", "/api/v1/daily-goals/2026-02-29", nil, 422)
}

func TestIntegrationFoodsAndRecipes(t *testing.T) {
	f := setup(t)
	rice := f.food(t, "熟米饭", "g", nutrition.Values{"energy_kcal": nutrition.Number(100), "protein_g": nutrition.Number(2), "sodium_mg": nutrition.Number(0)})
	milk := f.food(t, "牛奶", "ml", nutrition.Values{"energy_kcal": nutrition.Number(60), "protein_g": nil, "sodium_mg": nil})
	if rice.NutritionPer100["sodium_mg"] == nil || *rice.NutritionPer100["sodium_mg"] != 0 || rice.NutritionPer100["fat_g"] != nil || len(rice.NutritionPer100) != len(nutrition.Keys) {
		t.Fatalf("wrong nutrition: %v", rice.NutritionPer100)
	}
	patched := response[model.Food](t, f, "PATCH", path("foods", rice.ID), map[string]any{"nutrition_per_100": map[string]any{"protein_g": nil, "fat_g": 0}}, 200)
	if patched.NutritionPer100["protein_g"] != nil || patched.NutritionPer100["fat_g"] == nil || *patched.NutritionPer100["energy_kcal"] != 100 {
		t.Fatal("partial nutrition patch lost fields")
	}
	f.request(t, "PATCH", path("foods", rice.ID), map[string]any{"nutrition_per_100": map[string]any{"energy_kcal": nil}}, 422)
	f.request(t, "PATCH", path("foods", rice.ID), map[string]any{"nutrition_per_100": map[string]any{"fiber_g": 1}}, 422)
	input := service.RecipeInput{Name: "米饭牛奶", Ingredients: []service.IngredientInput{{FoodID: rice.ID, Quantity: .15, Unit: "kg"}, {FoodID: milk.ID, Quantity: .25, Unit: "L"}}}
	recipe := response[service.RecipeView](t, f, "POST", "/api/v1/recipes", input, 201)
	if recipe.Ingredients[0].Quantity != 150 || recipe.Ingredients[1].Quantity != 250 {
		t.Fatal("units were not normalized")
	}
	amount(t, recipe.NutritionTotals["energy_kcal"], nutrition.Number(300), true)
	amount(t, recipe.NutritionTotals["protein_g"], nil, false)
	amount(t, recipe.NutritionTotals["sodium_mg"], nutrition.Number(0), false)
	for _, invalid := range []service.RecipeInput{
		{Name: "empty"},
		{Name: "invalid reference", Ingredients: []service.IngredientInput{{FoodID: rice.ID, Quantity: 10, Unit: "g"}, {FoodID: 99999, Quantity: 100, Unit: "g"}}},
		{Name: "cross dimension", Ingredients: []service.IngredientInput{{FoodID: rice.ID, Quantity: 10, Unit: "ml"}}},
		{Name: "invalid quantity", Ingredients: []service.IngredientInput{{FoodID: milk.ID, Quantity: 0, Unit: "ml"}}},
	} {
		f.request(t, "PUT", path("recipes", recipe.ID), invalid, 422)
	}
	unchanged := response[service.RecipeView](t, f, "GET", path("recipes", recipe.ID), nil, 200)
	if unchanged.Name != recipe.Name || len(unchanged.Ingredients) != 2 {
		t.Fatal("failed update partially replaced recipe")
	}
	f.request(t, "PATCH", path("foods", milk.ID), map[string]any{"nutrition_per_100": map[string]any{"energy_kcal": 80, "protein_g": 3}}, 200)
	current := response[service.RecipeView](t, f, "GET", path("recipes", recipe.ID), nil, 200)
	amount(t, current.NutritionTotals["energy_kcal"], nutrition.Number(350), true)
	amount(t, current.NutritionTotals["protein_g"], nutrition.Number(7.5), false)
	foods := response[service.List[model.Food]](t, f, "GET", "/api/v1/foods?q="+url.QueryEscape("牛奶")+"&page_size=1", nil, 200)
	if foods.Total != 1 || len(foods.Items) != 1 || foods.Items[0].ID != milk.ID || foods.PageSize != 1 {
		t.Fatalf("wrong search: %+v", foods)
	}
	for _, query := range []string{"%", "_"} {
		list := response[service.List[model.Food]](t, f, "GET", "/api/v1/foods?q="+url.QueryEscape(query), nil, 200)
		if list.Total != 0 {
			t.Fatal("search wildcards were not escaped")
		}
	}
	list := response[service.List[service.RecipeView]](t, f, "GET", "/api/v1/recipes?q="+url.QueryEscape("米饭"), nil, 200)
	if list.Total != 1 || len(list.Items) != 1 {
		t.Fatal("recipe list is wrong")
	}
	f.request(t, "GET", "/api/v1/foods?page_size=101", nil, 422)
	f.request(t, "GET", "/api/v1/recipes?page=0", nil, 422)
	f.request(t, "GET", "/api/v1/foods/99999", nil, 404)
}

func TestIntegrationSnapshotsAndSummary(t *testing.T) {
	f := setup(t)
	f.request(t, "PUT", "/api/v1/goal-settings/default", goalInput(2000), 200)
	rice := f.food(t, "米饭", "g", nutrition.Values{"energy_kcal": nutrition.Number(100), "protein_g": nutrition.Number(2), "carbohydrate_g": nutrition.Number(20), "fat_g": nutrition.Number(0)})
	milk := f.food(t, "牛奶", "ml", nutrition.Values{"energy_kcal": nutrition.Number(60), "protein_g": nil, "fat_g": nutrition.Number(3)})
	recipe := response[service.RecipeView](t, f, "POST", "/api/v1/recipes", service.RecipeInput{Name: "牛奶饭", Ingredients: []service.IngredientInput{{FoodID: rice.ID, Quantity: 100, Unit: "g"}, {FoodID: milk.ID, Quantity: 200, Unit: "ml"}}}, 201)
	foodEntry := response[model.MealEntry](t, f, "POST", "/api/v1/meal-entries", map[string]any{"meal_type": "lunch", "item_type": "food", "food_id": rice.ID, "quantity": .15, "unit": "kg"}, 201)
	recipeEntry := response[model.MealEntry](t, f, "POST", "/api/v1/meal-entries", map[string]any{"meal_type": "lunch", "item_type": "recipe", "recipe_id": recipe.ID}, 201)
	if recipeEntry.Quantity != nil || recipeEntry.FoodID != nil || len(recipeEntry.Snapshot.Ingredients) != 2 {
		t.Fatal("recipe entry did not capture entire recipe")
	}
	if foodEntry.Date != "2026-09-08" || *foodEntry.Quantity != 150 {
		t.Fatal("wrong default date or quantity")
	}
	summary := response[service.Summary](t, f, "GET", "/api/v1/daily-summary?date=2026-09-08", nil, 200)
	amount(t, summary.NutritionTotals["energy_kcal"], nutrition.Number(370), true)
	amount(t, summary.NutritionTotals["protein_g"], nutrition.Number(5), false)
	amount(t, summary.NutritionTotals["sodium_mg"], nil, false)
	amount(t, summary.Meals["breakfast"]["sodium_mg"], nutrition.Number(0), true)
	if summary.Comparison["protein_g"].Remaining != nil || *summary.Comparison["energy_kcal"].Remaining != 1630 {
		t.Fatal("incorrect remaining allowance")
	}
	f.request(t, "PATCH", path("foods", rice.ID), map[string]any{"name": "新米饭", "nutrition_per_100": map[string]any{"energy_kcal": 200, "protein_g": 4}}, 200)
	f.request(t, "PUT", path("recipes", recipe.ID), service.RecipeInput{Name: "新配方", Ingredients: []service.IngredientInput{{FoodID: rice.ID, Quantity: 500, Unit: "g"}}}, 200)
	old := response[model.MealEntry](t, f, "GET", path("meal-entries", recipeEntry.ID), nil, 200)
	if old.Snapshot.Name != "牛奶饭" || len(old.Snapshot.Ingredients) != 2 {
		t.Fatal("historical recipe changed")
	}
	amount(t, old.NutritionTotals["energy_kcal"], nutrition.Number(220), true)
	edited := response[model.MealEntry](t, f, "PATCH", path("meal-entries", foodEntry.ID), map[string]any{"quantity": 200}, 200)
	amount(t, edited.NutritionTotals["energy_kcal"], nutrition.Number(200), true)
	if edited.Snapshot.Name != "米饭" || *edited.Snapshot.NutritionPer100["energy_kcal"] != 100 {
		t.Fatal("quantity edit refreshed snapshot")
	}
	moved := response[model.MealEntry](t, f, "PATCH", path("meal-entries", recipeEntry.ID), map[string]any{"date": "2026-09-07", "meal_type": "dinner"}, 200)
	amount(t, moved.NutritionTotals["energy_kcal"], nutrition.Number(220), true)
	summary = response[service.Summary](t, f, "GET", "/api/v1/daily-summary?date=2026-09-08", nil, 200)
	amount(t, summary.NutritionTotals["energy_kcal"], nutrition.Number(200), true)
	yesterday := response[service.Summary](t, f, "GET", "/api/v1/daily-summary?date=2026-09-07", nil, 200)
	if yesterday.Goal.Goal != nil || yesterday.Comparison["energy_kcal"].Remaining != nil {
		t.Fatal("missing target produced allowance")
	}
	amount(t, yesterday.Meals["dinner"]["energy_kcal"], nutrition.Number(220), true)
	newEntry := response[model.MealEntry](t, f, "POST", "/api/v1/meal-entries", map[string]any{"date": "2026-09-08", "meal_type": "breakfast", "item_type": "recipe", "recipe_id": recipe.ID}, 201)
	amount(t, newEntry.NutritionTotals["energy_kcal"], nutrition.Number(1000), true)
	switched := response[model.MealEntry](t, f, "PATCH", path("meal-entries", foodEntry.ID), map[string]any{"food_id": milk.ID, "quantity": .25, "unit": "L"}, 200)
	if switched.Snapshot.Name != "牛奶" || *switched.Quantity != 250 {
		t.Fatal("new food not snapshotted")
	}
	amount(t, switched.NutritionTotals["energy_kcal"], nutrition.Number(150), true)
	asRecipe := response[model.MealEntry](t, f, "PATCH", path("meal-entries", foodEntry.ID), map[string]any{"item_type": "recipe", "recipe_id": recipe.ID}, 200)
	if asRecipe.FoodID != nil || asRecipe.Quantity != nil {
		t.Fatal("cross type switch did not clear food fields")
	}
	asFood := response[model.MealEntry](t, f, "PATCH", path("meal-entries", foodEntry.ID), map[string]any{"item_type": "food", "food_id": rice.ID, "quantity": 2000, "unit": "g"}, 200)
	if asFood.RecipeID != nil {
		t.Fatal("cross type switch did not clear recipe ID")
	}
	summary = response[service.Summary](t, f, "GET", "/api/v1/daily-summary?date=2026-09-08", nil, 200)
	if *summary.Comparison["energy_kcal"].Remaining != -3000 {
		t.Fatalf("negative remaining lost: %+v", summary.Comparison["energy_kcal"])
	}
	for _, id := range []int64{foodEntry.ID, newEntry.ID} {
		f.request(t, "DELETE", path("meal-entries", id), nil, 204)
	}
	f.request(t, "GET", path("meal-entries", foodEntry.ID), nil, 404)
	empty := response[service.Summary](t, f, "GET", "/api/v1/daily-summary?date=2026-09-08", nil, 200)
	for _, k := range nutrition.Keys {
		amount(t, empty.NutritionTotals[k], nutrition.Number(0), true)
	}
	entries := response[struct {
		Items []model.MealEntry `json:"items"`
	}](t, f, "GET", "/api/v1/meal-entries?date=2026-09-07&meal_type=dinner", nil, 200)
	if len(entries.Items) != 1 || entries.Items[0].ID != recipeEntry.ID {
		t.Fatal("wrong date/meal filter")
	}
}

func TestIntegrationEntryValidationAndRollback(t *testing.T) {
	f := setup(t)
	food := f.food(t, "食物", "g", nutrition.Values{"energy_kcal": nutrition.Number(100)})
	for _, body := range []map[string]any{
		{"meal_type": "snack", "item_type": "food", "food_id": food.ID, "quantity": 100, "unit": "g"},
		{"meal_type": "lunch", "item_type": "food", "food_id": food.ID, "quantity": 100, "unit": "ml"},
		{"meal_type": "lunch", "item_type": "food", "food_id": 99999, "quantity": 100, "unit": "g"},
		{"meal_type": "lunch", "item_type": "recipe", "recipe_id": 99999},
		{"meal_type": "lunch", "item_type": "recipe", "recipe_id": 1, "quantity": 1},
		{"meal_type": "lunch", "item_type": "food", "food_id": food.ID, "quantity": -1, "unit": "g"},
		{"meal_type": "lunch", "item_type": "food", "food_id": food.ID},
		{"meal_type": "lunch", "item_type": "food", "food_id": food.ID, "quantity": 100, "unit": "g", "recipe_id": 1},
		{"meal_type": "lunch", "item_type": "food", "food_id": food.ID, "quantity": 100, "unit": "g", "date": nil},
	} {
		f.request(t, "POST", "/api/v1/meal-entries", body, 422)
	}
	var count int64
	if err := f.db.Model(&model.MealEntry{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatal("invalid writes persisted")
	}
	entry := response[model.MealEntry](t, f, "POST", "/api/v1/meal-entries", map[string]any{"meal_type": "breakfast", "item_type": "food", "food_id": food.ID, "quantity": 100, "unit": "g"}, 201)
	f.request(t, "PATCH", path("meal-entries", entry.ID), map[string]any{"date": "2026-09-07", "unit": "ml"}, 422)
	preserved := response[model.MealEntry](t, f, "GET", path("meal-entries", entry.ID), nil, 200)
	if preserved.Date != entry.Date {
		t.Fatal("failed patch persisted metadata")
	}
	f.request(t, "GET", "/api/v1/meal-entries?date=2026-09-08&meal_type=snack", nil, 422)
	f.request(t, "GET", "/api/v1/daily-summary", nil, 422)
}

func TestIntegrationRecipeRollbackAfterWrite(t *testing.T) {
	f := setup(t)
	food := f.food(t, "正常食材", "g", nutrition.Values{"energy_kcal": nutrition.Number(100)})
	huge := f.food(t, "超出计算范围的测试数据", "g", nutrition.Values{"energy_kcal": nutrition.Number(1e308)})
	recipe := response[service.RecipeView](t, f, "POST", "/api/v1/recipes", service.RecipeInput{Name: "原配方", Ingredients: []service.IngredientInput{{FoodID: food.ID, Quantity: 100, Unit: "g"}}}, 201)
	// Overflow is detected while calculating the response, after the replacement rows were written.
	invalid := service.RecipeInput{Name: "不能保存", Ingredients: []service.IngredientInput{{FoodID: huge.ID, Quantity: 1000, Unit: "g"}}}
	f.request(t, "PUT", path("recipes", recipe.ID), invalid, 422)
	f.request(t, "POST", "/api/v1/recipes", invalid, 422)
	preserved := response[service.RecipeView](t, f, "GET", path("recipes", recipe.ID), nil, 200)
	if preserved.Name != "原配方" || len(preserved.Ingredients) != 1 || preserved.Ingredients[0].FoodID != food.ID {
		t.Fatal("transaction did not restore the original recipe")
	}
	var count int64
	if err := f.db.Model(&model.Recipe{}).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("failed creation persisted: %d, %v", count, err)
	}
}
