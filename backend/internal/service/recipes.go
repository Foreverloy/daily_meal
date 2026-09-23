package service

import (
	"errors"
	"fmt"

	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/nutrition"
	"gorm.io/gorm"
)

type IngredientInput struct {
	FoodID   int64   `json:"food_id"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

type RecipeInput struct {
	Name        string            `json:"name"`
	Ingredients []IngredientInput `json:"ingredients"`
}

type RecipeView struct {
	model.Recipe
	Ingredients     []model.IngredientSnapshot `json:"ingredients"`
	NutritionTotals nutrition.Totals           `json:"nutrition_totals"`
}

func (s *Service) recipeView(recipe *model.Recipe) (*RecipeView, error) {
	rows, err := s.store.Ingredients(recipe.ID)
	if err != nil {
		return nil, err
	}
	result := &RecipeView{Recipe: *recipe, Ingredients: []model.IngredientSnapshot{}}
	parts := make([]nutrition.Totals, 0, len(rows))
	for _, row := range rows {
		result.Ingredients = append(result.Ingredients, model.IngredientSnapshot{FoodID: row.FoodID, Name: row.Food.Name, BasisUnit: row.Food.BasisUnit, Quantity: row.Quantity, NutritionPer100: row.Food.NutritionPer100})
		totals, err := nutrition.Scale(row.Food.NutritionPer100, row.Quantity)
		if err != nil {
			return nil, Invalid("ingredients", err.Error())
		}
		parts = append(parts, totals)
	}
	result.NutritionTotals, err = nutrition.Sum(parts...)
	if err != nil {
		return nil, Invalid("ingredients", err.Error())
	}
	return result, nil
}

func (s *Service) Recipe(id int64) (*RecipeView, error) {
	var result *RecipeView
	err := s.transaction(func(tx *Service) error {
		recipe, err := tx.store.Recipe(id, false)
		if err != nil {
			return resourceError(err)
		}
		result, err = tx.recipeView(recipe)
		return err
	})
	return result, err
}

func (s *Service) Recipes(page Page) (List[RecipeView], error) {
	result := List[RecipeView]{Items: []RecipeView{}, Page: page.Number, PageSize: page.Size}
	if err := page.Validate(); err != nil {
		return result, err
	}
	err := s.transaction(func(tx *Service) error {
		rows, total, err := tx.store.Recipes(page.Query, page.Number, page.Size)
		if err != nil {
			return err
		}
		result.Total = total
		for _, row := range rows {
			view, err := tx.recipeView(&row)
			if err != nil {
				return err
			}
			result.Items = append(result.Items, *view)
		}
		return nil
	})
	return result, err
}

// PutRecipe uses id=0 for creation; replacement is all-or-nothing.
func (s *Service) PutRecipe(id int64, in RecipeInput) (*RecipeView, error) {
	name, err := cleanName(in.Name)
	if err != nil {
		return nil, err
	}
	if len(in.Ingredients) == 0 {
		return nil, Invalid("ingredients", "至少需要一种食材")
	}
	var result *RecipeView
	err = s.transaction(func(tx *Service) error {
		recipe := &model.Recipe{Name: name}
		if id != 0 {
			var err error
			recipe, err = tx.store.Recipe(id, true)
			if err != nil {
				return resourceError(err)
			}
			recipe.Name = name
		}
		rows := make([]model.RecipeIngredient, 0, len(in.Ingredients))
		for i, item := range in.Ingredients {
			field := fmt.Sprintf("ingredients.%d", i)
			if item.FoodID <= 0 {
				return Invalid(field+".food_id", "食物 ID 必须大于 0")
			}
			food, err := tx.store.Food(item.FoodID, false)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return Invalid(field+".food_id", "引用的食物不存在")
			}
			if err != nil {
				return err
			}
			quantity, err := nutrition.Quantity(item.Quantity, item.Unit, food.BasisUnit)
			if err != nil {
				return Invalid(field+".quantity", err.Error())
			}
			rows = append(rows, model.RecipeIngredient{FoodID: food.ID, Quantity: quantity})
		}
		if id == 0 {
			err = tx.store.CreateRecipe(recipe)
		} else {
			err = tx.store.SaveRecipe(recipe)
		}
		if err != nil {
			return err
		}
		if err := tx.store.ReplaceIngredients(recipe.ID, rows); err != nil {
			return err
		}
		result, err = tx.recipeView(recipe)
		return err
	})
	return result, err
}
