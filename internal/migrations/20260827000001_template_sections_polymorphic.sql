-- +goose Up
-- +goose StatementBegin

-- Make form_sections polymorphic: belongs to either a form or a template
ALTER TABLE form_sections ADD COLUMN template_id UUID REFERENCES form_templates(id) ON DELETE CASCADE;
ALTER TABLE form_sections ALTER COLUMN form_id DROP NOT NULL;
ALTER TABLE form_sections ADD CONSTRAINT chk_form_sections_owner CHECK (
  (form_id IS NULL) != (template_id IS NULL)
);
CREATE INDEX idx_form_sections_template_id ON form_sections(template_id);

-- Make form_fields.form_id nullable (was NOT NULL in initial engine migration)
ALTER TABLE form_fields ALTER COLUMN form_id DROP NOT NULL;

-- Fix form_fields: already has form_id nullable + template_id, but constraints are wrong
-- Drop old unique constraints that don't handle NULL correctly
ALTER TABLE form_fields DROP CONSTRAINT IF EXISTS uq_form_fields_form_id_key;
ALTER TABLE form_fields DROP CONSTRAINT IF EXISTS uq_form_fields_template_id_key;

-- Replace with partial unique indexes (NULLs are properly excluded)
CREATE UNIQUE INDEX uq_form_fields_form_id_key ON form_fields(form_id, key) WHERE form_id IS NOT NULL;
CREATE UNIQUE INDEX uq_form_fields_template_id_key ON form_fields(template_id, key) WHERE template_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS uq_form_fields_template_id_key;
DROP INDEX IF EXISTS uq_form_fields_form_id_key;
ALTER TABLE form_fields ADD CONSTRAINT uq_form_fields_form_id_key UNIQUE (form_id, key);
ALTER TABLE form_fields ADD CONSTRAINT uq_form_fields_template_id_key UNIQUE (template_id, key);

ALTER TABLE form_sections DROP CONSTRAINT IF EXISTS chk_form_sections_owner;
DROP INDEX IF EXISTS idx_form_sections_template_id;
ALTER TABLE form_sections DROP COLUMN IF EXISTS template_id;
ALTER TABLE form_sections ALTER COLUMN form_id SET NOT NULL;
-- +goose StatementEnd
