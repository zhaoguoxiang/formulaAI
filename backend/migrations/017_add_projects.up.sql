-- Migration 017: Add multi-project/workspace support
-- Creates projects table and adds project_id to formulas and test_outlines.
-- All existing data will need a default project before this migration runs.

-- 1. Create projects table
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 2. Add project_id to formulas
ALTER TABLE formulas ADD COLUMN IF NOT EXISTS project_id UUID;

-- Drop old unique constraint on code, add new project-scoped one
ALTER TABLE formulas DROP CONSTRAINT IF EXISTS formulas_code_key;
ALTER TABLE formulas ADD CONSTRAINT formulas_project_code_unique UNIQUE (project_id, code);

-- Make project_id NOT NULL after ensuring data exists
-- (callers should create a default project and assign it before running this)
ALTER TABLE formulas ALTER COLUMN project_id SET NOT NULL;

-- Add FK and index
ALTER TABLE formulas ADD CONSTRAINT fk_formulas_project
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_formulas_project_id ON formulas(project_id);

-- 3. Add project_id to test_outlines
ALTER TABLE test_outlines ADD COLUMN IF NOT EXISTS project_id UUID;

-- Drop old unique constraint on (name, version), add new project-scoped one
ALTER TABLE test_outlines DROP CONSTRAINT IF EXISTS test_outlines_name_version_key;
ALTER TABLE test_outlines ADD CONSTRAINT test_outlines_project_name_version_unique
    UNIQUE (project_id, name, version);

ALTER TABLE test_outlines ALTER COLUMN project_id SET NOT NULL;

ALTER TABLE test_outlines ADD CONSTRAINT fk_test_outlines_project
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_test_outlines_project_id ON test_outlines(project_id);

-- Rebuild the name index (was dropped with the constraint)
CREATE INDEX IF NOT EXISTS idx_test_outlines_name ON test_outlines(name);
