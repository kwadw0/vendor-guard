-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

-- Add general-purpose roles (not scoped to any specific domain)
INSERT INTO roles (name, description) VALUES
  ('manager', 'Can manage resources across the platform'),
  ('viewer', 'Read-only access across assigned domains');

-- Add partner_id to users (nullable FK)
-- A user with partner_id is a partner user; mutually exclusive with organization_id
ALTER TABLE users ADD COLUMN partner_id UUID REFERENCES partners(id) ON DELETE SET NULL;
CREATE INDEX idx_users_partner_id ON users(partner_id);

-- Partner invitations table
CREATE TABLE partner_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES partners(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    invited_by UUID NOT NULL REFERENCES users(id),
    role_id UUID NOT NULL REFERENCES roles(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'expired', 'cancelled')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_partner_invitations_partner_id ON partner_invitations(partner_id);
CREATE INDEX idx_partner_invitations_email ON partner_invitations(email);
CREATE INDEX idx_partner_invitations_token ON partner_invitations(token);

CREATE TRIGGER trg_partner_invitations_updated_at
  BEFORE UPDATE ON partner_invitations
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_partner_invitations_updated_at ON partner_invitations;
DROP TABLE IF EXISTS partner_invitations CASCADE;
DROP INDEX IF EXISTS idx_users_partner_id;
ALTER TABLE users DROP COLUMN IF EXISTS partner_id;
DELETE FROM roles WHERE name IN ('manager', 'viewer');
SELECT 'down SQL query';
-- +goose StatementEnd
