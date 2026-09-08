-- +goose Up
CREATE TABLE goal_settings (
    id BIGSERIAL PRIMARY KEY,
    effective_on DATE NOT NULL UNIQUE,
    energy_kcal NUMERIC NOT NULL CHECK (energy_kcal > 0 AND energy_kcal < 'Infinity'::numeric),
    protein_percent NUMERIC NOT NULL CHECK (protein_percent BETWEEN 0 AND 100),
    carbohydrate_percent NUMERIC NOT NULL CHECK (carbohydrate_percent BETWEEN 0 AND 100),
    fat_percent NUMERIC NOT NULL CHECK (fat_percent BETWEEN 0 AND 100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (abs(protein_percent + carbohydrate_percent + fat_percent - 100) <= 0.00000001)
);

CREATE TABLE daily_goals (
    id BIGSERIAL PRIMARY KEY,
    date DATE NOT NULL UNIQUE,
    energy_kcal NUMERIC NOT NULL CHECK (energy_kcal > 0 AND energy_kcal < 'Infinity'::numeric),
    protein_percent NUMERIC NOT NULL CHECK (protein_percent BETWEEN 0 AND 100),
    carbohydrate_percent NUMERIC NOT NULL CHECK (carbohydrate_percent BETWEEN 0 AND 100),
    fat_percent NUMERIC NOT NULL CHECK (fat_percent BETWEEN 0 AND 100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (abs(protein_percent + carbohydrate_percent + fat_percent - 100) <= 0.00000001)
);

CREATE TABLE foods (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL CHECK (char_length(btrim(name)) BETWEEN 1 AND 200),
    basis_unit TEXT NOT NULL CHECK (basis_unit IN ('g', 'ml')),
    nutrition_per_100 JSONB NOT NULL CHECK (jsonb_typeof(nutrition_per_100) = 'object' AND
        nutrition_per_100 ? 'energy_kcal' AND jsonb_typeof(nutrition_per_100->'energy_kcal') = 'number'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE recipes (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL CHECK (char_length(btrim(name)) BETWEEN 1 AND 200),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE recipe_ingredients (
    id BIGSERIAL PRIMARY KEY,
    recipe_id BIGINT NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    food_id BIGINT NOT NULL REFERENCES foods(id),
    quantity NUMERIC NOT NULL CHECK (quantity > 0 AND quantity < 'Infinity'::numeric),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX recipe_ingredients_recipe_id_idx ON recipe_ingredients(recipe_id);
CREATE INDEX recipe_ingredients_food_id_idx ON recipe_ingredients(food_id);

CREATE TABLE meal_entries (
    id BIGSERIAL PRIMARY KEY,
    date DATE NOT NULL,
    meal_type TEXT NOT NULL CHECK (meal_type IN ('breakfast', 'lunch', 'dinner')),
    item_type TEXT NOT NULL CHECK (item_type IN ('food', 'recipe')),
    food_id BIGINT REFERENCES foods(id),
    recipe_id BIGINT REFERENCES recipes(id),
    quantity NUMERIC CHECK (quantity > 0 AND quantity < 'Infinity'::numeric),
    snapshot JSONB NOT NULL CHECK (jsonb_typeof(snapshot) = 'object'),
    nutrition_totals JSONB NOT NULL CHECK (jsonb_typeof(nutrition_totals) = 'object'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((item_type = 'food' AND food_id IS NOT NULL AND recipe_id IS NULL AND quantity IS NOT NULL)
        OR (item_type = 'recipe' AND recipe_id IS NOT NULL AND food_id IS NULL AND quantity IS NULL))
);
CREATE INDEX meal_entries_date_meal_type_idx ON meal_entries(date, meal_type);

-- +goose Down
DROP TABLE meal_entries;
DROP TABLE recipe_ingredients;
DROP TABLE recipes;
DROP TABLE foods;
DROP TABLE daily_goals;
DROP TABLE goal_settings;
