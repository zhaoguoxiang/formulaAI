CREATE TABLE formula_dosing_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    step_id UUID NOT NULL REFERENCES formula_steps(id) ON DELETE CASCADE,
    ingredient_id UUID NOT NULL REFERENCES formula_ingredients(id) ON DELETE CASCADE,
    dosing_order INT NOT NULL DEFAULT 0,
    use_ratio DECIMAL(7,4) NOT NULL CHECK (use_ratio > 0 AND use_ratio <= 100),
    dosing_method VARCHAR(100)
);
