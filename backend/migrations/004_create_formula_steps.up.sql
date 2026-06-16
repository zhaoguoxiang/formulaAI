CREATE TABLE formula_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    formula_id UUID NOT NULL REFERENCES formulas(id) ON DELETE CASCADE,
    part_id UUID REFERENCES formula_parts(id) ON DELETE SET NULL,
    step_no INT NOT NULL,
    name VARCHAR(200) NOT NULL,
    temperature VARCHAR(50),
    duration VARCHAR(50)
);
