# Agent guidelines

Guidance for AI coding agents (and humans) working in this repo. It encodes the
supply-chain and code-quality practices described in [SECURITY.md](SECURITY.md).

## Golden rules

- **Don't weaken the security steps.** Never disable or relax `golangci-lint`/`gosec`,
  `govulncheck`, `go mod verify`/tidy, ESLint, or the pnpm hardening settings to make
  something pass.
- **Don't commit secrets.** `.env` is git-ignored; keep real credentials out of the repo.
- **Commits must be signed** and use conventional messages (`feat:`, `fix:`, `chore:`, ...).
- **Run the full checks before claiming a task is done** (see below).

## Backend (Go) — `backend/`

- Prefer the standard library. Before adding a dependency, run `go mod graph` and reject
  ones that drag in large transitive trees. Avoid deprecated packages (e.g. `github.com/pkg/errors`).
- Commit both `go.mod` and `go.sum`. Never hand-edit `go.sum`.
- `go mod tidy` must produce no changes.
- Queries are typed via sqlc — edit SQL in `internal/db/{migrations,queries}` and
  regenerate; don't hand-write the generated files in `internal/db/*.sql.go`.
- Before done: `cd backend && make check` (lint + gosec, mod verify/tidy, govulncheck,
  unit + race tests, hardened build).

## Frontend (TypeScript/Vue) — `frontend/`

- Use **pnpm** (pinned via `mise`). Use `pnpm exec <tool>` — never `npx`, `pnpm dlx`, or
  global installs (`npm i -g`).
- Install with `pnpm install --frozen-lockfile`; commit `pnpm-lock.yaml`.
- Build/lifecycle scripts are **blocked by default** (`pnpm.onlyBuiltDependencies: []`).
  Only add a package to that allowlist if it genuinely needs a build script and you trust it.
- Supply-chain settings live in `pnpm-workspace.yaml` (`minimumReleaseAge`,
  `blockExoticSubdeps`) — don't remove them.
- Prefer built-ins/minimal libraries over dependencies for trivial utilities.
- Before done: `cd frontend && pnpm typecheck && pnpm lint && pnpm format:check && pnpm build`.

## Toolchain

- JS: Node + pnpm are pinned in `mise.toml` (+ `mise.lock`). Run `mise install`.
- Go: version + toolchain are pinned in `backend/go.mod`.
