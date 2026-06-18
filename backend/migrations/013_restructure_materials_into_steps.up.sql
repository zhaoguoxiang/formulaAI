-- 1. Drop dosing_actions (no longer needed - materials are in steps)
DROP TABLE IF EXISTS formula_dosing_actions;

-- 2. Drop old ingredient/category tables (materials move to steps)
DROP TABLE IF EXISTS formula_ingredients;
DROP TABLE IF EXISTS formula_ingredient_categories;

-- 3. Remove mix_ratio from parts (becomes a step parameter)
ALTER TABLE formula_parts DROP COLUMN IF EXISTS mix_ratio;

-- 4. Create step material categories (replaces formula_ingredient_categories)
CREATE TABLE formula_step_material_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    step_id UUID NOT NULL REFERENCES formula_steps(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

-- 5. Create step materials (replaces formula_ingredients)
CREATE TABLE formula_step_materials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES formula_step_material_categories(id) ON DELETE CASCADE,
    material VARCHAR(200) NOT NULL,
    percentage NUMERIC(7,4) NOT NULL DEFAULT 0,
    weight NUMERIC(10,4) NOT NULL DEFAULT 0,
    batch_no VARCHAR(100) NOT NULL DEFAULT '',
    unit VARCHAR(50) NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0
);
