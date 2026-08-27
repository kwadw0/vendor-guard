create table "roles" (
  "id"          uuid         primary key default gen_random_uuid(),
  "name"        varchar(255) not null unique,
  "description" text         null,
  "created_at"  timestamptz  not null default now(),
  "updated_at"  timestamptz  not null default now()
);

insert into "roles" ("name", "description") values
  ('owner',  'Full access to the organization'),
  ('admin',  'Can manage members and settings'),
  ('member', 'Standard access'),
  ('manager','Can manage resources across the platform'),
  ('viewer', 'Read-only access across assigned domains');

create trigger trg_roles_updated_at
  before update on "roles"
  for each row execute function set_updated_at();

-- ============================================================
-- ORGANIZATIONS
-- ============================================================

create table "organizations" (
  "id"                    uuid         primary key default gen_random_uuid(),
  "name"                  varchar(255) not null unique,
  "description"           text         null,
  "website_url"           varchar(255) null,
  "industry"              varchar(255) null check (industry in (
                            'fashion_and_apparel',
                            'food_and_beverage',
                            'beauty_and_cosmetics',
                            'health_and_wellness',
                            'retail_and_ecommerce',
                            'professional_services',
                            'education_and_tutoring',
                            'real_estate',
                            'logistics_and_delivery',
                            'events_and_entertainment',
                            'tech_and_software',
                            'finance_and_fintech',
                            'agriculture',
                            'travel_and_tourism',
                            'manufacturing',
                            'cybersecurity',
                            'cloud_services',
                            'legal',
                            'consulting',
                            'other'
                          )),
  "team_size"             varchar(10)  null check (team_size in (
                            '1',
                            '2-5',
                            '6-10',
                            '11-50',
                            '51+'
                          )),
  "primary_customer_type" varchar(10)  null check (primary_customer_type in (
                            'b2b',
                            'b2c',
                            'both'
                          )),

  "owner_role"            varchar(255) not null check (owner_role in (
                            'business_owner',
                            'sales_manager',
                            'customer_support_manager',
                            'marketing_manager',
                            'operations_manager',
                            'freelancer_or_consultant',
                            'developer_or_technical',
                            'other'
                          )),

  "is_active"             boolean      not null default true,
  "created_at"            timestamptz  not null default now(),
  "updated_at"            timestamptz  not null default now()
);

create index idx_organizations_name      on "organizations"("name");
create index idx_organizations_industry  on "organizations"("industry");
create index idx_organizations_is_active on "organizations"("is_active");

create trigger trg_organizations_updated_at
  before update on "organizations"
  for each row execute function set_updated_at();

create table "users" (
  "id"              uuid         primary key default gen_random_uuid(),
  "first_name"      varchar(255) not null,
  "last_name"       varchar(255) not null,
  "email"           varchar(255) not null unique,
  "password"        varchar(255) not null,
  "phone"           varchar(255) not null unique,
  "organization_id"  uuid         null references organizations(id),
  "partner_id"       uuid         null,
  "role_id"         uuid         not null references roles(id),
  "avatar_url"      varchar(255) null,
  "is_active"       boolean      not null default true,
  "email_verified"  boolean      not null default false,
  "phone_verified"  boolean      not null default false,
  "refresh_token"   text        null,
  "refresh_token_expires_at" timestamptz null,
  "created_at"      timestamptz  not null default now(),
  "updated_at"      timestamptz  not null default now()
);

create index idx_users_email           on "users"("email");
create index idx_users_phone           on "users"("phone");
create index idx_users_role_id         on "users"("role_id");

create trigger trg_users_updated_at
  before update on "users"
  for each row execute function set_updated_at();


CREATE TABLE partners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),

    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL unique,
    phone VARCHAR(255) null,

    status VARCHAR(20) DEFAULT 'pending',

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

create index idx_partners_email          on "partners"("email");
CREATE INDEX idx_partners_org_id         on "partners"("organization_id");

create trigger trg_partners_updated_at
  before update on "partners"
  for each row execute function set_updated_at();

CREATE TABLE partner_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL,
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    invited_by UUID NOT NULL,
    role_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_partner_invitations_partner_id ON partner_invitations(partner_id);
CREATE INDEX idx_partner_invitations_email ON partner_invitations(email);
CREATE INDEX idx_partner_invitations_token ON partner_invitations(token);

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
-- FORM SECTIONS (organizes fields within a form or template)
-- ============================================================
CREATE TABLE form_sections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_id UUID REFERENCES forms(id) ON DELETE CASCADE,
    template_id UUID REFERENCES form_templates(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_form_sections_owner CHECK ((form_id IS NULL) != (template_id IS NULL))
);

CREATE INDEX idx_form_sections_form_id ON form_sections(form_id);
CREATE INDEX idx_form_sections_template_id ON form_sections(template_id);

CREATE TRIGGER trg_form_sections_updated_at
  BEFORE UPDATE ON form_sections
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- FORM FIELDS (under forms or templates)
-- ============================================================
CREATE TABLE form_fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_id UUID REFERENCES forms(id) ON DELETE CASCADE,
    template_id UUID REFERENCES form_templates(id) ON DELETE CASCADE,
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
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_form_fields_form_id_key ON form_fields(form_id, key) WHERE form_id IS NOT NULL;
CREATE UNIQUE INDEX uq_form_fields_template_id_key ON form_fields(template_id, key) WHERE template_id IS NOT NULL;
CREATE INDEX idx_form_fields_form_id ON form_fields(form_id);
CREATE INDEX idx_form_fields_template_id ON form_fields(template_id);
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
