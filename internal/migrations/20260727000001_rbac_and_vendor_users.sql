-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

-- Add general-purpose roles (not scoped to any specific domain)
INSERT INTO roles (name, description) VALUES
  ('manager', 'Can manage resources across the platform'),
  ('viewer', 'Read-only access across assigned domains');

-- Add vendor_id to users (nullable FK)
-- A user with vendor_id is a vendor user; mutually exclusive with organization_id
ALTER TABLE users ADD COLUMN vendor_id UUID REFERENCES vendors(id) ON DELETE SET NULL;
CREATE INDEX idx_users_vendor_id ON users(vendor_id);

-- Vendor invitations table
CREATE TABLE vendor_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_id UUID NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    invited_by UUID NOT NULL REFERENCES users(id),
    role_id UUID NOT NULL REFERENCES roles(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'expired', 'cancelled')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vendor_invitations_vendor_id ON vendor_invitations(vendor_id);
CREATE INDEX idx_vendor_invitations_email ON vendor_invitations(email);
CREATE INDEX idx_vendor_invitations_token ON vendor_invitations(token);

CREATE TRIGGER trg_vendor_invitations_updated_at
  BEFORE UPDATE ON vendor_invitations
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_vendor_invitations_updated_at ON vendor_invitations;
DROP TABLE IF EXISTS vendor_invitations CASCADE;
DROP INDEX IF EXISTS idx_users_vendor_id;
ALTER TABLE users DROP COLUMN IF EXISTS vendor_id;
DELETE FROM roles WHERE name IN ('manager', 'viewer');
SELECT 'down SQL query';
-- +goose StatementEnd
