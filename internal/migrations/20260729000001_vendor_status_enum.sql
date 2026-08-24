-- +goose Up
-- +goose StatementBegin
CREATE TYPE partner_status AS ENUM ('pending', 'active', 'inactive', 'suspended');

ALTER TABLE partners
  ALTER COLUMN status DROP DEFAULT,
  ALTER COLUMN status TYPE partner_status USING status::partner_status,
  ALTER COLUMN status SET DEFAULT 'pending';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE partners
  ALTER COLUMN status DROP DEFAULT,
  ALTER COLUMN status TYPE VARCHAR(20) USING status::text,
  ALTER COLUMN status SET DEFAULT 'pending';

DROP TYPE IF EXISTS partner_status;
-- +goose StatementEnd
