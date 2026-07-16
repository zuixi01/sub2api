# Sub2API

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.26%2B-00ADD8.svg)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3-4FC08D.svg)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-supported-336791.svg)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-supported-DC382D.svg)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED.svg)](https://www.docker.com/)

**A self-hosted AI API gateway for managing accounts, quotas, keys, and access.**

[English](README.md) · [中文](README_CN.md) · [日本語](README_JA.md)

</div>

## Overview

Sub2API turns configured upstream AI accounts into a managed API service. It
provides a web dashboard for administrators and users, issues API keys, tracks
usage, enforces quotas and limits, and routes requests through account groups.

The gateway exposes compatible request paths for Anthropic Messages, OpenAI
Chat Completions and Responses, and Gemini `v1beta` clients. Configure only
accounts and integrations you are authorized to use.

## Important notice

- You are responsible for complying with the terms of every upstream provider
  and with all laws that apply to your deployment.
- Do not use the project to bypass provider restrictions, access controls, or
  billing requirements.
- This software is provided under the [LGPL-3.0-or-later](LICENSE), without
  warranty. Operate it only after reviewing your security, privacy, and
  compliance requirements.

## Highlights

- **Account and group management** — Organize OAuth and API-key based upstream
  accounts into routable groups.
- **API keys and access control** — Issue user keys, assign groups, and apply
  per-user or per-account concurrency and rate limits.
- **Usage and quota controls** — Track usage, manage balances and
  subscriptions, and apply token-aware billing rules.
- **Operations dashboard** — Manage users, accounts, groups, routing, and
  system settings from the Vue-based web interface.
- **Deployment choices** — Run with Docker Compose, an Apple `container`
  workflow, or a self-built binary. PostgreSQL and Redis are the standard
  runtime dependencies.

## Quick start with Docker Compose

The following Linux/macOS workflow creates a deployment directory, downloads
the repository's Docker preparation script, generates required secrets, and
starts the stack.

```bash
mkdir sub2api-deploy && cd sub2api-deploy
curl -sSL https://raw.githubusercontent.com/zuixi01/sub2api/main/deploy/docker-deploy.sh | bash
docker compose up -d
docker compose logs -f sub2api
```

Open `http://YOUR_SERVER_IP:8080` after the service is healthy. The script
creates a local `.env` file and data directories. If you did not set
`ADMIN_PASSWORD`, find the one-time generated administrator password in the
first-start logs.

```bash
docker compose logs sub2api | grep "admin password"
```

Before exposing the service publicly, set a stable `JWT_SECRET`, use strong
database credentials, restrict the host firewall, and terminate TLS at your
reverse proxy. See the [deployment guide](deploy/README.md) for configuration,
upgrades, backups, reverse-proxy guidance, and alternative deployment methods.

## Develop from source

Prerequisites: a Go toolchain compatible with [backend/go.mod](backend/go.mod),
Node.js with pnpm, PostgreSQL, and Redis.

```bash
git clone https://github.com/zuixi01/sub2api.git
cd sub2api

# Build the frontend into the backend's embedded web assets.
cd frontend
pnpm install
pnpm run build

# Build a server that embeds the frontend.
cd ../backend
go build -tags embed -o sub2api ./cmd/server
```

For local development, configure PostgreSQL and Redis first, then run the
backend and frontend in separate terminals:

```bash
cd backend
go run ./cmd/server
```

```bash
cd frontend
pnpm run dev
```

Read the [development guide](DEV_GUIDE.md) for the project toolchain, test
commands, code-generation workflow, and local-environment notes.

## Documentation

| Topic | Where to start |
| --- | --- |
| Docker Compose, backups, upgrades, and operations | [Deployment guide](deploy/README.md) |
| Docker image usage | [Docker guide](deploy/DOCKER.md) |
| Environment and server configuration | [Configuration example](deploy/config.example.yaml) |
| Local development and testing | [Development guide](DEV_GUIDE.md) |
| Database migration conventions | [Migration guide](backend/migrations/README.md) |
| Payments | [Payment guide](docs/PAYMENT.md) |
| Chinese documentation | [README_CN.md](README_CN.md) |
| Japanese documentation | [README_JA.md](README_JA.md) |

## Project layout

```text
backend/       Go service, gateway, persistence, migrations, and embedded web assets
frontend/      Vue 3 dashboard
deploy/        Docker Compose files, deployment scripts, and runtime examples
docs/          Feature and integration documentation
```

## Contributing

Please review the [development guide](DEV_GUIDE.md) before changing the
project. Keep changes focused, run the relevant backend or frontend checks, and
include documentation updates when behavior or configuration changes.

## License

Sub2API is licensed under the [GNU Lesser General Public License v3.0 or
later](LICENSE).
