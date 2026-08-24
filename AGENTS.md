# Preuvio — Agent Guide

## Code Standards
- Prioritize **clean, idiomatic Go** and **efficiency** in all changes — clear naming, small focused functions, minimal allocations, no dead code or unnecessary abstraction
- Prefer editing existing patterns over introducing new ones; keep the 3-layer separation strict (no business logic in handlers, no HTTP concerns in services)
- All API endpoints are documented via **Swagger** (swag annotations on handlers, regenerated with `swag init` command) — when adding or changing a route, always add/update the corresponding swagger annotations so `docs/index.html` stays in sync

## Stack
- **Go 1.26.2** (future/unreleased — may cause tooling issues), **go-chi/chi/v5**, **pgx/v5**, **sqlc** (codegen), **goose** (migrations)
- **JWT** with `golang-jwt/jwt/v5`, **validator** with `go-playground/validator/v10`
- No test files exist, no CI/CD, no Docker, no DI framework (manual wiring in `cmd/api.go`)

## Commands
- `go build -o preuvio ./cmd` — build binary
- `air` — hot reload (config: `.air.toml`)
- `sqlc generate` — regenerate `internal/repo/` from `internal/sqlc/query.sql` + `internal/migrations/`
- `swag init` — regenerate `docs/` (Swagger spec) from handler annotations — run this after any endpoint change
- `goose up` — run pending migrations (config in `.env`, uses Supabase pooler port **6543**)

## Architecture
- **3-layer per domain**: handler (HTTP/validation) → service (business logic) → repo (sqlc-generated)
- Domains: `auth/`, `users/`, `organizations/`, `partners/` — each with `*.handler.go`, `*.service.go`, `*.dto.go`
- JWT internals live in `auth/jwt/` (separate from middleware)
- Router assembled at `cmd/api.go:mount()` — all dependency injection happens here
- `go-chi` subrouters: no auth on `/api/users` or `/api/auth`, bearer required on `/api/organizations` (POST, GET /me) and all `/api/partners`

## API Documentation
- All endpoints must carry swag/Swagger annotations (`@Summary`, `@Router`, `@Param`, `@Success`, `@Failure`, etc.) directly above the handler function
- Source of truth is the generated `docs/` folder — never hand-edit it; change annotations then run `swag init`
- Treat missing or stale swagger annotations on a new/changed endpoint as incomplete work

## Response format
- Success: `{"success":true, "message":"...", "data":...}`
- Error: `{"success":false, "message":"...", "code":"UNAUTHORIZED|VALIDATION_ERROR|NOT_FOUND|EMAIL_EXISTS|..."}`

## File conventions
- Naming: dots as separators e.g. `auth.handler.go`, `org.service.go` (unusual for Go but consistent)
- Packages: lowercase, match directory name
- `middleware/auth.go` imports JWT package as `jwtutil` alias

## Database & codegen
- `.env` contains live Supabase credentials (committed — do not rotate)
- `goose` uses `GOOSE_DBSTRING` (port 6543), `sqlc` reads raw `DATABASE_URL` (port 5432)
- `internal/repo/` is **generated** — do not edit directly; modify `internal/sqlc/query.sql` or migration files, then re-run `sqlc generate`
- Migrations are in `internal/migrations/` with goose-compatible timestamps

## JWT
- Access: 15min, Refresh: 7d, both HS256
- Refresh token stored in `users.refresh_token` column
- Auth middleware injects `user_id` into request context; retrieve with `middleware.GetUserID(r)`

## Route protection quirks
- Org GET by ID, PUT, DELETE are **public** (no bearer)
- All user routes are **public**
- All partner routes require bearer
