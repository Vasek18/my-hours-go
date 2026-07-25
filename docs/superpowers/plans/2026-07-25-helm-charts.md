# Helm Charts for ArgoCD Deployment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a Helm umbrella chart (`charts/my-hours`) with two subcharts (`backend`, `frontend`) that deploy this repo's Go API and Vue SPA to Kubernetes, ready for ArgoCD to sync.

**Architecture:** Two independent, standalone Helm subcharts (`charts/my-hours/charts/backend`, `charts/my-hours/charts/frontend`) — each with its own `Chart.yaml`/`values.yaml`/templates so either can be `helm template`d or installed on its own — wrapped by a thin umbrella chart (`charts/my-hours`) whose top-level `values.yaml` namespaces overrides under `backend:`/`frontend:` keys. Helm auto-discovers subcharts placed under `charts/<name>/`, so no `dependencies:` block is needed in the umbrella `Chart.yaml`.

**Tech Stack:** Helm 3 (installed at `~/.local/bin/helm`, v3.21.3). No cluster is available in this environment, so verification uses `helm lint` and `helm template` only (no `--dry-run` against a live API server).

## Global Constraints

- Images: `ghcr.io/vasek18/my-hours-backend` and `ghcr.io/vasek18/my-hours-frontend`, tag via `image.tag` in values (empty defaults to the chart's `appVersion`).
- No Postgres/Redis manifests, no Ingress/TLS, no ArgoCD `Application` manifests, no in-chart `Secret` templating — all out of scope per the approved design.
- Backend sensitive env vars (`DATABASE_URL`, `SESSION_SECRET`) come only from an `existingSecret` (name via values); the chart must `fail` a clear error if `existingSecret` is unset, rather than deploying pods that boot with empty secrets.
- Backend container: port `8080`, health probe `GET /api/health`, `securityContext` fully hardened (`runAsNonRoot: true`, `runAsUser: 65532`, `readOnlyRootFilesystem: true`, `allowPrivilegeEscalation: false`, drop `ALL` capabilities) — matches the distroless nonroot image in `backend/Dockerfile`.
- Frontend container: port `80`, health probe `GET /`, runs as the image's default root user (stock `nginx:1.27-alpine` needs to bind privileged port 80 — see corrected design note in `docs/superpowers/specs/2026-07-25-helm-charts-design.md`), `readOnlyRootFilesystem: false`, `allowPrivilegeEscalation: false`, capabilities untouched.
- Frontend `API_UPSTREAM` defaults to `<release-name>-backend:8080` (the backend subchart Service's in-cluster DNS name under the default fullname template) when not overridden in values.
- All resource names use a `fullname` helper following the standard `helm create` convention (`<release>-<chart>` unless the chart name is already contained in the release name).

---

### Task 1: Backend subchart — chart metadata, values, and helpers

**Files:**
- Create: `charts/my-hours/charts/backend/Chart.yaml`
- Create: `charts/my-hours/charts/backend/values.yaml`
- Create: `charts/my-hours/charts/backend/templates/_helpers.tpl`

**Interfaces:**
- Produces: template helpers `backend.name`, `backend.fullname`, `backend.labels`, `backend.selectorLabels` — consumed by Task 2's `deployment.yaml`/`service.yaml`/`NOTES.txt`.
- Produces: values keys `replicaCount`, `image.repository`, `image.tag`, `image.pullPolicy`, `env.appEnv`, `env.appUrl`, `env.redisAddr`, `existingSecret`, `secretKeys.databaseUrl`, `secretKeys.sessionSecret`, `service.port`, `resources` — consumed by Task 2.

- [ ] **Step 1: Create the backend subchart directory and `Chart.yaml`**

```bash
mkdir -p charts/my-hours/charts/backend/templates
```

`charts/my-hours/charts/backend/Chart.yaml`:

```yaml
apiVersion: v2
name: backend
description: My Hours API (Go + Gin) Helm chart
type: application
version: 0.1.0
appVersion: "0.1.0"
```

- [ ] **Step 2: Create `charts/my-hours/charts/backend/values.yaml`**

```yaml
replicaCount: 1

image:
  repository: ghcr.io/vasek18/my-hours-backend
  tag: ""
  pullPolicy: IfNotPresent

env:
  appEnv: production
  appUrl: https://your-domain.example
  redisAddr: redis:6379

# Name of a pre-existing Secret (created outside this chart) with the keys
# named below. Required — the chart fails to render without it.
existingSecret: ""
secretKeys:
  databaseUrl: DATABASE_URL
  sessionSecret: SESSION_SECRET

service:
  port: 8080

resources:
  requests:
    cpu: 50m
    memory: 64Mi
  limits:
    cpu: 250m
    memory: 128Mi
```

- [ ] **Step 3: Create `charts/my-hours/charts/backend/templates/_helpers.tpl`**

```
{{- define "backend.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "backend.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "backend.labels" -}}
app.kubernetes.io/name: {{ include "backend.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
{{- end -}}

{{- define "backend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "backend.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
```

- [ ] **Step 4: Verify the partial chart lints clean (templates are empty so far, so `lint` will warn about no templates — that's expected at this step)**

Run: `helm lint charts/my-hours/charts/backend`
Expected: `1 chart(s) linted, 0 chart(s) failed` (an `[INFO] Chart.yaml: icon is recommended` note is fine; a warning about missing templates is expected and resolved in Task 2)

- [ ] **Step 5: Commit**

```bash
git add charts/my-hours/charts/backend/Chart.yaml charts/my-hours/charts/backend/values.yaml charts/my-hours/charts/backend/templates/_helpers.tpl
git commit -m "feat: scaffold backend Helm subchart metadata and helpers"
```

---

### Task 2: Backend subchart — Deployment, Service, NOTES

**Files:**
- Create: `charts/my-hours/charts/backend/templates/deployment.yaml`
- Create: `charts/my-hours/charts/backend/templates/service.yaml`
- Create: `charts/my-hours/charts/backend/templates/NOTES.txt`

**Interfaces:**
- Consumes: `backend.fullname`, `backend.labels`, `backend.selectorLabels` from Task 1's `_helpers.tpl`; all values keys defined in Task 1's `values.yaml`.
- Produces: a Deployment named via `backend.fullname` exposing container port `8080` (name `http`), and a Service of the same name exposing port `.Values.service.port` → target port `http`. Task 4 (umbrella chart) and Task 3 (frontend `API_UPSTREAM` default) rely on this Service's DNS name being `<release>-backend` when the umbrella chart's backend subchart is named `backend` (matches `backend.fullname`'s default `<release>-<chart-name>` pattern).

- [ ] **Step 1: Create `charts/my-hours/charts/backend/templates/deployment.yaml`**

```yaml
{{- if not .Values.existingSecret }}
{{- fail "backend.existingSecret must be set to the name of a Secret containing DATABASE_URL and SESSION_SECRET keys" }}
{{- end }}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "backend.fullname" . }}
  labels:
    {{- include "backend.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      {{- include "backend.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "backend.selectorLabels" . | nindent 8 }}
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 65532
        runAsGroup: 65532
        fsGroup: 65532
      containers:
        - name: backend
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - name: http
              containerPort: 8080
              protocol: TCP
          env:
            - name: APP_ENV
              value: {{ .Values.env.appEnv | quote }}
            - name: APP_URL
              value: {{ .Values.env.appUrl | quote }}
            - name: REDIS_ADDR
              value: {{ .Values.env.redisAddr | quote }}
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: {{ .Values.existingSecret }}
                  key: {{ .Values.secretKeys.databaseUrl }}
            - name: SESSION_SECRET
              valueFrom:
                secretKeyRef:
                  name: {{ .Values.existingSecret }}
                  key: {{ .Values.secretKeys.sessionSecret }}
          readinessProbe:
            httpGet:
              path: /api/health
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /api/health
              port: http
            initialDelaySeconds: 10
            periodSeconds: 20
          securityContext:
            readOnlyRootFilesystem: true
            allowPrivilegeEscalation: false
            capabilities:
              drop: ["ALL"]
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
```

- [ ] **Step 2: Create `charts/my-hours/charts/backend/templates/service.yaml`**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: {{ include "backend.fullname" . }}
  labels:
    {{- include "backend.labels" . | nindent 4 }}
spec:
  type: ClusterIP
  ports:
    - port: {{ .Values.service.port }}
      targetPort: http
      protocol: TCP
      name: http
  selector:
    {{- include "backend.selectorLabels" . | nindent 4 }}
```

- [ ] **Step 3: Create `charts/my-hours/charts/backend/templates/NOTES.txt`**

```
Backend "{{ include "backend.fullname" . }}" deployed.
In-cluster address: {{ include "backend.fullname" . }}:{{ .Values.service.port }}
```

- [ ] **Step 4: Verify the chart fails to render without `existingSecret` (required-value guard)**

Run: `helm template charts/my-hours/charts/backend`
Expected: fails with `execution error at (backend/templates/deployment.yaml:2:4): backend.existingSecret must be set to the name of a Secret containing DATABASE_URL and SESSION_SECRET keys`

- [ ] **Step 5: Verify the chart renders correctly once `existingSecret` is set**

Run: `helm template charts/my-hours/charts/backend --set existingSecret=my-hours-backend-secret`
Expected: valid YAML output containing one `Deployment` and one `Service`, both named `release-name-backend` (Helm's default release name in `helm template` without `--release-name` is `release-name`), the Deployment's container image `ghcr.io/vasek18/my-hours-backend:0.1.0`, container port `8080`, `DATABASE_URL`/`SESSION_SECRET` env vars sourced via `secretKeyRef` from `my-hours-backend-secret`, and probes hitting `/api/health`

- [ ] **Step 6: Run `helm lint` clean**

Run: `helm lint charts/my-hours/charts/backend --set existingSecret=my-hours-backend-secret`
Expected: `1 chart(s) linted, 0 chart(s) failed`

- [ ] **Step 7: Commit**

```bash
git add charts/my-hours/charts/backend/templates/deployment.yaml charts/my-hours/charts/backend/templates/service.yaml charts/my-hours/charts/backend/templates/NOTES.txt
git commit -m "feat: add backend Deployment and Service templates"
```

---

### Task 3: Frontend subchart — full chart (metadata, values, helpers, Deployment, Service, NOTES)

**Files:**
- Create: `charts/my-hours/charts/frontend/Chart.yaml`
- Create: `charts/my-hours/charts/frontend/values.yaml`
- Create: `charts/my-hours/charts/frontend/templates/_helpers.tpl`
- Create: `charts/my-hours/charts/frontend/templates/deployment.yaml`
- Create: `charts/my-hours/charts/frontend/templates/service.yaml`
- Create: `charts/my-hours/charts/frontend/templates/NOTES.txt`

**Interfaces:**
- Consumes: nothing from other tasks — the frontend subchart is fully self-contained. Its `API_UPSTREAM` default (`<release-name>-backend:8080`) assumes the backend subchart is deployed under the same release with its default `backend.fullname` (Task 2), but does not reference the backend chart's templates directly (they're separate charts).
- Produces: values keys `replicaCount`, `image.repository`, `image.tag`, `image.pullPolicy`, `apiUpstream`, `service.port`, `resources` — consumed by Task 4 (umbrella `values.yaml`).

- [ ] **Step 1: Create the frontend subchart directory and `Chart.yaml`**

```bash
mkdir -p charts/my-hours/charts/frontend/templates
```

`charts/my-hours/charts/frontend/Chart.yaml`:

```yaml
apiVersion: v2
name: frontend
description: My Hours SPA (Vue 3 + nginx) Helm chart
type: application
version: 0.1.0
appVersion: "0.1.0"
```

- [ ] **Step 2: Create `charts/my-hours/charts/frontend/values.yaml`**

```yaml
replicaCount: 1

image:
  repository: ghcr.io/vasek18/my-hours-frontend
  tag: ""
  pullPolicy: IfNotPresent

# In-cluster host:port of the backend API. Empty defaults to
# "<release-name>-backend:8080" (the backend subchart's default Service name).
apiUpstream: ""

service:
  port: 80

resources:
  requests:
    cpu: 25m
    memory: 32Mi
  limits:
    cpu: 150m
    memory: 64Mi
```

- [ ] **Step 3: Create `charts/my-hours/charts/frontend/templates/_helpers.tpl`**

```
{{- define "frontend.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "frontend.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "frontend.labels" -}}
app.kubernetes.io/name: {{ include "frontend.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
{{- end -}}

{{- define "frontend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "frontend.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
```

- [ ] **Step 4: Create `charts/my-hours/charts/frontend/templates/deployment.yaml`**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "frontend.fullname" . }}
  labels:
    {{- include "frontend.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      {{- include "frontend.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "frontend.selectorLabels" . | nindent 8 }}
    spec:
      containers:
        - name: frontend
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - name: http
              containerPort: 80
              protocol: TCP
          env:
            - name: PORT
              value: "80"
            - name: API_UPSTREAM
              value: {{ .Values.apiUpstream | default (printf "%s-backend:8080" .Release.Name) | quote }}
          readinessProbe:
            httpGet:
              path: /
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /
              port: http
            initialDelaySeconds: 10
            periodSeconds: 20
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: false
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
```

Note: no pod- or container-level `runAsNonRoot`/`runAsUser` here — the stock `nginx:1.27-alpine` image's master process binds port 80 directly and needs root (or `CAP_NET_BIND_SERVICE`, which isn't worth the added template complexity for a single low-risk static-file+proxy container). See the corrected design note in `docs/superpowers/specs/2026-07-25-helm-charts-design.md`.

- [ ] **Step 5: Create `charts/my-hours/charts/frontend/templates/service.yaml`**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: {{ include "frontend.fullname" . }}
  labels:
    {{- include "frontend.labels" . | nindent 4 }}
spec:
  type: ClusterIP
  ports:
    - port: {{ .Values.service.port }}
      targetPort: http
      protocol: TCP
      name: http
  selector:
    {{- include "frontend.selectorLabels" . | nindent 4 }}
```

- [ ] **Step 6: Create `charts/my-hours/charts/frontend/templates/NOTES.txt`**

```
Frontend "{{ include "frontend.fullname" . }}" deployed.
In-cluster address: {{ include "frontend.fullname" . }}:{{ .Values.service.port }}
```

- [ ] **Step 7: Verify default `API_UPSTREAM` rendering**

Run: `helm template charts/my-hours/charts/frontend`
Expected: valid YAML with one `Deployment` and one `Service` named `release-name-frontend`, container image `ghcr.io/vasek18/my-hours-frontend:0.1.0`, container port `80`, env var `API_UPSTREAM` equal to `release-name-backend:8080`, and `PORT` equal to `"80"`

- [ ] **Step 8: Verify `apiUpstream` override works**

Run: `helm template charts/my-hours/charts/frontend --set apiUpstream=custom-api:9000`
Expected: rendered `API_UPSTREAM` env var equals `custom-api:9000`

- [ ] **Step 9: Run `helm lint` clean**

Run: `helm lint charts/my-hours/charts/frontend`
Expected: `1 chart(s) linted, 0 chart(s) failed`

- [ ] **Step 10: Commit**

```bash
git add charts/my-hours/charts/frontend
git commit -m "feat: add frontend Helm subchart"
```

---

### Task 4: Umbrella chart wiring both subcharts

**Files:**
- Create: `charts/my-hours/Chart.yaml`
- Create: `charts/my-hours/values.yaml`

**Interfaces:**
- Consumes: `charts/my-hours/charts/backend` (Task 1+2) and `charts/my-hours/charts/frontend` (Task 3) as filesystem subcharts — Helm auto-discovers any chart directory under `charts/`, so no `dependencies:` entry is required in `Chart.yaml` for local, non-repository subcharts.
- Produces: a single `helm template`/`helm install` entry point at `charts/my-hours` that renders both subcharts' resources together, with per-subchart overrides nested under top-level `backend:`/`frontend:` keys (Helm's convention: a top-level values key matching a subchart's directory name is passed down as that subchart's own `.Values`).

- [ ] **Step 1: Create `charts/my-hours/Chart.yaml`**

```yaml
apiVersion: v2
name: my-hours
description: My Hours — Go API + Vue SPA umbrella Helm chart
type: application
version: 0.1.0
appVersion: "0.1.0"
```

- [ ] **Step 2: Create `charts/my-hours/values.yaml`**

```yaml
backend:
  image:
    repository: ghcr.io/vasek18/my-hours-backend
    tag: ""
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
    requests:
      cpu: 50m
      memory: 64Mi
    limits:
      cpu: 250m
      memory: 128Mi

frontend:
  image:
    repository: ghcr.io/vasek18/my-hours-frontend
    tag: ""
  replicaCount: 1
  apiUpstream: ""
  resources:
    requests:
      cpu: 25m
      memory: 32Mi
    limits:
      cpu: 150m
      memory: 64Mi
```

- [ ] **Step 3: Verify the umbrella chart renders both subcharts together**

Run: `helm template charts/my-hours`
Expected: valid YAML containing exactly two `Deployment` resources (`release-name-backend`, `release-name-frontend`) and two `Service` resources with the same names; the frontend's `API_UPSTREAM` env var equals `release-name-backend:8080`; the backend's `DATABASE_URL`/`SESSION_SECRET` env vars reference secret `my-hours-backend-secret`

- [ ] **Step 4: Verify overriding a nested value works end-to-end through the umbrella chart**

Run: `helm template charts/my-hours --set backend.existingSecret=other-secret --set frontend.apiUpstream=other-backend:8080`
Expected: backend's `secretKeyRef.name` is `other-secret`; frontend's `API_UPSTREAM` is `other-backend:8080`

- [ ] **Step 5: Run `helm lint` on the umbrella chart**

Run: `helm lint charts/my-hours`
Expected: `1 chart(s) linted, 0 chart(s) failed` (dependent subcharts are linted as part of the parent)

- [ ] **Step 6: Commit**

```bash
git add charts/my-hours/Chart.yaml charts/my-hours/values.yaml
git commit -m "feat: add my-hours umbrella Helm chart"
```

---

### Task 5: Full verification pass and README pointer

**Files:**
- Modify: `README.md` (the "Deploying" section already added by the design/cleanup work points to `charts/` — verify it's accurate, no further edit expected unless verification surfaces a gap)

**Interfaces:**
- Consumes: the complete `charts/my-hours` tree from Tasks 1–4.
- Produces: nothing new — this task is a final gate before considering the feature done.

- [ ] **Step 1: Lint every chart individually and the umbrella chart**

Run:
```bash
helm lint charts/my-hours/charts/backend --set existingSecret=my-hours-backend-secret
helm lint charts/my-hours/charts/frontend
helm lint charts/my-hours
```
Expected: all three report `0 chart(s) failed`

- [ ] **Step 2: Render the umbrella chart with production-like overrides and inspect the full output**

Run:
```bash
helm template my-hours charts/my-hours \
  --set backend.env.appUrl=https://myhours.example.com \
  --set backend.existingSecret=my-hours-backend-secret \
  --set backend.image.tag=v0.1.0 \
  --set frontend.image.tag=v0.1.0
```
Expected: two Deployments (`my-hours-backend`, `my-hours-frontend`) and two Services, images tagged `v0.1.0`, `APP_URL` set to `https://myhours.example.com`, no template errors

- [ ] **Step 3: Confirm `README.md`'s "Deploying" section references `charts/` accurately**

Read `README.md` around the "Deploying" heading (added when Railway mentions were removed earlier in this task) and confirm it says: "For Kubernetes via ArgoCD, see the Helm charts under `charts/`." No code change expected — this step is a check, not an edit, unless the sentence is missing or inaccurate, in which case add/fix it as a one-line edit.

- [ ] **Step 4: Final commit if Step 3 required a README edit**

```bash
git add README.md
git commit -m "docs: point README at charts/ for Kubernetes deployment"
```

(Skip this commit if Step 3 required no changes.)
