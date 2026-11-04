package service

import (
	"strings"

	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/nutrition"
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

// recipeView 用当前食材数据计算整道菜的营养，也供饮食记录生成快照。
func (s *Service) recipeView(recipe *model.Recipe) (*RecipeView, error) {
	rows, err := s.store.Ingredients(recipe.ID)
	if err != nil {
		return nil, err
	}
	result := &RecipeView{Recipe: *recipe, Ingredients: []model.IngredientSnapshot{}}
	parts := make([]nutrition.Totals, 0, len(rows))
	for _, row := range rows {
		result.Ingredients = append(result.Ingredients, model.IngredientSnapshot{
			FoodID: row.FoodID, Name: row.Food.Name, BasisUnit: row.Food.BasisUnit,
			Quantity: row.Quantity, NutritionPer100: row.Food.NutritionPer100,
		})
		parts = append(parts, nutrition.Scale(row.Food.NutritionPer100, row.Quantity))
	}
	result.NutritionTotals = nutrition.Sum(parts...)
	return result, nil
}

func (s *Service) Recipe(id int64) (*RecipeView, error) {
	recipe, err := s.store.Recipe(id)
	if err != nil {
		return nil, err
	}
	return s.recipeView(recipe)
}

func (s *Service) Recipes(page Page) (List[RecipeView], error) {
	result := List[RecipeView]{Items: []RecipeView{}, Page: page.Number, PageSize: page.Size}
	rows, total, err := s.store.Recipes(page.Query, page.Number, page.Size)
	if err != nil {
		return result, err
	}
	result.Total = total
	for _, row := range rows {
		view, err := s.recipeView(&row)
		if err != nil {
			return result, err
		}
		result.Items = append(result.Items, *view)
	}
	return result, nil
}

// PutRecipe 的 id=0 表示新建，其他值表示替换指定菜品的配方。
func (s *Service) PutRecipe(id int64, in RecipeInput) (*RecipeView, error) {
	// TODO：暂未校验名称、食材数量及用量，调用方提交完整配方。
	recipe := &model.Recipe{}
	if id != 0 {
		var err error
		recipe, err = s.store.Recipe(id)
		if err != nil {
			return nil, err
		}
	}
	recipe.Name = strings.TrimSpace(in.Name)
	rows := make([]model.RecipeIngredient, 0, len(in.Ingredients))
	for _, item := range in.Ingredients {
		rows = append(rows, model.RecipeIngredient{
			FoodID: item.FoodID, Quantity: nutrition.Quantity(item.Quantity, item.Unit),
		})
	}

	// TODO：以下写入暂未使用事务，也未处理并发替换；中途失败可能留下部分配方。
	var err error
	if id == 0 {
		err = s.store.CreateRecipe(recipe)
	} else {
		err = s.store.SaveRecipe(recipe)
	}
	if err != nil {
		return nil, err
	}
	if err := s.store.ReplaceIngredients(recipe.ID, rows); err != nil {
		return nil, err
	}
	return s.recipeView(recipe)
}
