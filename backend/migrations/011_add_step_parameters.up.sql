-- 1. Add instrument_name to formula_steps
ALTER TABLE formula_steps ADD COLUMN IF NOT EXISTS instrument_name VARCHAR(100) NOT NULL DEFAULT '';

-- 2. Create formula_step_parameters table
CREATE TABLE formula_step_parameters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    step_id UUID NOT NULL REFERENCES formula_steps(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    value VARCHAR(200) NOT NULL DEFAULT '',
    unit VARCHAR(50) NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0
);

-- 3. Migrate existing temperature data into parameters
INSERT INTO formula_step_parameters (step_id, name, value, unit, sort_order)
SELECT id, '温度', temperature, '°C', 0 FROM formula_steps WHERE temperature IS NOT NULL AND temperature != '';

-- 4. Migrate existing duration data into parameters
INSERT INTO formula_step_parameters (step_id, name, value, unit, sort_order)
SELECT id, '时长', duration, 'min', 1 FROM formula_steps WHERE duration IS NOT NULL AND duration != '';
