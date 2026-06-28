-- Revert 017: Remove project support

ALTER TABLE test_outlines DROP CONSTRAINT IF EXISTS fk_test_outlines_project;
ALTER TABLE test_outlines DROP CONSTRAINT IF EXISTS test_outlines_project_name_version_unique;
DROP INDEX IF EXISTS idx_test_outlines_project_id;
ALTER TABLE test_outlines DROP COLUMN IF EXISTS project_id;
ALTER TABLE test_outlines ADD CONSTRAINT test_outlines_name_version_key UNIQUE (name, version);

ALTER TABLE formulas DROP CONSTRAINT IF EXISTS fk_formulas_project;
ALTER TABLE formulas DROP CONSTRAINT IF EXISTS formulas_project_code_unique;
DROP INDEX IF EXISTS idx_formulas_project_id;
ALTER TABLE formulas DROP COLUMN IF EXISTS project_id;
ALTER TABLE formulas ADD CONSTRAINT formulas_code_key UNIQUE (code);

DROP TABLE IF EXISTS projects;
