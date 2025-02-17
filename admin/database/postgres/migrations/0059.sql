ALTER TABLE magic_auth_tokens ADD COLUMN resources JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE magic_auth_tokens
SET resources = jsonb_build_array(
        jsonb_build_object(
                'type', resource_type,
                'name', resource_name
        )
                );

ALTER TABLE magic_auth_tokens DROP COLUMN resource_name, DROP COLUMN resource_type;
