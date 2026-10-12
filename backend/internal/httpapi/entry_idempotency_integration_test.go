package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/nutrition"
	"dailymeal/backend/internal/service"
	"dailymeal/backend/internal/store"
	"dailymeal/backend/migrations"
	"dailymeal/backend/openapi"
	"github.com/pressly/goose/v3"
)

func keyedEntry(t *testing.T, f *fixture, key string, body any, status int) []byte {
	t.Helper()
	return f.request(t, "POST", "/api/v1/meal-entries", body, status, http.Header{"Idempotency-Key": {key}})
}

func entryRequestCounts(t *testing.T, f *fixture, entries, requests int64) {
	t.Helper()
	for table, want := range map[string]int64{"meal_entries": entries, "meal_entry_requests": requests} {
		var got int64
		if err := f.db.Table(table).Count(&got).Error; err != nil || got != want {
			t.Fatalf("%s count=%d want=%d err=%v", table, got, want, err)
		}
	}
}

func TestIntegrationEntryIdempotency(t *testing.T) {
	f := setup(t)
	food := f.food(t, "米饭", "g", nutrition.Values{"energy_kcal": nutrition.Number(100)})
	body := map[string]any{"date": "2026-10-12", "meal_type": "lunch", "item_type": "food", "food_id": food.ID, "quantity": 200, "unit": "g"}
	first := keyedEntry(t, f, "lunch-1", body, 201)
	var entry model.MealEntry
	if err := json.Unmarshal(first, &entry); err != nil {
		t.Fatal(err)
	}
	// Field order, whitespace and 200.0 versus 200 preserve the parsed request.
	reordered := json.RawMessage(fmt.Sprintf(`{"unit":"g", "quantity":200.0,"food_id":%d,"item_type":"food","meal_type":"lunch","date":"2026-10-12"}`, food.ID))
	if got := keyedEntry(t, f, "lunch-1", reordered, 201); !bytes.Equal(got, first) {
		t.Fatalf("replay changed response: %s want %s", got, first)
	}
	body["quantity"] = 300
	keyedEntry(t, f, "lunch-1", body, 409)
	body["quantity"] = 200
	keyedEntry(t, f, "lunch-2", body, 201)
	// Existing callers can continue creating without a key.
	f.request(t, "POST", "/api/v1/meal-entries", body, 201)
	f.request(t, "POST", "/api/v1/meal-entries", body, 201)
	entryRequestCounts(t, f, 4, 2)
	response[model.Food](t, f, "PATCH", path("foods", food.ID), map[string]any{"nutrition_per_100": map[string]any{"energy_kcal": 999}}, 200)
	response[model.MealEntry](t, f, "PATCH", path("meal-entries", entry.ID), map[string]any{"quantity": 400}, 200)
	if got := keyedEntry(t, f, "lunch-1", body, 201); !bytes.Equal(got, first) {
		t.Fatal("editing food or entry changed the saved response")
	}
	f.request(t, "DELETE", path("meal-entries", entry.ID), nil, 204)
	if got := keyedEntry(t, f, "lunch-1", body, 201); !bytes.Equal(got, first) {
		t.Fatal("deleting entry changed the saved response")
	}
	entryRequestCounts(t, f, 3, 2)
	// Persisted results work through a newly constructed Service/Store/Router.
	repository := store.New(f.db)
	f.router = New(service.New(repository), repository.Ping, openapi.Document)
	if got := keyedEntry(t, f, "lunch-1", body, 201); !bytes.Equal(got, first) {
		t.Fatal("restarting the handler lost the saved response")
	}
}

func TestIntegrationRecipeEntryIdempotency(t *testing.T) {
	f := setup(t)
	food := f.food(t, "牛奶", "ml", nutrition.Values{"energy_kcal": nutrition.Number(60)})
	recipe := response[model.Recipe](t, f, "POST", "/api/v1/recipes", map[string]any{"name": "牛奶菜品", "ingredients": []any{map[string]any{"food_id": food.ID, "quantity": 250, "unit": "ml"}}}, 201)
	body := map[string]any{"meal_type": "breakfast", "item_type": "recipe", "recipe_id": recipe.ID}
	first := keyedEntry(t, f, "recipe-1", body, 201)
	// A replay must not re-read the current recipe or recalculate its nutrition.
	response[model.Recipe](t, f, "PUT", path("recipes", recipe.ID), map[string]any{"name": "新配方", "ingredients": []any{map[string]any{"food_id": food.ID, "quantity": 500, "unit": "ml"}}}, 200)
	if got := keyedEntry(t, f, "recipe-1", body, 201); !bytes.Equal(got, first) {
		t.Fatal("replay refreshed the recipe snapshot")
	}
	entryRequestCounts(t, f, 1, 1)
	// Supplying an explicit date differs from originally omitting it.
	var entry model.MealEntry
	if err := json.Unmarshal(first, &entry); err != nil {
		t.Fatal(err)
	}
	body["date"] = string(entry.Date)
	keyedEntry(t, f, "recipe-1", body, 409)
}

func TestIntegrationEntryIdempotencyRollback(t *testing.T) {
	f := setup(t)
	food := f.food(t, "米饭", "g", nutrition.Values{"energy_kcal": nutrition.Number(100)})
	body := map[string]any{"meal_type": "lunch", "item_type": "food", "food_id": food.ID, "quantity": 0, "unit": "g"}
	keyedEntry(t, f, "retry-1", body, 422)
	entryRequestCounts(t, f, 0, 0)
	body["quantity"] = 200
	// Force failure after the entry INSERT to verify both writes are atomic.
	if err := f.db.Exec(`CREATE FUNCTION reject_entry_request() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'controlled failure'; END $$`).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Exec(`CREATE TRIGGER reject_entry_request BEFORE INSERT ON meal_entry_requests FOR EACH ROW EXECUTE FUNCTION reject_entry_request()`).Error; err != nil {
		t.Fatal(err)
	}
	keyedEntry(t, f, "retry-1", body, 500)
	entryRequestCounts(t, f, 0, 0)
	if err := f.db.Exec(`DROP TRIGGER reject_entry_request ON meal_entry_requests`).Error; err != nil {
		t.Fatal(err)
	}
	keyedEntry(t, f, "retry-1", body, 201)
	entryRequestCounts(t, f, 1, 1)
}

func TestIntegrationEntryIdempotencyConcurrency(t *testing.T) {
	for _, conflict := range []bool{false, true} {
		t.Run(fmt.Sprintf("different_inputs=%v", conflict), func(t *testing.T) {
			f := setup(t)
			food := f.food(t, "米饭", "g", nutrition.Values{"energy_kcal": nutrition.Number(100)})
			// Hold the first INSERT briefly so other requests also observe a missing key.
			if err := f.db.Exec(`CREATE FUNCTION delay_entry_request() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM pg_sleep(0.1); RETURN NEW; END $$`).Error; err != nil {
				t.Fatal(err)
			}
			if err := f.db.Exec(`CREATE TRIGGER delay_entry_request BEFORE INSERT ON meal_entry_requests FOR EACH ROW EXECUTE FUNCTION delay_entry_request()`).Error; err != nil {
				t.Fatal(err)
			}
			type result struct {
				quantity int
				response *httptest.ResponseRecorder
			}
			const n = 12
			start := make(chan struct{})
			results := make(chan result, n)
			for i := range n {
				quantity := 200
				if conflict && i%2 == 1 {
					quantity = 300
				}
				go func() {
					<-start
					body := fmt.Sprintf(`{"meal_type":"lunch","item_type":"food","food_id":%d,"quantity":%d,"unit":"g"}`, food.ID, quantity)
					r := httptest.NewRequest("POST", "/api/v1/meal-entries", strings.NewReader(body))
					r.Header.Set("Content-Type", "application/json")
					r.Header.Set("Idempotency-Key", "concurrent-1")
					w := httptest.NewRecorder()
					f.router.ServeHTTP(w, r)
					results <- result{quantity, w}
				}()
			}
			close(start)
			var responses []result
			var first []byte
			var winner model.MealEntry
			for range n {
				got := <-results
				responses = append(responses, got)
				if got.response.Code == 201 && first == nil {
					first = got.response.Body.Bytes()
					if err := json.Unmarshal(first, &winner); err != nil {
						t.Fatal(err)
					}
				}
			}
			if first == nil || winner.Quantity == nil {
				t.Fatal("no request succeeded")
			}
			for _, got := range responses {
				want := 201
				if float64(got.quantity) != *winner.Quantity {
					want = 409
				}
				if got.response.Code != want || (want == 201 && !bytes.Equal(got.response.Body.Bytes(), first)) {
					t.Fatalf("concurrent response=%d %s want=%d %s", got.response.Code, got.response.Body, want, first)
				}
			}
			entryRequestCounts(t, f, 1, 1)
		})
	}
}

func TestIntegrationEntryIdempotencyMigration(t *testing.T) {
	f := setup(t)
	food := f.food(t, "米饭", "g", nutrition.Values{"energy_kcal": nutrition.Number(100)})
	body := map[string]any{"meal_type": "lunch", "item_type": "food", "food_id": food.ID, "quantity": 200, "unit": "g"}
	keyedEntry(t, f, "migration-1", body, 201)
	db, err := f.db.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := goose.DownTo(db, ".", 2); err != nil {
		t.Fatal(err)
	}
	if f.db.Migrator().HasTable("meal_entry_requests") {
		t.Fatal("down did not remove the idempotency table")
	}
	var count int64
	if err := f.db.Model(&model.MealEntry{}).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("down changed meal entries: count=%d err=%v", count, err)
	}
	if err := migrations.Run(db, "up"); err != nil {
		t.Fatal(err)
	}
	entryRequestCounts(t, f, 1, 0)
	keyedEntry(t, f, "migration-2", body, 201)
	entryRequestCounts(t, f, 2, 1)
}
