---
name: new-endpoint
description: Scaffold a new API endpoint following project conventions
---

When creating a new endpoint, follow the existing handler -> service -> repository pattern:

1. **Model** (`internal/model/`): Add the MongoDB struct, request DTO(s), and response DTO(s) in one file. Use `bson:"snake_case"` and `json:"camelCase"` tags. Use pointer fields (`*string`, `*int`) for optional/partial-update request fields.

2. **Repository** (`internal/repository/`): Define an interface and unexported struct implementation in the same file. Return `nil, nil` for not-found (not an error). Constructor: `NewXxxRepository(db *mongo.Database)`.

3. **Service** (`internal/service/`): Accept `*auth.Client` + repository interfaces. Use `authenticateUser()` from `helpers.go` when the service needs to verify tokens directly. Use `context.Context` as the first parameter on all methods.

4. **Handler** (`internal/handler/`): Parse request, create `context.WithTimeout(ctx, 10*time.Second)`, call service, map errors to HTTP status via `switch err.Error()`. Return `model.ErrorResponse{Message: "..."}` for errors. Add Swagger godoc annotations.

5. **Wire up** in `internal/router/router.go`: instantiate repo, service, handler, then add routes with the appropriate middleware chain.

6. **Regenerate Swagger**: run `/swagger`
