# CLAUDE.md

Guidance for AI agents (and humans) working in this repo. Read this first.
Related docs: [AGENTS.md](AGENTS.md) (rules of engagement), [CONTRIBUTING.md](CONTRIBUTING.md)
(setup + pre-PR checklist), [SECURITY.md](SECURITY.md), and [README.md](README.md) (user-facing).

## What this is

A project boilerplate used **instead of a Laravel instance**: a Go API + Vue SPA with auth, sessions,
and the usual account pages. Two standalone, independently deployable services + Postgres + Redis.

| Layer    | Choice                                                                 |
|----------|-----------------------------------------------------------------------|
| Backend  | Go + Gin                                                              |
| Auth     | Server-side sessions in Redis, first-party HttpOnly cookie (bcrypt)   |
| Database | Postgres via **sqlc** (typed SQL) + an embedded SQL migrator         |
| Frontend | Vue 3 + TypeScript + Vite + Tailwind v4, vue-router + Pinia           |
| Cache    | Redis                                                                 |

**Single origin from the browser:** the web service serves the SPA *and* proxies `/api` to the Go
service (Vite dev proxy locally; nginx in prod). This keeps session cookies first-party while the API
stays a clean, standalone service (a future mobile client would consume it directly with token auth —
not built yet).

## Repository layout

```
backend/
  cmd/server/main.go         # entrypoint: load config, connect pgx pool, run migrations, sessions, gin
  internal/
    config/                  # env -> typed Config (config.go)
    db/                      # sqlc-generated code (*.sql.go, models.go, querier.go, db.go)
                             #   migrate.go      embedded startup migrator
                             #   migrations/     NNNN_name.up.sql/.down.sql  (also sqlc's schema)
                             #   queries/        *.sql  (sqlc input)
    auth/                    # password.go (bcrypt), token.go (GenerateToken/HashToken), session.go
    mailer/                  # Mailer interface + ConsoleMailer (logs links; no real SMTP)
    server/                  # server.go (router/DI), response.go (helpers+validator), middleware.go,
                             #   handlers_auth.go / handlers_user.go / handlers_profile.go / handlers_email.go
  Makefile  .golangci.yml  sqlc.yaml  Dockerfile  Dockerfile.dev  .air.toml
frontend/
  src/ main.ts  App.vue  router/index.ts  stores/auth.ts  lib/api.ts  components/  views/
  package.json  pnpm-workspace.yaml  eslint.config.js  .prettierrc.json  vite.config.ts  tsconfig.json
  Dockerfile  Dockerfile.dev  nginx.conf
docker-compose.yml  .env.example  mise.toml  mise.lock  renovate.json
```

## Toolchain (pinned)

- **Go**: version + `toolchain go1.25.11` pinned in `backend/go.mod`. Run Go commands with
  `GOTOOLCHAIN=auto` so the pinned toolchain is auto-downloaded.
- **JS**: Node 24.14.1 + **pnpm 10.34.4** pinned in `mise.toml` (`mise.lock` has hashes).
  Run `mise install` once. Use **pnpm via mise** (`mise exec -- pnpm …`) — never `npm`/`npx`/global installs.

## Common commands

**Run the whole stack locally** (Docker required):
```bash
cp .env.example .env
docker compose up --build      # open http://localhost:5173
```
Services: `postgres`, `redis`, `backend` (air hot-reload :8080), `frontend` (Vite :5173 → proxies /api).
Password-reset / email-change links are **printed to the backend log** (`docker compose logs backend`).

**Backend** (`cd backend`, prefix with `GOTOOLCHAIN=auto`):
```bash
make check     # lint(+gosec) + mod verify/tidy + govulncheck + unit&race tests + build  ← the gate
make test      # unit + race      make lint / make vuln / make build / make fmt
```

**Frontend** (`cd frontend`):
```bash
pnpm install --frozen-lockfile
pnpm dev | build | typecheck | lint | lint:fix | format | format:check
```

## Database & sqlc

- Queries are **type-safe via sqlc**. Edit SQL in `internal/db/queries/*.sql` and the schema in
  `internal/db/migrations/*.up.sql`, then **regenerate** — do **not** hand-edit `internal/db/*.sql.go`.
- Regenerate (local `go run` of sqlc fails to compile its CGO parser on macOS — use the Docker image):
  ```bash
  cd backend && docker run --rm -v "$PWD":/src -w /src sqlc/sqlc:1.27.0 generate
  ```
- **Migrations**: add an `NNNN_name.up.sql` + `.down.sql` pair in `internal/db/migrations`. They are
  embedded and applied on startup by `migrate.go` (golang-migrate-compatible naming; sqlc ignores `.down.sql`).
- Type overrides (see `sqlc.yaml`): `uuid → google/uuid.UUID`, `timestamptz → time.Time` for clean,
  JSON-friendly structs.

## Auth & request conventions

- Sessions: server-side in Redis, cookie `my_hours_session` (HttpOnly; `Secure`+strict `SameSite` when
  `APP_ENV=production`). Helpers in `internal/auth/session.go`: `Login`, `Logout`, `CurrentUserID`.
- `requireAuth()` middleware (server/middleware.go) puts the user id in the gin context (`contextUserIDKey`).
- Handler pattern: `ShouldBindJSON` into a request struct → validate with the `validator` in
  `response.go` → `respondValidation(c, fields)` (422 `{error, errors:{field:msg}}`) or
  `respondError(c, status, msg)` (`{error}`). Never return `db.User` directly (it has `PasswordHash`) —
  map via `toUserDTO`.
- **Anti-enumeration:** password-reset and email-change endpoints always return a generic message and
  never disclose whether an email exists. Preserve this when touching those flows.
- API surface lives under `/api` (see README for the full table): register / login / logout / me /
  password forgot+reset / profile / password change / email change+confirm / health.

## Frontend conventions

- `src/lib/api.ts` is the fetch wrapper: base `/api`, `credentials: 'include'`, throws `ApiError`
  (`status`, `message`, `errors`). The Pinia `auth` store (`stores/auth.ts`) holds the user and all
  auth actions; `router/index.ts` guards routes via `meta.requiresAuth` / `requiresGuest`.
- Reusable components: `AppButton`, `FormInput`, `AppAlert`, `AuthCard`, `AppHeader`, `UserMenu`, `AppLogo`.
- Tailwind v4 (config in `src/style.css` via `@theme`; brand = indigo). ESLint flat config + Prettier
  (single quotes, no semicolons, width 100).

## Gotchas

- **Hot reload uses polling on purpose.** fsnotify/inotify events don't cross macOS/Windows Docker bind
  mounts, so air (`backend/.air.toml` `poll=true`) and Vite (`vite.config.ts` `server.watch.usePolling`)
  both poll. If a newly added file isn't picked up, restart the relevant container — and keep polling on.
- **pnpm supply-chain hardening** (`pnpm-workspace.yaml`): `minimumReleaseAge` (48h cooldown),
  `blockExoticSubdeps`, and `onlyBuiltDependencies: []` (lifecycle scripts blocked). Only add a package
  to that allowlist if it truly needs a build script and you trust it. Don't weaken these.
- **Host ports may be remapped** in `.env` (e.g. `POSTGRES_PORT`, `REDIS_PORT`) to avoid local clashes;
  container ports are fixed and inter-service URLs use service names.
- **Don't weaken security steps** to make checks pass: gosec/golangci-lint, govulncheck, ESLint, the
  pnpm settings. Bump the pinned Go `toolchain` if govulncheck flags stdlib CVEs.

## Before claiming a task is done

- Backend: `cd backend && GOTOOLCHAIN=auto make check` (must exit 0).
- Frontend: `cd frontend && pnpm typecheck && pnpm lint && pnpm format:check && pnpm build`.
- For behavior changes, verify end-to-end against the running stack (curl the API through
  `http://localhost:5173/api`, check `docker compose logs backend`).
- Commits: signed, conventional messages (`feat:`, `fix:`, …).
