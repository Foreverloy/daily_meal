package store

import "dailymeal/backend/internal/model"

func (s *Store) Recipe(id int64, forUpdate bool) (*model.Recipe, error) {
	var recipe model.Recipe
	err := lock(s.db, forUpdate).First(&recipe, id).Error
	return &recipe, err
}

func (s *Store) Recipes(q string, page, size int) ([]model.Recipe, int64, error) {
	items := []model.Recipe{}
	var total int64
	query := named(s.db.Model(&model.Recipe{}), q)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&items).Error
	return items, total, err
}

func (s *Store) Ingredients(recipeID int64) ([]model.RecipeIngredient, error) {
	items := []model.RecipeIngredient{}
	err := s.db.Preload("Food").Where("recipe_id = ?", recipeID).Order("id ASC").Find(&items).Error
	return items, err
}

func (s *Store) CreateRecipe(recipe *model.Recipe) error {
	return s.db.Create(recipe).Error
}
func (s *Store) SaveRecipe(recipe *model.Recipe) error {
	return s.db.Save(recipe).Error
}

func (s *Store) ReplaceIngredients(recipeID int64, ingredients []model.RecipeIngredient) error {
	if err := s.db.Where("recipe_id = ?", recipeID).Delete(&model.RecipeIngredient{}).Error; err != nil {
		return err
	}
	for i := range ingredients {
		ingredients[i].RecipeID = recipeID
	}
	return s.db.Omit("Food").Create(&ingredients).Error
}
