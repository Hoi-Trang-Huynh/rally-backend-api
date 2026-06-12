# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Quick Reference

- `make run` / `make dev` (hot reload) / `make lint` / `go test ./...`
- `/build` — build binary with version info
- `/swagger` — regenerate OpenAPI docs
- `/new-endpoint` — scaffold following project conventions

## Architecture

GoFiber REST API with **handler -> service -> repository** layers. All code in `internal/`.

- **DI wiring & routes**: `internal/router/router.go` — single place to add routes or wire dependencies
- **Models**: each file bundles MongoDB struct + request DTOs + response DTOs (e.g. `rally.go`)
- **Repositories**: interface + unexported impl in same file; return `nil, nil` for not-found
- **Services**: business logic; take repo interfaces and the resolved `*model.User` — never raw tokens (verification happens once, in middleware)
- **Handlers**: parse request, `context.WithTimeout(10s)`, call service, map errors via `switch err.Error()`

## Auth Middleware Chain

```
auth -> resolveUser -> loadParticipant -> requireJoined -> requireRole
```

- `auth` (`AuthRequired`) verifies the Firebase ID token from the Authorization header — the ONLY place tokens are verified — and stores the verified `*auth.Token` in `c.Locals("authToken")`.
- `resolveUser` (`ResolveFirebaseUser`) JIT-provisions/loads the MongoDB user from the verified claims (syncing `email`/`email_verified` from the token, which is their source of truth) and stores `*model.User` in `c.Locals("user")`.
- `POST /auth/register` is an idempotent register-or-login: middleware provisions the user; the optional body only carries initial profile fields. Identity/trust fields (`email`, `isEmailVerified`, `isActive`) are never accepted from clients.
- Unique indexes on `firebase_uid` and non-empty `username` (created at startup via `repository.EnsureUserIndexes`) make provisioning and username selection race-free.

Event/activity routes skip the rally part of this chain and validate access in their service layer via `validateRallyAccess()`.

## Key Details

- Two MongoDB databases on one client: `rally_db` (main) and `rally_dashboard` (feedback only), initialized as singletons in `internal/infrastructure/database/database.go`
- JSON: camelCase, BSON: snake_case. IDs are ObjectID hex strings in responses.
- Partial updates use pointer fields — nil means "don't update"
- Required env: `MONGODB_URI`, `MONGODB_DB`, `FIREBASE_CREDENTIALS_PATH`, `CLOUDINARY_URL`. Optional: `ALLOWED_ORIGINS` (comma-separated CORS origins; defaults to `*` for development — set it in production)

## Versioning & Releases

Full guide: [`VERSIONING.md`](../VERSIONING.md). Quick rules:

- **Production** = push a `v*` git tag **created on `master`** → builds & deploys to Cloud Run prod (`ENV=production`). The tag string becomes the binary version (`internal/version`).
- **Staging** = push to the **`staging`** branch → deploys to Cloud Run staging (version `dev`).
- The workflow ([`.github/workflows/cicd.yml`](../.github/workflows/cicd.yml)) has **no branch guard** on tags — only ever tag commits that are on `master`, or you'll ship un-reviewed code to prod.
- **SemVer `vMAJOR.MINOR.PATCH`**: bump **`z`** (patch) for bug fixes / small changes, **`y`** (minor) for bigger changes / new features (reset patch to 0), **`x`** (major) for breaking changes.
- **Bump the version before opening a PR** — if you don't, a PR check will remind you in the comments.
- Tags are immutable: never move or re-push an existing tag, cut a new one.

## Coding Style & Conventions

### Naming
- Files: `{resource}_handler.go`, `{resource}_service.go`, `{resource}_repository.go`
- Models: `{resource}.go` bundling struct + request/response DTOs
- Public functions: `CamelCase` — `CreateRally()`, `GetUserByID()`
- Unexported: `camelCase` — `userService`, `eventRepository`
- Error messages: lowercase descriptive strings — `"user not found"`, `"rally not found"`

### Pagination
- Offset-based: query params `page` (1-indexed, default 1) and `pageSize` (default 20, max 50)
- Use `utils.ClampPagination(page, pageSize, maxPageSize)` to normalize inputs
- Use `utils.CalcTotalPages(total, pageSize)` for total pages
- List responses always include: `total`, `page`, `pageSize`, `totalPages` alongside the items array

### Error Handling
- Standard `errors.New()` / `fmt.Errorf("%w", err)` — no custom error types
- Handlers map errors to HTTP status via `switch err.Error()` string matching
- Common error strings: `"user not found"` (404), `"unauthorized: insufficient permissions"` (403), `"invalid or expired token"` (401)
- Error response: `model.ErrorResponse{Message: "..."}` — single `message` field
- Default to 500 with generic message for unexpected errors

### Request Handling
- Parse with `c.BodyParser(&req)` → return 400 on failure
- Validate required fields manually in handler after parsing
- Pointer fields (`*string`, `*int`) for optional/partial update fields — nil = don't update
- `context.WithTimeout(ctx, 10*time.Second)` before calling service

### Response Format
- Single items: return struct directly with `c.Status(fiber.StatusOK).JSON(response)`
- Lists: envelope with items array + pagination metadata (`total`, `page`, `pageSize`, `totalPages`)
- Success actions: `201 Created` for POST, `200 OK` for GET/PUT
- Use `ConvertTo{Resource}Response()` helpers in services to transform models → DTOs

### MongoDB Patterns
- ObjectIDs: store as `primitive.ObjectID`, convert to hex string in responses via `.Hex()`
- Accept string IDs from API, convert with `primitive.ObjectIDFromHex(id)`
- Not-found: `mongo.ErrNoDocuments` → return `nil, nil` (not an error)
- Filters: `bson.M{...}`, case-insensitive search with `primitive.Regex{Pattern: q, Options: "i"}`
- Partial updates: build `bson.M{"$set": updateDoc}` only including non-nil fields
- Counters: `bson.M{"$inc": bson.M{"field": 1}}`
- Transactions: `session.WithTransaction()` for multi-document operations (e.g. create rally + add owner participant)

### Middleware Access
- After auth chain, access via `c.Locals()`: `c.Locals("authToken").(*auth.Token)`, `c.Locals("user").(*model.User)`, `c.Locals("rallyParticipant").(*model.RallyParticipant)`
- Handlers pass `*model.User` into services; services never see or verify tokens
- Event/activity routes skip rally middleware — validate access in service via `validateRallyAccess()`

### DI Pattern
- Constructor injection: `NewUserService(firebaseAuth, userRepo, ...)` returning struct pointer
- All wiring in `internal/router/router.go`: repos → services → handlers → routes
