---
name: build
description: Build the server binary with version info
---

Build with embedded version metadata:

```bash
go build -ldflags "-X github.com/Hoi-Trang-Huynh/rally-backend-api/internal/version.Version=dev \
  -X github.com/Hoi-Trang-Huynh/rally-backend-api/internal/version.CommitSHA=$(git rev-parse --short HEAD) \
  -X github.com/Hoi-Trang-Huynh/rally-backend-api/internal/version.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o bin/server ./cmd/server
```

Verify: `./bin/server` or `ls -lh bin/server`
