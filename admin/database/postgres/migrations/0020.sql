ALTER TABLE projects DROP COLUMN tags;
ALTER TABLE projects ADD COLUMN annotations JSONB DEFAULT '{}'::JSONB NOT NULL;
