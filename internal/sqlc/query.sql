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
  password = $5,
  phone = $6,
  role_id = $7,
  avatar_url = $8
WHERE id = $1 RETURNING *;

-- name: DeleteUser :one
DELETE FROM users WHERE id = $1 RETURNING *;

-- name: ListUsers :many
SELECT * FROM users;

-- -- name: UpdateUserOrganization :one
-- UPDATE users SET
--   organization_id = $2
-- WHERE id = $1 RETURNING *;

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