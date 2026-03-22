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
- **Services**: business logic; take `*auth.Client` + repo interfaces; shared `authenticateUser()` in `helpers.go`
- **Handlers**: parse request, `context.WithTimeout(10s)`, call service, map errors via `switch err.Error()`

## Rally Middleware Chain

```
auth -> resolveUser -> loadParticipant -> requireJoined -> requireRole
```

Stores in `c.Locals`: `"idToken"` -> `"user"` -> `"rallyParticipant"`. Event/activity routes skip this chain and validate access in their service layer via `validateRallyAccess()`.

## Key Details

- Two MongoDB databases on one client: `rally_db` (main) and `rally_dashboard` (feedback only), initialized as singletons in `internal/infrastructure/database/database.go`
- JSON: camelCase, BSON: snake_case. IDs are ObjectID hex strings in responses.
- Partial updates use pointer fields — nil means "don't update"
- CI deploys to GCP Cloud Run on `v*` tags
- Required env: `MONGODB_URI`, `MONGODB_DB`, `FIREBASE_CREDENTIALS_PATH`, `CLOUDINARY_URL`

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
- After auth chain, access via `c.Locals()`: `c.Locals("idToken").(string)`, `c.Locals("user").(*model.User)`, `c.Locals("rallyParticipant").(*model.RallyParticipant)`
- Event/activity routes skip rally middleware — validate access in service via `validateRallyAccess()`

### DI Pattern
- Constructor injection: `NewUserService(firebaseAuth, userRepo, ...)` returning struct pointer
- All wiring in `internal/router/router.go`: repos → services → handlers → routes
