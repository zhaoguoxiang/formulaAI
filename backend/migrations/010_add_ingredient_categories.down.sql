ALTER TABLE formula_ingredients DROP CONSTRAINT IF EXISTS fk_ingredient_category;
ALTER TABLE formula_ingredients DROP COLUMN IF EXISTS category_id;
DROP TABLE IF EXISTS formula_ingredient_categories;
