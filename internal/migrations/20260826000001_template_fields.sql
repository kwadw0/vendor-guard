-- +goose Up
-- +goose StatementBegin

-- Add template_id to form_fields (nullable — set for template fields, null for form fields)
ALTER TABLE form_fields ADD COLUMN template_id UUID REFERENCES form_templates(id) ON DELETE CASCADE;
CREATE INDEX idx_form_fields_template_id ON form_fields(template_id);

-- Drop the old per-form unique constraint and replace with one that handles both cases
ALTER TABLE form_fields DROP CONSTRAINT IF EXISTS uq_form_fields_form_id_key;
ALTER TABLE form_fields ADD CONSTRAINT uq_form_fields_form_id_key UNIQUE (form_id, key);
ALTER TABLE form_fields ADD CONSTRAINT uq_form_fields_template_id_key UNIQUE (template_id, key);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE form_fields DROP CONSTRAINT IF EXISTS uq_form_fields_template_id_key;
ALTER TABLE form_fields DROP CONSTRAINT IF EXISTS uq_form_fields_form_id_key;
ALTER TABLE form_fields ADD CONSTRAINT uq_form_fields_form_id_key UNIQUE (form_id, key);
DROP INDEX IF EXISTS idx_form_fields_template_id;
ALTER TABLE form_fields DROP COLUMN IF EXISTS template_id;
-- +goose StatementEnd
