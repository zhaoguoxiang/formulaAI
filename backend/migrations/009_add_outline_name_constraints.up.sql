ALTER TABLE test_outlines ADD UNIQUE(name, version);
CREATE INDEX idx_test_outlines_name ON test_outlines(name);
