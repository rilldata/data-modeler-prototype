-- keeping the fields nullable as we need to be able to distinguish between plan not cached vs no plan
ALTER TABLE orgs ADD COLUMN billing_plan_name TEXT;
ALTER TABLE orgs ADD COLUMN billing_plan_display_name TEXT;
