# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.26.7
ARG NODE_VERSION=22-alpine

FROM golang:${GO_VERSION}-alpine AS backend-build
WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY backend/ ./
ARG TARGETOS=linux
ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/zeorcode .

FROM node:${NODE_VERSION} AS frontend-build
WORKDIR /src/frontend

COPY frontend/package.json frontend/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci

COPY frontend/ ./
RUN npm run build

FROM node:${NODE_VERSION} AS admin-build
ENV PNPM_HOME=/pnpm
ENV PATH=${PNPM_HOME}:${PATH}
WORKDIR /src/admin

RUN corepack enable
COPY admin-frontend/package.json admin-frontend/pnpm-lock.yaml admin-frontend/.npmrc ./
RUN --mount=type=cache,id=pnpm-store,target=/pnpm/store \
    pnpm config set store-dir /pnpm/store && \
    pnpm install --frozen-lockfile

COPY admin-frontend/ ./
RUN pnpm build

FROM alpine:3.22 AS backend
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S -g 10001 zeorcode && \
    adduser -S -D -H -u 10001 -G zeorcode zeorcode

WORKDIR /app
COPY --from=backend-build --chown=10001:10001 /out/zeorcode /usr/local/bin/zeorcode
COPY --chown=10001:10001 ops/docker/config.yaml /app/config.yaml
RUN mkdir -p /app/uploads /app/ojdata/problems /app/runtime && \
    chown -R 10001:10001 /app

USER 10001:10001
EXPOSE 9090
ENTRYPOINT ["zeorcode"]
CMD ["http"]

FROM nginx:1.28-alpine AS web
COPY ops/docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=frontend-build /src/frontend/dist/ /usr/share/nginx/html/
COPY --from=admin-build /src/admin/dist/ /usr/share/nginx/html/admin/
EXPOSE 80
