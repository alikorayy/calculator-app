# syntax=docker/dockerfile:1

# Build stage: type-check and bundle the app with the Node version the project supports.
FROM node:22.12.0-alpine AS build
WORKDIR /app

# Install dependencies in their own layer, so they are cached until the lockfile changes.
COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

# Runtime stage: only nginx and the static files. This nginx image runs as an
# unprivileged user and therefore listens on 8080 instead of 80.
FROM nginxinc/nginx-unprivileged:1.28-alpine

# Where /api requests go. The image fills it into the template at startup.
ENV API_UPSTREAM=http://api:8080

COPY nginx/security-headers.conf /etc/nginx/snippets/security-headers.conf
COPY nginx/default.conf.template /etc/nginx/templates/default.conf.template
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --spider http://127.0.0.1:8080/ || exit 1
