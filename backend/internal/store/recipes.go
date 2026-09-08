package store

import (
	"context"
	"dailymeal/backend/internal/model"
)

func (s *Store) Recipe(ctx context.Context, id int64, forUpdate bool) (*model.Recipe, error) {
	var recipe model.Recipe
	err := lock(s.db.WithContext(ctx), forUpdate).First(&recipe, id).Error
	return &recipe, err
}

func (s *Store) Recipes(ctx context.Context, q string, page, size int) ([]model.Recipe, int64, error) {
	items := []model.Recipe{}
	var total int64
	query := named(s.db.WithContext(ctx).Model(&model.Recipe{}), q)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&items).Error
	return items, total, err
}

func (s *Store) Ingredients(ctx context.Context, recipeID int64) ([]model.RecipeIngredient, error) {
	items := []model.RecipeIngredient{}
	err := s.db.WithContext(ctx).Preload("Food").Where("recipe_id = ?", recipeID).Order("id ASC").Find(&items).Error
	return items, err
}

func (s *Store) CreateRecipe(ctx context.Context, recipe *model.Recipe) error {
	return s.db.WithContext(ctx).Create(recipe).Error
}
func (s *Store) SaveRecipe(ctx context.Context, recipe *model.Recipe) error {
	return s.db.WithContext(ctx).Save(recipe).Error
}

func (s *Store) ReplaceIngredients(ctx context.Context, recipeID int64, ingredients []model.RecipeIngredient) error {
	if err := s.db.WithContext(ctx).Where("recipe_id = ?", recipeID).Delete(&model.RecipeIngredient{}).Error; err != nil {
		return err
	}
	for i := range ingredients {
		ingredients[i].RecipeID = recipeID
	}
	return s.db.WithContext(ctx).Omit("Food").Create(&ingredients).Error
}
