# Jenosize Affiliate Platform

Affiliate Web App for Promotion & Marketplace Price Comparison (Lazada / Shopee).

## Features

- **Product & Price Comparison**: Admin can add products via Lazada/Shopee URLs and compare prices across marketplaces
- **Affiliate Link Generator**: Generate short affiliate links with UTM parameters for tracking
- **Campaign Management**: Create marketing campaigns with multiple products
- **Click Tracking & Analytics**: Track clicks and view analytics dashboard
- **Public Landing Pages**: Public-facing campaign pages with price comparison
- **Background Jobs**: Automatic price refresh worker (cron-based)

## Tech Stack

- **Backend**: Go 1.21, Echo Framework, GORM, PostgreSQL, Redis
- **Frontend**: Next.js 14, React 18, TypeScript, Tailwind CSS
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **Documentation**: Swagger/OpenAPI (swaggo/swag)
- **CI/CD**: GitHub Actions

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- Make

**Note**: If you have `golang-migrate` installed, ensure it's compiled with PostgreSQL driver:
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.16.2
```
Or the Makefile will install it automatically when running `make mu`.

### First Time Setup

```bash
# 1. Clone repository
git clone <repo-url>
cd jonosize/project

# 2. Initialize project (installs dependencies, starts Docker)
make init

# 3. Run database migrations
make mu

# 4. (Optional) Seed database with demo data
make seed

# 5. Generate Swagger docs
make swagger

# 6. Start project (frontend + backend)
make start
```

### Available Make Commands

```bash
make init          # Initialize project (install deps, setup config, start Docker)
make mu            # Run database migrations up
make md            # Run database migrations down (rollback)
make seed          # Seed database with demo data
make start         # Start both frontend and backend
make start-backend # Start backend only
make start-frontend # Start frontend only
make stop          # Stop Docker services
make clean         # Clean up (stop services, remove volumes)
make test          # Run tests
make lint          # Run linters
make swagger       # Generate Swagger docs
make build         # Build backend binary
make help          # Show all available commands
```

## Project Structure

```
project/
├── cmd/
│   ├── api/          # API server entry point
│   └── seed/         # Database seeding command
├── internal/
│   ├── api/          # HTTP handlers and routes
│   ├── config/       # Configuration management (Viper wrapper)
│   ├── database/     # Database connection (GORM, read/write separation)
│   ├── dto/          # Data Transfer Objects
│   ├── logger/       # Logger abstraction (Zap wrapper)
│   ├── model/        # GORM models
│   ├── repository/   # Data access layer
│   ├── service/      # Business logic layer
│   ├── validator/    # Input validation utilities
│   ├── worker/       # Background jobs
│   └── seed/         # Seeding functions
├── migrations/       # SQL migrations
├── docs/             # Generated Swagger docs
├── configs/          # Configuration files
├── apps/
│   └── web/          # Next.js frontend
└── pkg/
    └── adapters/     # Marketplace adapters (Lazada, Shopee, Mock)

```

## API Documentation

Swagger documentation is available at:

- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **JSON**: `http://localhost:8080/swagger/doc.json`
- **YAML**: `http://localhost:8080/swagger/doc.yaml`

To regenerate Swagger docs:

```bash
make swagger
```

## Configuration

Configuration is managed via JSON file (`configs/config.json`) with support for environment variable overrides.

Example configuration: `configs/config.example.json`

### Required Environment Variables

- `DATABASE_WRITE_PASSWORD`: PostgreSQL write password

### Optional Environment Variables

- `CONFIG_PATH`: Path to config.json directory (default: `./configs`)
- `DATABASE_WRITE_HOST`, `DATABASE_WRITE_PORT`, etc.
- `SERVER_PORT`: Server port (default: 8080)
- `WORKER_PRICE_REFRESH_CRON`: Cron schedule (default: `0 */6 * * *`)

## API Endpoints

### Admin Endpoints (Basic Auth Required)

- `POST /api/products` - Add product from URL
- `GET /api/products/:id/offers` - Get product offers
- `POST /api/campaigns` - Create campaign
- `POST /api/links` - Generate affiliate link
- `GET /api/dashboard` - Get analytics dashboard
- `POST /api/admin/worker/refresh-prices` - Trigger price refresh

### Public Endpoints

- `GET /api/campaigns/:id/public` - Get public campaign details
- `GET /go/:short_code` - Redirect affiliate link (tracks click)

## Frontend URLs

- Home: `http://localhost:3000`
- Admin Products: `http://localhost:3000/admin/products`
- Admin Campaigns: `http://localhost:3000/admin/campaigns`
- Admin Dashboard: `http://localhost:3000/admin/dashboard`
- Public Campaign: `http://localhost:3000/campaign/[campaign-id]`

## Development

### Backend

```bash
cd project
make start-backend
```

Backend runs on `http://localhost:8080`

### Frontend

```bash
cd project/apps/web
npm install
npm run dev
```

Frontend runs on `http://localhost:3000`

### Database Migrations

```bash
# Run migrations up
make mu

# Rollback migrations
make md
```

### Seeding

```bash
make seed
```

This populates the database with sample products, offers, campaigns, and links for development/demo purposes.

## Testing

```bash
# Run all tests
make test

# Run linters
make lint
```

## CI/CD

GitHub Actions workflows:

- **CI**: Runs tests and linters on push/PR
- **Swagger**: Regenerates Swagger docs on handler changes

## License

MIT
