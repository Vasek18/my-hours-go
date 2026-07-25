# Contributing

## One-time local setup

Install the pinned JS toolchain (Node + pnpm) with hash verification:

```bash
mise install
```

Harden your Go environment (enforces checksum verification):

```bash
go env -w GOPROXY="https://proxy.golang.org,direct"
go env -w GOSUMDB="sum.golang.org"
```

## Pre-PR checklist

**Backend** (`backend/`):

```bash
make check
```

Runs `golangci-lint` (errcheck, govet, staticcheck, unused, **gosec** + gofmt/goimports),
`go mod verify` + tidy hygiene, unit and race tests, `govulncheck`, and a hardened build.
Run `govulncheck` before every push.

**Frontend** (`frontend/`):

```bash
pnpm install --frozen-lockfile
pnpm typecheck && pnpm lint && pnpm format:check && pnpm build
```

## Conventions

- **Signed commits** are required.
- Use **conventional commit** messages (`feat:`, `fix:`, `chore:`, `docs:`, `test:`).
- Use `pnpm exec <tool>` in the frontend — never `npx`, `pnpm dlx`, or global installs.
- Don't add dependencies casually — prefer the standard library (Go) and built-ins (JS).
  Before adding one, check its transitive footprint and maintenance health.
- Never weaken or disable the security/lint steps to make CI pass.
