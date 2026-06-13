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



create table "users" (
  "id"              uuid         primary key default gen_random_uuid(),
  "first_name"      varchar(255) not null,
  "last_name"       varchar(255) not null,
  "email"           varchar(255) not null unique,
  "password"        varchar(255) not null,
  "phone"           varchar(255) not null unique,
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
