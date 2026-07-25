# Security

## Reporting a vulnerability

Please report suspected vulnerabilities privately to the maintainers (do not open a public
issue). Include steps to reproduce and impact. We aim to acknowledge within a few business days.

## Practices baked into this project

**Backend (Go)**
- `gosec` (via golangci-lint) and `govulncheck` call-graph vulnerability scanning.
- Pinned Go version + toolchain (`backend/go.mod`); `go.mod`/`go.sum` committed and verified.
- Hardened, reproducible builds (`CGO_ENABLED=0 -trimpath -ldflags="-s -w"`); distroless,
  non-root runtime image with pinned base images.
- Passwords hashed with bcrypt; sessions are server-side (Redis) with HttpOnly cookies
  (`Secure` + strict `SameSite` in production).
- Anti-enumeration: password-reset (and email-change) responses are generic and never
  disclose whether an address is registered.

**Frontend (TypeScript)**
- pnpm with supply-chain hardening: `minimumReleaseAge` cool-down, `blockExoticSubdeps`,
  and lifecycle scripts blocked by default (`onlyBuiltDependencies`).
- Toolchain pinned with hash verification via `mise` (`mise.lock`).
- ESLint + strict TypeScript; lockfile committed and installed with `--frozen-lockfile`.

**Repo-wide**
- Renovate for scheduled dependency updates (security alerts labeled `security`).
- Signed, conventional commits.

## Incident response (suspected supply-chain compromise)

- Go: `go clean -modcache && go clean -cache`, review `go.sum` history, pin/fork via a
  `replace` directive if needed.
- JS: `rm -rf $(pnpm store path)` and `rm -rf node_modules`, reinstall with `--frozen-lockfile`.
- Rotate any CI/CD secrets and cloud credentials that may have been exposed.
