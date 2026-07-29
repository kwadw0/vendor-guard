-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE vendors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),

    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL unique,
    phone VARCHAR(255) null,

    status VARCHAR(20) DEFAULT 'pending',

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

create index idx_vendors_email          on "vendors"("email");
CREATE INDEX idx_vendors_org_id         on "vendors"("organization_id");

create trigger trg_vendors_updated_at
  before update on "vendors"
  for each row execute function set_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS vendors CASCADE;
SELECT 'down SQL query';
-- +goose StatementEnd