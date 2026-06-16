CREATE TABLE formula_parts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    formula_id UUID NOT NULL REFERENCES formulas(id) ON DELETE CASCADE,
    name VARCHAR(20) NOT NULL CHECK (name IN ('A', 'B', 'MAIN')),
    mix_ratio DECIMAL(5,2) NOT NULL DEFAULT 100.00,
    sort_order INT NOT NULL DEFAULT 0,
    UNIQUE(formula_id, name)
);
