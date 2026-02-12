# Go Pub/Sub Uptime Monitor (Encore + Go)

This workspace follows the guide from:

- [Building an event-driven system in Go using Pub/Sub (DEV)](https://dev.to/encore/building-an-event-driven-system-in-go-using-pubsub-4l0h)

The main app lives in `go-pubs/uptime` and uses Encore to build an event-driven uptime monitoring system with:

- `monitor` service for pings and periodic checks
- `site` service for CRUD over monitored URLs
- Postgres databases for sites and check history
- Pub/Sub topic for up/down transitions
- Slack notifications via webhook subscriber

## Prerequisites

- Go (matching `go.mod`)
- Docker (for local Encore infra)
- [Encore CLI](https://encore.dev/docs/install)

Install Encore CLI:

```bash
# macOS
brew install encoredev/tap/encore

# Linux
curl -L https://encore.dev/install.sh | bash

# Windows (PowerShell)
iwr https://encore.dev/install.ps1 | iex
```

## Quick Start

```bash
cd uptime
encore run
```

Then open:

- Frontend: `http://localhost:4000/frontend/`
- Local dev dashboard: `http://localhost:9400`

## Build Order (Following the Article)

Use this as a checklist while implementing or reviewing the app.

1. **Create `monitor.Ping` endpoint**
   - Path: `monitor/ping.go`
   - Behavior:
     - Add `https://` when URL has no scheme
     - Return `{ "up": true|false }` based on HTTP status (< 400 means up)

2. **Add test for Ping**
   - Path: `monitor/ping_test.go`
   - Run:
     ```bash
     cd uptime
     encore test ./...
     ```

3. **Create `site` service (CRUD with GORM)**
   - Migration: `site/migrations/1_create_tables.up.sql`
   - Service init: `site/service.go`
   - Endpoints:
     - `GET /site/:siteID`
     - `POST /site`
     - `GET /site`
     - `DELETE /site/:siteID`

4. **Record checks in `monitor` database**
   - Migration: `monitor/migrations/1_create_tables.up.sql`
   - Endpoint: `POST /check/:siteID` in `monitor/check.go`
   - Writes to `checks(site_id, up, checked_at)`

5. **Add concurrent `CheckAll` + cron**
   - Endpoint: `POST /checkall`
   - Use `errgroup` with concurrency limit
   - Add cron job every 5 minutes in `monitor/check.go`

6. **Publish transition events with Pub/Sub**
   - Topic + event type: `monitor/alerts.go`
   - Compare current state with previous measurement
   - Publish only on transition (up->down or down->up)

7. **Subscribe from `slack` service**
   - API: `slack.Notify`
   - Subscription handler listens to monitor transition topic
   - Sends Slack message when site goes down/up

## Local API Smoke Tests

Run from `go-pubs/uptime` while `encore run` is active.

```bash
# Ping a URL
curl "http://localhost:4000/ping/google.com"

# Add a monitored site
curl -X POST "http://localhost:4000/site" -d '{"url":"https://encore.dev"}'

# List monitored sites
curl "http://localhost:4000/site"

# Check one site now
curl -X POST "http://localhost:4000/check/1"

# Check all sites now
curl -X POST "http://localhost:4000/checkall"
```

## Slack Setup

Set your webhook secret (local and cloud environments):

```bash
cd uptime
encore secret set --local,dev,prod SlackWebhookURL
```

Optional quick test:

```bash
curl "http://localhost:4000/slack.Notify" -d '{"Text":"Testing Slack webhook"}'
```

## Deploy

From `go-pubs/uptime`:

```bash
git add -A
git commit -m "Add/Update uptime monitor"
git push encore
```

Encore will build, provision infra, and deploy.

## Useful Notes

- Cron jobs do not automatically trigger in local development.
- Pub/Sub + subscribers are typed through Encore's Go SDK.
- Most app code remains standard Go; framework-specific pieces are annotations and infra bindings.
