-- +goose Up
-- +goose StatementBegin

-- Per-form resubmission policy (default: one submission per assignment)
ALTER TABLE forms ADD COLUMN allow_resubmit BOOLEAN NOT NULL DEFAULT false;

-- Form assignments: which vendors may fill which forms (the permission slip
-- that CreateSubmission checks before inserting a form_submissions row)
CREATE TABLE form_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_id UUID NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    partner_id UUID NOT NULL REFERENCES partners(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'assigned'
        CHECK (status IN ('assigned', 'submitted', 'revoked')),
    assigned_by UUID REFERENCES users(id) ON DELETE SET NULL,
    due_at TIMESTAMPTZ,
    submitted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One live assignment per pair; revoked rows stay as history, so a
-- revoke-then-reassign cycle can create a fresh row for the same pair.
CREATE UNIQUE INDEX uq_form_assignments_live
    ON form_assignments(form_id, partner_id) WHERE status != 'revoked';

CREATE INDEX idx_form_assignments_form_id ON form_assignments(form_id);
CREATE INDEX idx_form_assignments_partner_id ON form_assignments(partner_id);
CREATE INDEX idx_form_assignments_status ON form_assignments(status);

CREATE TRIGGER trg_form_assignments_updated_at
  BEFORE UPDATE ON form_assignments
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Backfill: historic submissions imply an assignment (keeps old flows working
-- once the submit guard goes live). Oldest submission wins submitted_at.
INSERT INTO form_assignments (form_id, partner_id, status, submitted_at)
SELECT form_id, partner_id, 'submitted', MIN(submitted_at)
FROM form_submissions
GROUP BY form_id, partner_id
ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_form_assignments_updated_at ON form_assignments;
DROP TABLE IF EXISTS form_assignments CASCADE;
ALTER TABLE forms DROP COLUMN IF EXISTS allow_resubmit;
-- +goose StatementEnd
