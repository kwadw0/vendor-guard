-- +goose Up
-- +goose StatementBegin

CREATE TABLE organizations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(255) NOT NULL UNIQUE,
  description text NULL,
  website_url varchar(255) NULL,
  industry varchar(255) NULL CHECK (industry IN (
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
  team_size varchar(10) NULL CHECK (team_size IN (
    '1',
    '2-5',
    '6-10',
    '11-50',
    '51+'
  )),
  primary_customer_type varchar(10) NULL CHECK (primary_customer_type IN (
    'b2b',
    'b2c',
    'both'
  )),
  owner_role varchar(255) NOT NULL CHECK (owner_role IN (
    'business_owner',
    'sales_manager',
    'customer_support_manager',
    'marketing_manager',
    'operations_manager',
    'freelancer_or_consultant',
    'developer_or_technical',
    'other'
  )),
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_organizations_name ON organizations(name);
CREATE INDEX idx_organizations_industry ON organizations(industry);
CREATE INDEX idx_organizations_is_active ON organizations(is_active);

CREATE TRIGGER trg_organizations_updated_at
BEFORE UPDATE ON organizations
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE organizations CASCADE;
-- +goose StatementEnd
