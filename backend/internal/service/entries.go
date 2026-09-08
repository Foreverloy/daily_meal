package service

import (
	"context"
	"errors"

	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/nutrition"
	"gorm.io/gorm"
)

type EntryInput struct {
	Date     Field[string]  `json:"date"`
	MealType Field[string]  `json:"meal_type"`
	ItemType Field[string]  `json:"item_type"`
	FoodID   Field[int64]   `json:"food_id"`
	RecipeID Field[int64]   `json:"recipe_id"`
	Quantity Field[float64] `json:"quantity"`
	Unit     Field[string]  `json:"unit"`
}

func validMeal(meal string) bool { return meal == "breakfast" || meal == "lunch" || meal == "dinner" }

func (s *Service) Entry(ctx context.Context, id int64) (*model.MealEntry, error) {
	entry, err := s.store.Entry(ctx, id, false)
	return entry, resourceError(err)
}

func (s *Service) Entries(ctx context.Context, rawDate, meal string) ([]model.MealEntry, error) {
	date, err := ParseDate(rawDate)
	if err != nil {
		return nil, err
	}
	if meal != "" && !validMeal(meal) {
		return nil, Invalid("meal_type", "餐次须为 breakfast、lunch 或 dinner")
	}
	return s.store.Entries(ctx, date, meal)
}

func (s *Service) DeleteEntry(ctx context.Context, id int64) error {
	return resourceError(s.store.DeleteEntry(ctx, id))
}

// SaveEntry uses id=0 for creation. Metadata/quantity edits reuse the original snapshot;
// only selecting a different item reads the current food or recipe.
func (s *Service) SaveEntry(ctx context.Context, id int64, in EntryInput) (*model.MealEntry, error) {
	for key, null := range map[string]bool{"date": in.Date.Null, "meal_type": in.MealType.Null, "item_type": in.ItemType.Null, "food_id": in.FoodID.Null, "recipe_id": in.RecipeID.Null, "quantity": in.Quantity.Null, "unit": in.Unit.Null} {
		if null {
			return nil, Invalid(key, "不能为 null；未修改的字段请省略")
		}
	}
	var result *model.MealEntry
	err := s.transaction(ctx, func(tx *Service) error {
		entry := &model.MealEntry{Date: tx.Today()}
		if id != 0 {
			var err error
			entry, err = tx.store.Entry(ctx, id, true)
			if err != nil {
				return resourceError(err)
			}
		}
		oldType := entry.ItemType
		if in.Date.Set {
			date, err := ParseDate(in.Date.Value)
			if err != nil {
				return err
			}
			entry.Date = date
		}
		if in.MealType.Set {
			entry.MealType = in.MealType.Value
		}
		if !validMeal(entry.MealType) {
			return Invalid("meal_type", "餐次须为 breakfast、lunch 或 dinner")
		}
		if in.ItemType.Set {
			entry.ItemType = in.ItemType.Value
		}
		newItem := id == 0 || oldType != entry.ItemType
		var err error
		switch entry.ItemType {
		case "food":
			err = tx.applyFoodEntry(ctx, entry, in, newItem)
		case "recipe":
			err = tx.applyRecipeEntry(ctx, entry, in, newItem)
		default:
			return Invalid("item_type", "类型须为 food 或 recipe")
		}
		if err != nil {
			return err
		}
		result = entry
		if id == 0 {
			return tx.store.CreateEntry(ctx, entry)
		}
		return tx.store.SaveEntry(ctx, entry)
	})
	return result, err
}

func (s *Service) applyFoodEntry(ctx context.Context, entry *model.MealEntry, in EntryInput, newItem bool) error {
	oldFoodID := entry.FoodID
	if in.RecipeID.Set {
		return Invalid("recipe_id", "食物记录不能填写 recipe_id")
	}
	if in.FoodID.Set {
		entry.FoodID = &in.FoodID.Value
	}
	if entry.FoodID == nil || *entry.FoodID <= 0 {
		return Invalid("food_id", "食物 ID 必填且须大于 0")
	}
	changed := newItem || oldFoodID == nil || *oldFoodID != *entry.FoodID
	if changed {
		food, err := s.store.Food(ctx, *entry.FoodID, false)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Invalid("food_id", "引用的食物不存在")
		}
		if err != nil {
			return err
		}
		entry.Snapshot = model.Snapshot{Name: food.Name, BasisUnit: food.BasisUnit, NutritionPer100: food.NutritionPer100}
		if !in.Quantity.Set || !in.Unit.Set {
			return Invalid("quantity", "选择新食物时须同时填写 quantity 和 unit")
		}
	}
	if in.Unit.Set && !in.Quantity.Set {
		return Invalid("quantity", "修改单位时须同时填写 quantity")
	}
	if in.Quantity.Set {
		unit := entry.Snapshot.BasisUnit
		if in.Unit.Set {
			unit = in.Unit.Value
		}
		quantity, err := nutrition.Quantity(in.Quantity.Value, unit, entry.Snapshot.BasisUnit)
		if err != nil {
			return Invalid("quantity", err.Error())
		}
		entry.Quantity = &quantity
	}
	if changed || in.Quantity.Set {
		totals, err := nutrition.Scale(entry.Snapshot.NutritionPer100, *entry.Quantity)
		if err != nil {
			return Invalid("quantity", err.Error())
		}
		entry.NutritionTotals = totals
	}
	entry.RecipeID = nil
	return nil
}

func (s *Service) applyRecipeEntry(ctx context.Context, entry *model.MealEntry, in EntryInput, newItem bool) error {
	oldRecipeID := entry.RecipeID
	if in.FoodID.Set || in.Quantity.Set || in.Unit.Set {
		return Invalid("item_type", "菜品整道计入，不能填写 food_id、quantity 或 unit")
	}
	if in.RecipeID.Set {
		entry.RecipeID = &in.RecipeID.Value
	}
	if entry.RecipeID == nil || *entry.RecipeID <= 0 {
		return Invalid("recipe_id", "菜品 ID 必填且须大于 0")
	}
	changed := newItem || oldRecipeID == nil || *oldRecipeID != *entry.RecipeID
	if changed {
		recipe, err := s.store.Recipe(ctx, *entry.RecipeID, false)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Invalid("recipe_id", "引用的菜品不存在")
		}
		if err != nil {
			return err
		}
		view, err := s.recipeView(ctx, recipe)
		if err != nil {
			return err
		}
		entry.Snapshot = model.Snapshot{Name: recipe.Name, Ingredients: view.Ingredients}
		entry.NutritionTotals = view.NutritionTotals
	}
	entry.FoodID, entry.Quantity = nil, nil
	return nil
}
