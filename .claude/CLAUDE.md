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
