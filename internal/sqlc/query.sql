-- name: CreateUser :one
INSERT INTO users (
  first_name,
  last_name,
  email,
  password,
  phone,
  role_id,
  avatar_url
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUser :one
UPDATE users SET
  first_name = $2,
  last_name = $3,
  email = $4,
  phone = $5,
  role_id = $6,
  avatar_url = $7
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users SET
  password = $2
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT *
FROM users
ORDER BY created_at DESC
LIMIT $1
OFFSET $2;

-- name: UpdateUserOrganization :one
UPDATE users SET
  organization_id = $2
WHERE id = $1 RETURNING *;

-- name: UpdateUserPartner :one
UPDATE users SET
  partner_id = $2
WHERE id = $1 RETURNING *;

-- name: UpdateUserRefreshToken :one
UPDATE users SET
  refresh_token = $2,
  refresh_token_expires_at = $3
WHERE id = $1 RETURNING *;

-- name: GetUserByRefreshToken :one
SELECT * FROM users
WHERE refresh_token = $1
  AND refresh_token_expires_at > now();

-- name: RevokeRefreshToken :exec
UPDATE users SET
  refresh_token = NULL,
  refresh_token_expires_at = NULL
WHERE id = $1;



-- ============================================================
-- ROLES
-- ============================================================

-- name: CreateRole :one
INSERT INTO roles (
  name,
  description
) VALUES (
  $1,
  $2
) RETURNING *;

-- name: GetRoleByID :one
SELECT * FROM roles WHERE id = $1;

-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = $1;

-- name: ListRoles :many
SELECT * FROM roles;

-- name: UpdateRole :one
UPDATE roles SET
  name = $2,
  description = $3
WHERE id = $1 RETURNING *;

-- name: DeleteRole :one
DELETE FROM roles WHERE id = $1 RETURNING *;

-- name: CreateOrganization :one
INSERT INTO organizations (
  name,
  description,
  website_url,
  industry,
  team_size,
  primary_customer_type,
  owner_role
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7
) RETURNING *;

-- name: GetOrganizationById :one
SELECT * FROM organizations WHERE id = $1;

-- name: GetAllOrganizations :many
SELECT * FROM organizations;

-- name: GetOrganizationByUserID :one
SELECT o.* FROM organizations o
INNER JOIN users u ON u.organization_id = o.id
WHERE u.id = $1::uuid;

-- name: UpdateOrganization :one
UPDATE organizations SET
  name = $2,
  description = $3,
  website_url = $4,
  industry = $5,
  team_size = $6,
  primary_customer_type = $7,
  owner_role = $8
  WHERE id = $1
  RETURNING *;
 
-- name: DeleteOrganization :exec
DELETE FROM organizations WHERE id = $1;

-- name: GetPartnerById :one
SELECT * FROM partners WHERE id = $1;

-- name: GetAllPartners :many
SELECT * FROM partners
ORDER BY created_at DESC;

-- name: CreatePartners :one 
INSERT INTO partners (
  organization_id,
  name,
  email,
  phone
) VALUES (
  $1,
  $2,
  $3,
  $4
) RETURNING *;

-- name: UpdatePartner :one

UPDATE partners SET
  name=$2,
  email=$3,
  phone=$4
WHERE id = $1
RETURNING *;

-- name: DeletePartner :exec
DELETE FROM partners 
WHERE id = $1;

-- name: GetPartnerByUserID :one
SELECT p.* FROM partners p
INNER JOIN users u ON u.partner_id = p.id
WHERE u.id = $1::uuid;

-- name: GetPartnersByOrg :many
SELECT * FROM partners
WHERE organization_id = $1
ORDER BY created_at DESC;

-- name: CreatePartnerInvitation :one
INSERT INTO partner_invitations (
  partner_id,
  email,
  token,
  invited_by,
  role_id,
  expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetPartnerInvitationByToken :one
SELECT * FROM partner_invitations WHERE token = $1;

-- name: GetPartnerInvitationsByPartner :many
SELECT * FROM partner_invitations WHERE partner_id = $1 ORDER BY created_at DESC;

-- name: UpdatePartnerInvitationStatus :one
UPDATE partner_invitations SET status = $2 WHERE id = $1 RETURNING *;


-- ============================================================
-- FORM TEMPLATES
-- ============================================================

-- name: CreateFormTemplate :one
INSERT INTO form_templates (
  title,
  description,
  category
) VALUES (
  $1,
  $2,
  $3
) RETURNING *;

-- name: GetFormTemplateByID :one
SELECT * FROM form_templates WHERE id = $1;

-- name: GetAllFormTemplates :many
SELECT * FROM form_templates
WHERE is_active = true
ORDER BY created_at DESC;

-- name: UpdateFormTemplate :one
UPDATE form_templates SET
  title = $2,
  description = $3,
  category = $4,
  is_active = $5
WHERE id = $1
RETURNING *;

-- name: DeleteFormTemplate :exec
DELETE FROM form_templates WHERE id = $1;


-- ============================================================
-- FORMS
-- ============================================================

-- name: CreateForm :one
INSERT INTO forms (
  organization_id,
  template_id,
  title,
  description,
  status
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5
) RETURNING *;

-- name: GetFormByID :one
SELECT * FROM forms WHERE id = $1;

-- name: GetFormsByOrg :many
SELECT * FROM forms
WHERE organization_id = $1
ORDER BY created_at DESC;

-- name: UpdateForm :one
UPDATE forms SET
  title = $2,
  description = $3,
  status = $4
WHERE id = $1
RETURNING *;

-- name: DeleteForm :exec
DELETE FROM forms WHERE id = $1;

-- name: CloneTemplateToForm :one
INSERT INTO forms (
  organization_id,
  template_id,
  title,
  description,
  status
)
SELECT
  $1,
  $2,
  $3,
  ft.description,
  'draft'
FROM form_templates ft
WHERE ft.id = $2
RETURNING *;


-- ============================================================
-- FORM FIELDS
-- ============================================================

-- name: CreateFormField :one
INSERT INTO form_fields (
  form_id,
  section_id,
  field_type,
  label,
  key,
  description,
  placeholder,
  is_required,
  sort_order,
  validation,
  options
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11
) RETURNING *;

-- name: GetFormFieldByID :one
SELECT * FROM form_fields WHERE id = $1;

-- name: GetFormFieldsByFormID :many
SELECT * FROM form_fields
WHERE form_id = $1
ORDER BY sort_order ASC;

-- name: GetFormFieldsBySectionID :many
SELECT * FROM form_fields
WHERE section_id = $1
ORDER BY sort_order ASC;

-- name: UpdateFormField :one
UPDATE form_fields SET
  field_type = $2,
  label = $3,
  key = $4,
  description = $5,
  placeholder = $6,
  is_required = $7,
  sort_order = $8,
  validation = $9,
  options = $10
WHERE id = $1
RETURNING *;

-- name: DeleteFormField :exec
DELETE FROM form_fields WHERE id = $1;


-- ============================================================
-- FORM SUBMISSIONS
-- ============================================================

-- name: CreateFormSubmission :one
INSERT INTO form_submissions (
  form_id,
  partner_id,
  submitted_by,
  status,
  responses,
  submitted_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6
) RETURNING *;

-- name: GetFormSubmissionByID :one
SELECT * FROM form_submissions WHERE id = $1;

-- name: GetFormSubmissionsByFormID :many
SELECT * FROM form_submissions
WHERE form_id = $1
ORDER BY created_at DESC;

-- name: GetFormSubmissionsByPartnerID :many
SELECT * FROM form_submissions
WHERE partner_id = $1
ORDER BY created_at DESC;

-- name: UpdateFormSubmission :one
UPDATE form_submissions SET
  status = $2,
  responses = $3
WHERE id = $1
RETURNING *;

-- name: ReviewFormSubmission :one
UPDATE form_submissions SET
  status = $2,
  reviewed_by = $3,
  reviewed_at = now()
WHERE id = $1
RETURNING *;
