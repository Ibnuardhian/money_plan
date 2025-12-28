# money_plan (Go Fiber)

Minimal setup using Go Fiber with a health route.

## Prerequisites
- Go 1.18+ installed

## Quick Start

```powershell
# From the project directory
go run .
```

Server starts on `http://localhost:3000` (configurable via `PORT` in `.env`).

## Live Reload (Air)

```powershell
# Install once (if not yet)
go install github.com/air-verse/air@latest

# Ensure Air is on PATH (Windows PowerShell)
$env:PATH += ";$env:USERPROFILE\go\bin"

# Run with config
air
```

Air is configured in `.air.toml` to build `./cmd/web` and exclude `vendor`, `tmp`, and `test` directories.

## Project Structure
- `main.go` — Fiber app with logger, CORS, and `/health` route
- `.env` — environment variables (`PORT`)
- `.gitignore` — common ignore rules

## Next Steps
- Add more routes under `/api` as needed
- Integrate a database (e.g., Postgres) if required
