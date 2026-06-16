CREATE TABLE formula_ingredients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    part_id UUID NOT NULL REFERENCES formula_parts(id) ON DELETE CASCADE,
    sort_order INT NOT NULL DEFAULT 0,
    material VARCHAR(200) NOT NULL,
    percentage DECIMAL(7,4) NOT NULL CHECK (percentage > 0 AND percentage <= 100),
    weight DECIMAL(10,4)
);
