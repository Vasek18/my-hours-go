# Production image: build the SPA, serve it with nginx, proxy /api to the API.
FROM node:24.20.0-alpine AS build

WORKDIR /app

# Activate the pinned pnpm via corepack — no global installs.
RUN corepack enable && corepack prepare pnpm@10.34.4 --activate

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .
RUN pnpm build

FROM nginx:1.31-alpine

# nginx processes templates in /etc/nginx/templates with envsubst at startup,
# substituting only the env vars below (its own $variables are left untouched).
ENV PORT=80
ENV API_UPSTREAM=backend:8080

COPY nginx.conf /etc/nginx/templates/default.conf.template
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
