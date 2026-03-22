---
name: swagger
description: Regenerate Swagger/OpenAPI docs from godoc annotations
---

Run: `swag init -g cmd/server/main.go -o ./api/docs`

After generating, verify the output exists at `api/docs/docs.go`.
