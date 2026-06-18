ALTER TABLE formulas ADD COLUMN IF NOT EXISTS formula_type VARCHAR(20) NOT NULL DEFAULT 'formula';
ALTER TABLE formulas ADD CONSTRAINT chk_formula_type CHECK (formula_type IN ('formula', 'material'));
ALTER TABLE formulas ADD COLUMN IF NOT EXISTS labels TEXT[] NOT NULL DEFAULT '{}';

ALTER TABLE formula_parts ADD COLUMN IF NOT EXISTS batch_no VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE formula_parts ADD COLUMN IF NOT EXISTS material_id UUID REFERENCES formulas(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_formulas_type ON formulas(formula_type);
CREATE INDEX IF NOT EXISTS idx_formula_parts_material_id ON formula_parts(material_id);
