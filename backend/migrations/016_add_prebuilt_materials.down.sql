ALTER TABLE formula_parts DROP COLUMN IF EXISTS material_id;
ALTER TABLE formula_parts DROP COLUMN IF EXISTS batch_no;

ALTER TABLE formulas DROP COLUMN IF EXISTS labels;
ALTER TABLE formulas DROP COLUMN IF EXISTS formula_type;
