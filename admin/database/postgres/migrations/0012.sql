ALTER TABLE user_auth_tokens ADD COLUMN used_on TIMESTAMPTZ DEFAULT now() NOT NULL;
ALTER TABLE service_auth_tokens ADD COLUMN used_on TIMESTAMPTZ DEFAULT now() NOT NULL;
