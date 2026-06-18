-- Restore material management under parts
CREATE TABLE IF NOT EXISTS formula_ingredient_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    part_id UUID NOT NULL REFERENCES formula_parts(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS formula_ingredients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES formula_ingredient_categories(id) ON DELETE CASCADE,
    material VARCHAR(200) NOT NULL,
    percentage NUMERIC(7,4) NOT NULL DEFAULT 0,
    weight NUMERIC(10,4) NOT NULL DEFAULT 0,
    batch_no VARCHAR(100) NOT NULL DEFAULT '',
    unit VARCHAR(50) NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0
);

-- Keep formula_step_material_categories and formula_step_materials for now
