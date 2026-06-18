-- 1. Create formula_ingredient_categories table
CREATE TABLE formula_ingredient_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    part_id UUID NOT NULL REFERENCES formula_parts(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

-- 2. Add category_id column to formula_ingredients (nullable first for migration)
ALTER TABLE formula_ingredients ADD COLUMN IF NOT EXISTS category_id UUID;

-- 3. Create default categories for existing parts that have ingredients
INSERT INTO formula_ingredient_categories (part_id, name, sort_order)
SELECT id, '默认分类', 0 FROM formula_parts
WHERE EXISTS (SELECT 1 FROM formula_ingredients WHERE part_id = formula_parts.id);

-- 4. Assign existing ingredients to their part's default category
UPDATE formula_ingredients fi
SET category_id = (
    SELECT c.id FROM formula_ingredient_categories c WHERE c.part_id = fi.part_id
);

-- 5. Make category_id NOT NULL
ALTER TABLE formula_ingredients ALTER COLUMN category_id SET NOT NULL;

-- 6. Add foreign key constraint
ALTER TABLE formula_ingredients ADD CONSTRAINT fk_ingredient_category
    FOREIGN KEY (category_id) REFERENCES formula_ingredient_categories(id);
