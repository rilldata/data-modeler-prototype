ALTER TABLE orgs ADD COLUMN favicon_asset_id UUID REFERENCES assets(id) ON DELETE SET NULL;
