# Graph Report - /home/qhojo23/vendor-guard  (2026-07-27)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 214 nodes · 493 edges · 18 communities (11 shown, 7 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 74 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- ErrorJSON
- AuthService
- OrganizationService
- VendorService
- New
- application
- query.sql.go
- UserService
- User
- UUID
- Queries
- Context
- UpdateUserRefreshTokenParams
- check_db.sh
- querier.go
- .ListUsers
- vendor-guard

## God Nodes (most connected - your core abstractions)
1. `Queries` - 28 edges
2. `ErrorJSON()` - 22 edges
3. `WriteJSON()` - 21 edges
4. `New()` - 16 edges
5. `User` - 14 edges
6. `ReadJSON()` - 12 edges
7. `OrganizationService` - 11 edges
8. `Organization` - 10 edges
9. `Role` - 10 edges
10. `OrganizationHandler` - 10 edges

## Surprising Connections (you probably didn't know these)
- `RequireAuth()` --calls--> `ValidateToken()`  [INFERRED]
  middleware/auth.go → auth/jwt/jwt.go
- `main()` --calls--> `New()`  [INFERRED]
  cmd/main.go → internal/repo/db.go
- `RequireAuth()` --calls--> `ErrorJSON()`  [INFERRED]
  middleware/auth.go → utils/json.go
- `GenerateTokenPair()` --calls--> `New()`  [INFERRED]
  auth/jwt/jwt.go → internal/repo/db.go
- `ValidateToken()` --calls--> `New()`  [INFERRED]
  auth/jwt/jwt.go → internal/repo/db.go

## Import Cycles
- None detected.

## Communities (18 total, 7 thin omitted)

### Community 0 - "ErrorJSON"
Cohesion: 0.13
Nodes (23): Request, ResponseWriter, GetUserID(), Request, Request, ResponseWriter, OrganizationHandler, Handler (+15 more)

### Community 1 - "AuthService"
Cohesion: 0.18
Nodes (13): Validate, NewHandler(), Context, Queries, NewService(), AuthService, Handler, LoginDto (+5 more)

### Community 2 - "OrganizationService"
Cohesion: 0.20
Nodes (13): CreateOrganizationDto, Time, UUID, Context, Queries, Text, UUID, mapToDto() (+5 more)

### Community 3 - "VendorService"
Cohesion: 0.20
Nodes (11): CreateVendorDto, UpdateVendorDto, VendorResponse, VendorResponseWrapper, Validate, NewVendorHandler(), Context, Queries (+3 more)

### Community 4 - "New"
Cohesion: 0.16
Nodes (13): GenerateTokenPair(), Time, ValidateRefreshToken(), ValidateToken(), main(), Queries, Queries, New() (+5 more)

### Community 5 - "application"
Cohesion: 0.14
Nodes (12): Handler, Validate, application, config, dbConfig, Logger, Handler, RequireAuth() (+4 more)

### Community 6 - "query.sql.go"
Cohesion: 0.20
Nodes (9): Text, CreateOrganizationParams, CreateRoleParams, CreateUserParams, CreateVendorsParams, UpdateOrganizationParams, UpdateRoleParams, UpdateUserParams (+1 more)

### Community 7 - "UserService"
Cohesion: 0.29
Nodes (8): CreateUserDto, UpdateUserDto, UserResponseDto, Context, Queries, mapUserToResponse(), NewService(), UserService

### Community 8 - "User"
Cohesion: 0.27
Nodes (5): Text, Timestamptz, UUID, User, Vendor

### Community 9 - "UUID"
Cohesion: 0.22
Nodes (3): UUID, UpdateUserOrganizationParams, UpdateUserPasswordParams

## Knowledge Gaps
- **7 isolated node(s):** `check_db.sh script`, `vendor-guard`, `Querier`, `contextKey`, `EmptyData` (+2 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **7 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `New()` connect `New` to `ErrorJSON`, `OrganizationService`, `application`?**
  _High betweenness centrality (0.328) - this node is a cross-community bridge._
- **Why does `mapToDto()` connect `OrganizationService` to `Context`?**
  _High betweenness centrality (0.211) - this node is a cross-community bridge._
- **Why does `Organization` connect `Context` to `User`, `OrganizationService`?**
  _High betweenness centrality (0.210) - this node is a cross-community bridge._
- **Are the 20 inferred relationships involving `ErrorJSON()` (e.g. with `.Login()` and `.RefreshToken()`) actually correct?**
  _`ErrorJSON()` has 20 INFERRED edges - model-reasoned connections that need verification._
- **Are the 19 inferred relationships involving `WriteJSON()` (e.g. with `.Login()` and `.RefreshToken()`) actually correct?**
  _`WriteJSON()` has 19 INFERRED edges - model-reasoned connections that need verification._
- **What connects `check_db.sh script`, `vendor-guard`, `Querier` to the rest of the system?**
  _7 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ErrorJSON` be split into smaller, more focused modules?**
  _Cohesion score 0.12684989429175475 - nodes in this community are weakly interconnected._