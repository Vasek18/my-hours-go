# Helm charts for ArgoCD deployment — design

Date: 2026-07-25

## Goal

Deploy the two standalone services in this repo (Go API, Vue SPA behind nginx) to a Kubernetes
cluster managed by ArgoCD. This task produces the Helm chart(s) only — no ArgoCD `Application`
manifests, no Ingress, no in-cluster Postgres/Redis. Those are handled outside this repo/chart.

## Scope decisions

- **Postgres/Redis**: external only. The chart never deploys a database or cache; it only wires
  connection info into the backend via values/secrets.
- **Chart structure**: one umbrella chart (`charts/my-hours`) with two subcharts (`backend`,
  `frontend`), mirroring the two independently-deployable services described in CLAUDE.md while
  still allowing a single `helm install`/ArgoCD sync of both.
- **ArgoCD Application manifests**: out of scope. The user wires the chart into ArgoCD separately.
- **Ingress/TLS**: out of scope. The chart only creates `Service` resources (ClusterIP).
- **Secrets**: the chart never templates a `Secret`. Sensitive env vars are sourced from an
  `existingSecret` (name supplied via values) that the user creates out-of-band (kubectl,
  Sealed Secrets, External Secrets Operator, etc). Nothing sensitive is templated or committed.
- **Images**: published to GHCR — `ghcr.io/vasek18/my-hours-backend` and
  `ghcr.io/vasek18/my-hours-frontend` — tag supplied via `image.tag` in values.

## Layout

```
charts/
  my-hours/                    # umbrella chart
    Chart.yaml                 # apiVersion v2, type: application
    values.yaml                # top-level, namespaced under `backend:` / `frontend:`
    charts/
      backend/
        Chart.yaml
        values.yaml
        templates/
          deployment.yaml
          service.yaml
          _helpers.tpl
          NOTES.txt
      frontend/
        Chart.yaml
        values.yaml
        templates/
          deployment.yaml
          service.yaml
          _helpers.tpl
          NOTES.txt
```

## Backend subchart

- **Deployment**: image `ghcr.io/vasek18/my-hours-backend:{{ .Values.image.tag }}` (default tag
  from `Chart.yaml` `appVersion` if unset), container port `8080`.
- **Env vars**:
  - Plain, from values: `APP_ENV`, `APP_URL`, `REDIS_ADDR`.
  - From `existingSecret` (name via `Values.existingSecret`, required when secret-backed env vars
    are used): `DATABASE_URL`, `SESSION_SECRET`. Secret key names configurable via values
    (`secretKeys.databaseUrl` / `secretKeys.sessionSecret`), defaulting to `DATABASE_URL` /
    `SESSION_SECRET`.
- **Probes**: readiness and liveness both `GET /api/health` on the container port.
- **securityContext**: `runAsNonRoot: true`, `runAsUser: 65532`, `readOnlyRootFilesystem: true`,
  `allowPrivilegeEscalation: false`, drop `ALL` capabilities — matches the distroless nonroot
  runtime image (`backend/Dockerfile`).
- **Service**: ClusterIP, port 8080 → container port 8080.
- **Configurable**: `replicaCount` (default 1), `resources` (small sane defaults, e.g.
  `requests: {cpu: 50m, memory: 64Mi}`, `limits: {cpu: 250m, memory: 128Mi}`).

## Frontend subchart

- **Deployment**: image `ghcr.io/vasek18/my-hours-frontend:{{ .Values.image.tag }}`, container
  port `80`.
- **Env vars**:
  - `PORT`: fixed at `80` (container always listens here; Service maps to it).
  - `API_UPSTREAM`: defaults to `{{ .Release.Name }}-backend:8080` (in-cluster DNS name of the
    backend subchart's Service), overridable via values for non-standard release names/namespaces.
- **Probes**: readiness and liveness both `GET /` on the container port.
- **securityContext**: correction from the originally approved design — `runAsNonRoot: true` is
  **not** used here. The stock `nginx:1.27-alpine` image's master process binds the privileged
  port `80` directly (no setuid step of its own under Kubernetes' `runAsUser`), which requires
  either root or the `CAP_NET_BIND_SERVICE` capability; forcing non-root without that capability
  would crash-loop the pod. So the frontend container runs as the image's default (root) user,
  keeps capabilities untouched (no `drop: ["ALL"]`), and only sets `allowPrivilegeEscalation:
  false` and `readOnlyRootFilesystem: false` (nginx's entrypoint renders `${PORT}`/`${API_UPSTREAM}`
  via envsubst into `/etc/nginx/conf.d/default.conf` and needs to write PID/cache files at
  startup). The backend keeps its full non-root + read-only + drop-all hardening since its
  distroless image was built for exactly that.
- **Service**: ClusterIP, port 80 → container port 80.
- **Configurable**: `replicaCount` (default 1), `resources` (small sane defaults).

## Values structure (top-level `charts/my-hours/values.yaml`)

```yaml
backend:
  image:
    repository: ghcr.io/vasek18/my-hours-backend
    tag: ""             # falls back to subchart's appVersion
  replicaCount: 1
  env:
    appEnv: production
    appUrl: https://your-domain.example
    redisAddr: redis:6379
  existingSecret: my-hours-backend-secret
  secretKeys:
    databaseUrl: DATABASE_URL
    sessionSecret: SESSION_SECRET
  resources:
    requests: { cpu: 50m, memory: 64Mi }
    limits: { cpu: 250m, memory: 128Mi }

frontend:
  image:
    repository: ghcr.io/vasek18/my-hours-frontend
    tag: ""
  replicaCount: 1
  apiUpstream: ""       # empty -> template default "<release>-backend:8080"
  resources:
    requests: { cpu: 25m, memory: 32Mi }
    limits: { cpu: 150m, memory: 64Mi }
```

Each subchart also carries its own standalone `values.yaml` with the same defaults, so either
subchart can be `helm template`d/installed independently for testing.

## Testing / verification

- `helm lint charts/my-hours` (and each subchart individually).
- `helm template charts/my-hours` with default values — sanity-check rendered Deployment/Service
  manifests (image refs, env vars, probes, securityContext).
- `helm template charts/my-hours --set backend.existingSecret=test-secret` to confirm secret
  wiring renders correctly.
- `helm install --dry-run --debug` against a real or kind cluster if available.

## Out of scope (explicitly, per user answers)

- Postgres/Redis manifests or subchart dependencies.
- Ingress/TLS resources or cert-manager annotations.
- ArgoCD `Application`/`ApplicationSet` manifests.
- In-chart `Secret` templating.
- HorizontalPodAutoscaler, PodDisruptionBudget, NetworkPolicy — not requested; can be added later
  if needed.

## Related cleanup (done alongside this task)

Removed stray "Railway" brand mentions from `README.md`, `backend/internal/config/config.go`, and
`frontend/nginx.conf` (comments/docs only — no behavior change), since this project is not being
deployed there.
