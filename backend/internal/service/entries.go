package service

import (
	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/nutrition"
)

// EntryInput 是新建或修改饮食记录的参数；Field 记录字段是否提交，便于部分更新。
// 例如食物用量可提交 {"quantity":0.2,"unit":"kg"}；菜品记录按整道菜计入。
type EntryInput struct {
	// Date 是用餐日期，如 "2026-11-09"；新建时省略则取今天，更新时省略则保留原日期。
	Date Field[string] `json:"date"`
	// MealType 是餐次：breakfast（早）、lunch（中）、dinner（晚）。
	MealType Field[string] `json:"meal_type"`
	// ItemType 是记录类型：food 或 recipe，决定读取哪一种来源对象。
	ItemType Field[string] `json:"item_type"`
	// FoodID 是 food 类型对应的食物 ID。
	FoodID Field[int64] `json:"food_id"`
	// RecipeID 是 recipe 类型对应的菜品 ID。
	RecipeID Field[int64] `json:"recipe_id"`
	// Quantity 是食物用量，与 Unit 配合提交；菜品记录不使用此字段。
	Quantity Field[float64] `json:"quantity"`
	// Unit 是食物输入用量的单位：g/kg 或 ml/L；保存前换算为 g/ml。
	Unit Field[string] `json:"unit"`
}

func (s *Service) Entry(id int64) (*model.MealEntry, error) {
	return s.store.Entry(id)
}

func (s *Service) Entries(rawDate, meal string) ([]model.MealEntry, error) {
	date, err := ParseDate(rawDate)
	if err != nil {
		return nil, err
	}
	return s.store.Entries(date, meal)
}

func (s *Service) DeleteEntry(id int64) error {
	return s.store.DeleteEntry(id)
}

// SaveEntry 的 id=0 表示新建；修改日期、餐次或用量时沿用原快照，换食物或菜品才读取新数据。
func (s *Service) SaveEntry(id int64, in EntryInput) (*model.MealEntry, error) {
	// TODO：暂未校验字段组合与必填值，调用方按 food/recipe 类型提交对应字段。
	// TODO：暂未处理并发修改，也不保证读取菜品食材时来自同一数据库快照。
	entry := &model.MealEntry{Date: s.Today()}
	if id != 0 {
		var err error
		entry, err = s.store.Entry(id)
		if err != nil {
			return nil, err
		}
	}
	oldType := entry.ItemType
	if in.Date.Set {
		date, err := ParseDate(in.Date.Value)
		if err != nil {
			return nil, err
		}
		entry.Date = date
	}
	if in.MealType.Set {
		entry.MealType = in.MealType.Value
	}
	if in.ItemType.Set {
		entry.ItemType = in.ItemType.Value
	}
	newItem := id == 0 || oldType != entry.ItemType
	var err error
	switch entry.ItemType {
	case "food":
		err = s.applyFoodEntry(entry, in, newItem)
	case "recipe":
		err = s.applyRecipeEntry(entry, in, newItem)
	}
	if err != nil {
		return nil, err
	}
	if id == 0 {
		return entry, s.store.CreateEntry(entry)
	}
	return entry, s.store.SaveEntry(entry)
}

func (s *Service) applyFoodEntry(entry *model.MealEntry, in EntryInput, newItem bool) error {
	var foodID int64
	if entry.FoodID != nil {
		foodID = *entry.FoodID
	}
	changed := newItem || (in.FoodID.Set && in.FoodID.Value != foodID)
	if in.FoodID.Set {
		foodID = in.FoodID.Value
	}
	entry.FoodID = &foodID
	if changed {
		food, err := s.store.Food(foodID)
		if err != nil {
			return err
		}
		entry.Snapshot = model.Snapshot{Name: food.Name, BasisUnit: food.BasisUnit, NutritionPer100: food.NutritionPer100}
	}
	if changed || in.Quantity.Set {
		quantity := nutrition.Quantity(in.Quantity.Value, in.Unit.Value)
		entry.Quantity = &quantity
		entry.NutritionTotals = nutrition.Scale(entry.Snapshot.NutritionPer100, quantity)
	}
	entry.RecipeID = nil
	return nil
}

func (s *Service) applyRecipeEntry(entry *model.MealEntry, in EntryInput, newItem bool) error {
	var recipeID int64
	if entry.RecipeID != nil {
		recipeID = *entry.RecipeID
	}
	changed := newItem || (in.RecipeID.Set && in.RecipeID.Value != recipeID)
	if in.RecipeID.Set {
		recipeID = in.RecipeID.Value
	}
	entry.RecipeID = &recipeID
	if changed {
		recipe, err := s.store.Recipe(recipeID)
		if err != nil {
			return err
		}
		view, err := s.recipeView(recipe)
		if err != nil {
			return err
		}
		entry.Snapshot = model.Snapshot{Name: recipe.Name, Ingredients: view.Ingredients}
		entry.NutritionTotals = view.NutritionTotals
	}
	entry.FoodID, entry.Quantity = nil, nil
	return nil
}
