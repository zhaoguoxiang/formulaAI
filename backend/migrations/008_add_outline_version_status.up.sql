ALTER TABLE test_outlines ADD COLUMN version INT NOT NULL DEFAULT 1;
ALTER TABLE test_outlines ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'draft';
ALTER TABLE test_outlines ADD CHECK (status IN ('draft', 'active', 'archived'));
