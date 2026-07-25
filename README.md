# My hours — Go + Vue time tracker

A personal time tracker: log what you did on a weekly calendar, categorize activities by type,
and see where your hours actually go. Built on a Go API + Vue SPA with server-side sessions.

## Features

- **Weekly calendar** with an hour grid. On phones a view switcher toggles day / 3-day / week /
  month; on desktop it's the full week. Hours run 23:00 → 00:00 so the busy evening leads.
- **Inline activity editing** — click a slot, type, done. Blur or Enter saves; clearing the text
  deletes.
- **Split an hour** into equal slices (down to 10 min) with the + Before / + After actions, or
  **copy** an activity to the right or into the hour above.
- **Activity types** — named, color-coded categories (full CRUD). Picking a type colors the cell;
  text color is chosen automatically for contrast.
- **Time-by-type summary** for the visible range, sorted by most time spent, with percentages.
- **Accounts** — register, login, password reset, and a profile page to change name, email
  (via a confirmation link), and password.

## Stack

| Layer    | Choice                                                              |
|----------|--------------------------------------------------------------------|
| Backend  | Go + [Gin](https://github.com/gin-gonic/gin)                       |
| Auth     | Server-side sessions in Redis, first-party HttpOnly cookie         |
| Database | Postgres via [sqlc](https://sqlc.dev) (typed SQL) + golang-migrate |
| Frontend | Vue 3 + TypeScript + Vite + Tailwind, vue-router + Pinia           |
| Cache    | Redis                                                              |

The browser only ever talks to one origin: the web service (nginx in prod, Vite dev server locally)
serves the SPA **and** reverse-proxies `/api` to the Go service. That keeps session cookies
first-party (no CORS, no third-party-cookie problems) while the Go API stays a clean, standalone
service a future mobile client can consume directly.

## Run locally

Requires Docker + Docker Compose.

```bash
cp .env.example .env
docker compose up --build
```

Then open **http://localhost:5173**.

Services started:

- `postgres` — database (migrations run automatically on backend startup)
- `redis` — session store
- `backend` — Go API with hot reload (air) on :8080
- `frontend` — Vite dev server with HMR on :5173, proxying `/api` → backend

### Try the flow

1. **Register** (name, email, password) → you land on the calendar.
2. Open the top-right menu → **Activity types** and create a few (e.g. Sleep, Work, Gym) with colors.
3. Back on the calendar, **click an hour** and type an activity; pick a type from the popover to
   color it. Split or copy it with the popover actions.
4. Check the **Time by type** summary below the grid.
5. Password reset is available from the login page — the link is **printed to the backend logs**
   (no real email is sent locally): copy it from `docker compose logs backend`.

## Pages

- `/` — landing page
- `/login`, `/register`, `/forgot-password`, `/reset-password`
- `/dashboard` — the weekly calendar (authenticated)
- `/activity-types`, `/activity-types/new`, `/activity-types/:id/edit` — manage types
- `/profile` — change name, email (via confirmation link), and password
- `/confirm-email` — confirms an email change from the link

## API

All endpoints are JSON under `/api`. Activity and activity-type rows are scoped to the current
user — another user's rows return 404, never leaked.

| Method | Path                          | Auth | Purpose                                 |
|--------|-------------------------------|------|-----------------------------------------|
| POST   | `/api/auth/register`          | —    | Create account + start session          |
| POST   | `/api/auth/login`             | —    | Start session                           |
| POST   | `/api/auth/logout`            | ✓    | Destroy session                         |
| GET    | `/api/auth/me`                | ✓    | Current user                            |
| POST   | `/api/auth/password/forgot`   | —    | Issue reset token (logged to console)   |
| POST   | `/api/auth/password/reset`    | —    | Set a new password using a token        |
| POST   | `/api/auth/profile`           | ✓    | Update name                             |
| POST   | `/api/auth/password/change`   | ✓    | Change password (verifies current)      |
| POST   | `/api/auth/email/change`      | ✓    | Request email change (confirm link)     |
| POST   | `/api/auth/email/confirm`     | —    | Confirm email change via token          |
| GET    | `/api/activity-types`         | ✓    | List your activity types                |
| POST   | `/api/activity-types`         | ✓    | Create a type                           |
| GET/PUT/DELETE | `/api/activity-types/:id` | ✓ | Get / update / delete a type            |
| GET    | `/api/activities?from&to`     | ✓    | Activities whose start is in the range  |
| POST   | `/api/activities`             | ✓    | Create an activity                      |
| GET/PUT/DELETE | `/api/activities/:id` | ✓    | Get / update / delete an activity       |
| GET    | `/api/health`                 | —    | Liveness probe                          |

## Project layout

```
backend/   Go API
  cmd/server            entrypoint
  internal/config       env → typed config
  internal/db           sqlc-generated code, migrations, queries
  internal/auth         password hashing, tokens, sessions
  internal/mailer       Mailer interface + console mailer
  internal/server       router, middleware, handlers, DTOs/validation
frontend/  Vue SPA
  src/lib               pure helpers (datetime, slices, summary, contrast, api)
  src/composables       useMediaQuery, useCalendarView, useHourEditor
  src/components         shared UI + components/calendar
  src/views             pages
  src/stores            Pinia stores (auth, activities, activityTypes)
docker-compose.yml       local dev stack
```

## Development notes

### Regenerating sqlc code

SQL lives in `backend/internal/db/migrations` (schema) and `backend/internal/db/queries`.
After changing them, regenerate the typed Go (`sqlc`'s CGO parser doesn't build locally on macOS,
so use the official image):

```bash
cd backend
docker run --rm -v "$PWD":/src -w /src sqlc/sqlc:1.27.0 generate
```

Do not hand-edit the generated `internal/db/*.sql.go`.

### Adding a migration

Create `NNNN_name.up.sql` / `NNNN_name.down.sql` pairs in `backend/internal/db/migrations`.
They are embedded into the binary and applied on startup.

### Tests

```bash
cd backend  && go test ./...   # handler, validation, auth unit tests
cd frontend && pnpm test       # Vitest unit tests for the pure logic (datetime, slices, summary…)
```

## Security & code quality

This project follows a set of supply-chain and code-quality practices
— see [SECURITY.md](SECURITY.md), [CONTRIBUTING.md](CONTRIBUTING.md), and [AGENTS.md](AGENTS.md).

One-time local setup:

```bash
mise install   # pinned Node + pnpm (hash-verified via mise.lock)

# Harden your Go environment
go env -w GOPROXY="https://proxy.golang.org,direct"
go env -w GOSUMDB="sum.golang.org"
```

Run the checks:

```bash
cd backend  && make check                                   # lint+gosec, mod verify/tidy, govulncheck, tests, build
cd frontend && pnpm install --frozen-lockfile && pnpm typecheck && pnpm lint && pnpm format:check && pnpm test && pnpm build
```

What's enforced:

- **Go:** `golangci-lint` (errcheck, govet, staticcheck, unused, **gosec**) + `govulncheck`;
  pinned Go/toolchain; hardened, reproducible distroless build.
- **Frontend:** **pnpm** with `minimumReleaseAge` + `blockExoticSubdeps` + lifecycle scripts
  blocked by default; ESLint + strict TypeScript; `mise`-pinned toolchain.
- **Repo:** Renovate dependency updates, signed/conventional commits, `CODEOWNERS`.

## Deploying to Railway

Deploy two services from this repo, plus the managed add-ons:

- **API service** — build `backend/Dockerfile`. Set `DATABASE_URL`, `REDIS_ADDR`, `SESSION_SECRET`,
  `APP_ENV=production`, `APP_URL` (your web app's public URL).
- **Web service** — build `frontend/Dockerfile` (nginx). Set `API_UPSTREAM` to the API service's
  private hostname (e.g. `api.railway.internal:8080`) and `PORT` to the port Railway assigns. nginx
  serves the SPA and reverse-proxies `/api` to that upstream.
- **Postgres** and **Redis** — Railway add-ons; wire their connection strings into the API service.

In production set `APP_ENV=production` so the session cookie is sent with `Secure` and a strict
`SameSite`. A future mobile client can call the API service's public URL directly (token auth would
be added then).
