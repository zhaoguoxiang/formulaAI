DROP TABLE IF EXISTS formula_step_materials;
DROP TABLE IF EXISTS formula_step_material_categories;
ALTER TABLE formula_parts ADD COLUMN IF NOT EXISTS mix_ratio NUMERIC(5,2) NOT NULL DEFAULT 100;
-- Note: formula_ingredients, formula_ingredient_categories, formula_dosing_actions cannot be restored from this down migration
