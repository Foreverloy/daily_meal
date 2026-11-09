package service

import (
	"strings"

	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/nutrition"
)

// IngredientInput 是配方中的一项食材输入，进入业务层后用量统一换算成 g/ml。
type IngredientInput struct {
	// FoodID 是食物库中已有食物的 ID。
	FoodID int64 `json:"food_id"`
	// Quantity 是与 Unit 搭配的用量，如 Quantity=0.2、Unit="kg" 表示 200g。
	Quantity float64 `json:"quantity"`
	// Unit 是输入单位，支持 g/kg 或 ml/L，按约定须与食物的重量/体积基准同类。
	Unit string `json:"unit"`
}

// RecipeInput 是新建或整体替换菜品配方的参数，更新时提交完整食材列表。
type RecipeInput struct {
	// Name 是菜品名称，如“鸡胸肉沙拉”。
	Name string `json:"name"`
	// Ingredients 是整道菜的食材及用量，不按份数或成品重量拆分。
	Ingredients []IngredientInput `json:"ingredients"`
}

// RecipeView 是菜品详情响应，使用当前食物库数据计算，不是历史饮食记录。
type RecipeView struct {
	// Recipe 嵌入菜品的 ID、名称和时间字段，JSON 中这些字段直接位于外层。
	model.Recipe
	// Ingredients 是本次查询得到的食材详情，记餐时可复制到记录快照中。
	Ingredients []model.IngredientSnapshot `json:"ingredients"`
	// NutritionTotals 是整道菜各项食材的营养合计，保留缺失指标的不完整标记。
	NutritionTotals nutrition.Totals `json:"nutrition_totals"`
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
	catalog, err := s.store.Nutrients()
	if err != nil {
		return nil, err
	}
	result.NutritionTotals = nutrition.Sum(catalog.Keys(), parts...)
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
