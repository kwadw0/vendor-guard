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