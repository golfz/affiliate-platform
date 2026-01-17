# QUICKSTART (Local Development)

This guide helps you run the project locally in ~10 minutes.

## Prerequisites

- Go 1.21+
- Node.js 18+
- Docker + Docker Compose
- Make

## Step 0 (optional): Preflight check

```bash
./CHECK_SETUP.sh
```

## Step 1: Initialize (first time only)

```bash
make init
```

What it does:

- installs Go modules
- installs Web deps (`apps/web`)
- creates `configs/config.json` from `configs/config.example.json` (if missing)
- starts Docker services (Postgres + Redis)

## Step 2: Run database migrations

```bash
make mu
```

## Step 3 (optional): Configure Web API base URL

By default the web app calls `http://localhost:8080`. If you want to override it, create `apps/web/.env.local`
(this file is intentionally ignored by Git) from the template:

```bash
cp apps/web/env.example apps/web/.env.local
```

## Step 4: Start the stack

### Option A: start both API + Web

```bash
make start
```

- API: `http://localhost:8080`
- Web: `http://localhost:3000`

### Option B (recommended for debugging): start separately

Terminal 1:

```bash
make start-backend
```

Terminal 2:

```bash
make start-frontend
```

## Step 5: Smoke tests

### 5.1 Health check

```bash
curl http://localhost:8080/health
```

### 5.2 Swagger

Open:

`http://localhost:8080/swagger/index.html`

### 5.3 Web pages

- Admin Products: `http://localhost:3000/admin/products`
- Admin Campaigns: `http://localhost:3000/admin/campaigns`
- Admin Dashboard: `http://localhost:3000/admin/dashboard`

## Step 6: End-to-end flow (API cURL)

All `/api/*` admin endpoints require Basic Auth:

```bash
AUTH="-u admin:admin123"
```

### 6.1 Add product (mock adapter will map URLs to fixture data)

```bash
curl -X POST http://localhost:8080/api/products \
  $AUTH \
  -H "Content-Type: application/json" \
  -d '{
    "source": "https://www.lazada.co.th/products/pdp-i3603170719-s13480882463.html",
    "sourceType": "url",
    "lazada_url": "https://www.lazada.co.th/products/pdp-i3603170719-s13480882463.html",
    "shopee_url": "https://shopee.co.th/liferinger.th/26379553660"
  }'
```

### 6.2 List products and offers

```bash
curl http://localhost:8080/api/products $AUTH
```

### 6.3 Create a campaign

```bash
curl -X POST http://localhost:8080/api/campaigns \
  $AUTH \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Summer Deal 2025",
    "utm_campaign": "summer_2025",
    "start_at": "2026-01-01T00:00:00Z",
    "end_at": "2026-12-31T23:59:59Z",
    "product_ids": []
  }'
```

### 6.4 Create an affiliate short link

```bash
curl -X POST http://localhost:8080/api/links \
  $AUTH \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "<PRODUCT_ID>",
    "campaign_id": "<CAMPAIGN_ID>",
    "marketplace": "lazada"
  }'
```

### 6.5 Public campaign JSON

```bash
curl http://localhost:8080/api/campaigns/<CAMPAIGN_ID>/public
```

### 6.6 Trigger price refresh (manual)

```bash
curl -X POST http://localhost:8080/api/worker/refresh-prices $AUTH
```

### 6.7 Dashboard stats

```bash
curl http://localhost:8080/api/dashboard $AUTH
```

## Troubleshooting

### Database connection failed

```bash
docker-compose ps
docker-compose up -d
docker-compose logs postgres
```

### Port already in use (8080 / 3000)

```bash
lsof -i :8080
lsof -i :3000
```

### Migration failed: "unknown driver postgres"

The Makefile auto-installs `migrate` with the Postgres driver on `make mu`, but if you previously installed it without tags:

```bash
rm -f "$(go env GOPATH)/bin/migrate"
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.16.2
make mu
```

## Useful commands

```bash
make test
make lint
make swagger
make stop
make clean
```
