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
  ('member', 'Standard access');

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

