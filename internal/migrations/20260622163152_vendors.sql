-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
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

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS partners CASCADE;
SELECT 'down SQL query';
-- +goose StatementEnd
