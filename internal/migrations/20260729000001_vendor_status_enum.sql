-- +goose Up
-- +goose StatementBegin
CREATE TYPE vendor_status AS ENUM ('pending', 'active', 'inactive', 'suspended');

ALTER TABLE vendors
  ALTER COLUMN status DROP DEFAULT,
  ALTER COLUMN status TYPE vendor_status USING status::vendor_status,
  ALTER COLUMN status SET DEFAULT 'pending';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vendors
  ALTER COLUMN status DROP DEFAULT,
  ALTER COLUMN status TYPE VARCHAR(20) USING status::text,
  ALTER COLUMN status SET DEFAULT 'pending';

DROP TYPE IF EXISTS vendor_status;
-- +goose StatementEnd
