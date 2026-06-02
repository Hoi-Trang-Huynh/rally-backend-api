# Versioning & Releases — rally-backend-api

This service uses **[Semantic Versioning](https://semver.org/)** (`vMAJOR.MINOR.PATCH`, e.g. `v0.0.2`).
CI/CD reads the git **tag** to decide what to build and where to deploy.

> TL;DR
> - **Production** deploy = push a `v*` git tag **on `master`**.
> - **Staging** deploy = push to the **`staging`** branch.
> - Bump the version **before opening a PR**. If you forget, a PR check will remind you.

---

## How CI/CD is triggered

See [`.github/workflows/cicd.yml`](.github/workflows/cicd.yml). The pipeline triggers on:

| Trigger | Environment | `ENV` value | What happens |
| --- | --- | --- | --- |
| Push a `v*` **tag** | `production` | `production` | Builds binary with the tag as its version, pushes Docker image (`:latest`, `:<sha>`, `:<version>`), deploys to Cloud Run prod. |
| Push to **`staging`** branch | `staging` | `staging` | Builds with a derived version (`git describe`, e.g. `v0.0.2-5-gabc1234`), pushes image, deploys to Cloud Run staging. |
| `workflow_dispatch` | — | — | Manual run for debugging. |

The deployed binary embeds version info via `-ldflags` into `internal/version` and logs it on
startup (`Starting Rally Backend API <version>`).

### ⚠️ Tags must be created on `master`

GitHub fires the tag workflow for **any** `v*` tag, regardless of which commit it points at.
There is no branch guard in the workflow, so this is a **team convention you must follow**:

> **Only ever create a production `v*` tag from a commit that is on `master`.**

Tagging a feature/staging commit would ship un-reviewed code straight to production.

---

## Releasing to production (step by step)

1. Merge your PR into `master` (via the normal review flow).
2. Make sure your local `master` is up to date:
   ```bash
   git checkout master
   git pull origin master
   ```
3. Pick the next version (see [How to bump](#how-to-bump-the-version) below) and tag the
   tip of `master`:
   ```bash
   git tag v0.0.3
   git push origin v0.0.3
   ```
4. Watch the **CI Rally Backend API** workflow run → it builds, pushes the image, and deploys
   to production. A Teams notification reports success/failure.

To deploy to **staging**, just push to the `staging` branch — no tag needed.

---

## How to bump the version

Format: `vMAJOR.MINOR.PATCH` — bump the right segment and reset the ones to its right to `0`.

| Segment | When to bump | Example |
| --- | --- | --- |
| **PATCH** (`z`) | Bug fixes and small changes — no new behavior for clients. | `v0.0.2` → `v0.0.3` |
| **MINOR** (`y`) | Bigger changes / new features that are backwards-compatible. Reset patch to 0. | `v0.1.4` → `v0.2.0` |
| **MAJOR** (`x`) | Breaking API changes (clients must change to keep working). Reset minor & patch to 0. | `v1.4.2` → `v2.0.0` |

Rules of thumb:
- **Small change or bug fix → bump `z`** (patch).
- **Big change / new feature → bump `y`** (minor).
- While the service is pre-1.0 (`v0.x.y`), we still follow the same intent: `z` for fixes,
  `y` for features.
- A version is **immutable** once tagged and pushed. Never move or re-push an existing tag —
  cut a new one instead.

---

## Bump *before* the PR

**Decide and record the new version as part of the PR that introduces the change**, not after
merge. This keeps the version in lockstep with the code being reviewed and avoids "what version
is this?" confusion at release time.

> If you open a PR without bumping the version, **you'll be reminded in the PR comments anyway** —
> save the round-trip and do it up front.

---

## Checking the running version

The build stamps `version.Version`, `version.CommitSHA`, and `version.BuildTime` into the binary.
Check the Cloud Run startup logs:

```
Starting Rally Backend API v0.0.3
Commit SHA: a1b2c3d, Build Time: 2026-06-02T10:00:00Z
```

### Staging version format

Staging builds aren't tagged, so CI derives the version with `git describe --tags --always`:

```
v0.0.2-5-gabc1234
└──┬──┘ │  └──┬──┘
 last   │   short commit sha
 tag    commits since that tag
```

Read it as "**5 commits past `v0.0.2`, at commit `abc1234`**" — so on the dashboard you can
see which release a staging build is ahead of and by how much, instead of a bare `dev`.

A local `make build` / `make run` reports version `dev` (no ldflags injected), which is expected.
