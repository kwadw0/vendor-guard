-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- FORM TEMPLATES (global, reusable)
-- ============================================================
CREATE TABLE form_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_form_templates_category ON form_templates(category);
CREATE INDEX idx_form_templates_is_active ON form_templates(is_active);

CREATE TRIGGER trg_form_templates_updated_at
  BEFORE UPDATE ON form_templates
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- FORMS (org-scoped, cloned from templates or custom)
-- ============================================================
CREATE TABLE forms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    template_id UUID REFERENCES form_templates(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'active', 'archived')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_forms_organization_id ON forms(organization_id);
CREATE INDEX idx_forms_template_id ON forms(template_id);
CREATE INDEX idx_forms_status ON forms(status);

CREATE TRIGGER trg_forms_updated_at
  BEFORE UPDATE ON forms
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- FORM SECTIONS (organizes fields within a form)
-- ============================================================
CREATE TABLE form_sections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_id UUID NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_form_sections_form_id ON form_sections(form_id);

CREATE TRIGGER trg_form_sections_updated_at
  BEFORE UPDATE ON form_sections
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- FORM FIELDS (directly under forms, organized via section_id)
-- ============================================================
CREATE TABLE form_fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_id UUID NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    section_id UUID NOT NULL REFERENCES form_sections(id) ON DELETE CASCADE,
    field_type VARCHAR(50) NOT NULL CHECK (field_type IN (
        'text', 'textarea', 'number', 'email', 'phone',
        'date', 'checkbox', 'radio', 'select', 'multiselect',
        'file', 'richtext'
    )),
    label VARCHAR(255) NOT NULL,
    key VARCHAR(100) NOT NULL,
    description TEXT,
    placeholder VARCHAR(255),
    is_required BOOLEAN NOT NULL DEFAULT false,
    sort_order INT NOT NULL DEFAULT 0,
    validation JSONB DEFAULT '{}',
    options JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_form_fields_form_id_key UNIQUE (form_id, key)
);

CREATE INDEX idx_form_fields_form_id ON form_fields(form_id);
CREATE INDEX idx_form_fields_section_id ON form_fields(section_id);

CREATE TRIGGER trg_form_fields_updated_at
  BEFORE UPDATE ON form_fields
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- FORM SUBMISSIONS (partner responses)
-- ============================================================
CREATE TABLE form_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_id UUID NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    partner_id UUID NOT NULL REFERENCES partners(id) ON DELETE CASCADE,
    submitted_by UUID REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'in_review', 'approved', 'rejected', 'revision_required')),
    responses JSONB NOT NULL DEFAULT '{}',
    submitted_at TIMESTAMPTZ,
    reviewed_at TIMESTAMPTZ,
    reviewed_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_form_submissions_form_id ON form_submissions(form_id);
CREATE INDEX idx_form_submissions_partner_id ON form_submissions(partner_id);
CREATE INDEX idx_form_submissions_status ON form_submissions(status);

CREATE TRIGGER trg_form_submissions_updated_at
  BEFORE UPDATE ON form_submissions
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_form_submissions_updated_at ON form_submissions;
DROP TABLE IF EXISTS form_submissions CASCADE;
DROP TRIGGER IF EXISTS trg_form_fields_updated_at ON form_fields;
DROP TABLE IF EXISTS form_fields CASCADE;
DROP TRIGGER IF EXISTS trg_form_sections_updated_at ON form_sections;
DROP TABLE IF EXISTS form_sections CASCADE;
DROP TRIGGER IF EXISTS trg_forms_updated_at ON forms;
DROP TABLE IF EXISTS forms CASCADE;
DROP TRIGGER IF EXISTS trg_form_templates_updated_at ON form_templates;
DROP TABLE IF EXISTS form_templates CASCADE;
-- +goose StatementEnd
