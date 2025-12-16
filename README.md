# GWI Platform Go Challenge - Favorites API

A high-performance, production-ready REST API for managing user favorites (assets) built with Go. This solution demonstrates Go programming concepts including concurrency, security, performance optimization, and clean architecture.

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
- [Running the Application](#running-the-application)
- [Testing](#testing)
- [Docker](#docker)
- [Storage Options](#storage-options)
- [Performance](#performance)
- [Security Features](#security-features)
- [Project Structure](#project-structure)

## Features

### Core Functionality
- ✅ **Get All Lists**: Retrieve all favorites lists for a user. Created in order to allow users to have more sections of favorites.
- ✅ **Get User Favorites**: Retrieve all favorites for a user (with list support)
- ✅ **Add Favorite**: Add any asset type (Chart, Insight, Audience) to favorites in a specific list
- ✅ **Remove Favorite**: Remove an asset from favorites in a specific list
- ✅ **Update Description**: Edit asset descriptions
- ✅ **Multiple Lists**: Support for organizing favorites into multiple lists (e.g., "default", "work", "personal")

### Advanced Features
- 🚀 **High Performance**: MySQL database with optimized indexes + Redis caching layer
- 💾 **Persistent Storage**: MySQL with JSON columns for flexible asset data. Also can be used Mongo for more flexibility about data structure of assets. Mysql for analytics purposes
- ⚡ **Fast Retrieval**: Redis cache with cache-aside pattern and automatic invalidation
- 🔒 **Security**: JWT authentication (optional), rate limiting, security headers, input validation
- 🏥 **High Availability**: Health checks, graceful shutdown, panic recovery
- 📊 **Observability**: Request logging, structured error responses
- 🧪 **Testing**: Comprehensive unit tests with coverage of core functionality
- 🐳 **Containerization**: Dockerfile with multi-stage build and health checks

## Architecture

The application follows a clean, layered architecture:

```
┌─────────────────┐
│   HTTP Layer    │  (handlers, middleware)
├─────────────────┤
│  Service Layer  │  (business logic, validation)
├─────────────────┤
│  Storage Layer  │  (data persistence)
└─────────────────┘
```

### Components

1. **Models** (`internal/models/`): Data structures for assets (Chart, Insight, Audience) with interface-based polymorphism
2. **Storage** (`internal/storage/`): 
   - MySQL storage with optimized indexes for persistent data
   - Redis cache layer for high-performance reads
   - Cached storage wrapper implementing cache-aside pattern
3. **Service** (`internal/service/`): Business logic, validation, and asset creation from JSON
4. **API** (`internal/api/`): HTTP handlers and middleware (auth, rate limiting, logging, security headers)
5. **Auth** (`pkg/auth/`): JWT token generation and validation utilities

## Getting Started

### Prerequisites

- Go 1.18 or higher
- MySQL 5.7+ or MySQL 8.0+
- Redis 6.0+ or Redis 7.0+
- (Optional) Docker and Docker Compose

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd platform-go-challenge
```

2. Install dependencies:
```bash
go mod download
```

3. Set up MySQL database:
```bash
# Create database with schema and demo data (recommended)
mysql -u root -p < migrations/001_init_schema.sql

# Or create manually:
mysql -u root -p
CREATE DATABASE favorites_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

**Note:** The migration file includes demo data:
- **5 demo users** with references (user1-user5), email addresses, and names
  - Users have an auto-increment INT `id` (internal) and a VARCHAR `reference` (used in API)
  - API uses the `reference` field (e.g., "user1", "user2") instead of internal IDs
- **4 Chart assets** (chart1-chart4) with sample data
- **6 Insight assets** (insight1-insight6) with sample insights
- **6 Audience assets** (audience1-audience6) with demographic data
- **Sample favorites** linking users to various assets in different lists
  - Most favorites are in the "default" list
  - User1 has one favorite in the "work" list (chart3)

You can test the API immediately using these demo user references:
- `user1` through `user5` (these are the `reference` values, not internal IDs)
- Lists: `default` (most favorites), `work` (user1 has chart3 in work list)

4. Start Redis server:
```bash
# Using Docker:
docker run -d -p 6379:6379 redis:7-alpine

# Or using local installation:
redis-server
```

## Running the Application

### Local Development

1. Build the application:
```bash
go build -o server ./cmd/server
```

2. Run the server:
```bash
./server
```

Or run directly:
```bash
go run ./cmd/server
```

The server will start on port `8080` by default. You can change this by setting the `PORT` environment variable:

```bash
PORT=3000 go run ./cmd/server
```

### Environment Variables

The application uses environment variables for configuration. See [ENV_SETUP.md](ENV_SETUP.md) for detailed documentation.

**Quick Setup:**
```bash
# Copy the example environment file
cp .env.example .env

# Edit .env with your configuration
nano .env

# Run the application
go run ./cmd/server
```

**Available Environment Variables:**

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PORT` | Server port | `8080` | No |
| `MYSQL_DSN` | MySQL Data Source Name | `root:password@tcp(localhost:3306)/favorites_db?charset=utf8mb4&parseTime=True&loc=Local` | No |
| `REDIS_ADDR` | Redis server address | `localhost:6379` | No |
| `REDIS_PASSWORD` | Redis password | `` (empty) | No |
| `REDIS_DB` | Redis database number | `0` | No |
| `NEW_RELIC_LICENSE_KEY` | New Relic license key | `` (empty) | No |
| `NEW_RELIC_APP_NAME` | New Relic application name | `Favorites API` | No |
| `JWT_SECRET` | JWT secret key | `` (empty) | No |
| `DEBUG` | Enable debug logging | `false` | No |
| `ENVIRONMENT` | Environment name | `development` | No |

**Example:**
```bash
export MYSQL_DSN="user:pass@tcp(localhost:3306)/favorites_db?charset=utf8mb4&parseTime=True&loc=Local"
export REDIS_ADDR="localhost:6379"
export NEW_RELIC_LICENSE_KEY="your-license-key"
export NEW_RELIC_APP_NAME="Favorites API"
go run ./cmd/server
```

For detailed configuration instructions, see [ENV_SETUP.md](ENV_SETUP.md).

## Testing

Run all tests:
```bash
go test ./... -v
```

Run tests with coverage:
```bash
go test ./... -cover
```

Run tests for a specific package:
```bash
go test ./internal/storage -v
go test ./internal/service -v
go test ./internal/api -v
```

### Test Coverage

The test suite includes:
- Unit tests for storage layer (including concurrent access tests)
- Unit tests for service layer (validation, business logic)
- Integration tests for API handlers
- Tests for all asset types (Chart, Insight, Audience)

**Note:** MySQL+Redis benchmarks require:
- MySQL running (default: `localhost:3306`, database: `favorites_test`)
- Redis running (default: `localhost:6379`)
- Set environment variables if using different connection details:
  ```bash
  export TEST_MYSQL_DSN="user:pass@tcp(localhost:3306)/favorites_test?charset=utf8mb4&parseTime=True&loc=Local"
  export TEST_REDIS_ADDR="localhost:6379"
  export TEST_REDIS_PASSWORD=""  # Optional, if Redis has a password
  ```

## Docker

### Quick Start with Docker Compose (Recommended)

The easiest way to run the entire stack (MySQL, Redis, and API):

```bash
# Build and start all services
docker-compose up -d --build

# View logs
docker-compose logs -f api

# Test the API
curl http://localhost:8080/api/v1/health

# Stop all services
docker-compose down
```

**What this does:**
1. Starts MySQL with schema automatically initialized (from `migrations/` folder)
2. Starts Redis
3. Builds and starts the API server connected to both
4. All services have health checks and wait for dependencies

### Using Docker Directly

**Build the image:**
```bash
docker build -t favorites-api .
```

**Run the container:**
```bash
# With environment variables (requires MySQL/Redis running separately)
docker run -d \
  --name favorites-api \
  -p 8080:8080 \
  -e MYSQL_DSN="root:password@tcp(host.docker.internal:3306)/favorites_db?charset=utf8mb4&parseTime=True&loc=Local" \
  -e REDIS_ADDR="host.docker.internal:6379" \
  favorites-api

# Or with .env file
docker run -d --name favorites-api -p 8080:8080 --env-file .env favorites-api
```

**Note:** When using Docker directly, ensure MySQL and Redis are accessible (use `host.docker.internal` for local services).

### Docker Commands Reference

For detailed Docker commands and troubleshooting, see [DOCKER.md](DOCKER.md).

**Common commands:**
```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose stop

# Remove containers and volumes
docker-compose down -v

# Rebuild after code changes
docker-compose up -d --build
```

### Docker Compose Configuration

Create a `docker-compose.yml`:

```yaml
version: '3.8'
services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: favorites_db
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5

  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - MYSQL_DSN=root:password@tcp(mysql:3306)/favorites_db?charset=utf8mb4&parseTime=True&loc=Local
      - REDIS_ADDR=redis:6379
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/api/v1/health"]
      interval: 30s
      timeout: 3s
      retries: 3

volumes:
  mysql_data:
  redis_data:
```

Run with:
```bash
docker-compose up -d
```

This will:
1. Start MySQL with the schema automatically initialized
2. Start Redis
3. Start the API server connected to both

## Storage Architecture

### Current Implementation: MySQL + Redis

The application uses a **two-tier storage architecture**:

1. **MySQL Database**: Primary persistent storage with optimized indexes
2. **Redis Cache**: High-performance caching layer with cache-aside pattern

#### MySQL Storage

The MySQL database stores all persistent data with the following optimizations:

**Tables:**
- `assets`: Stores asset data (Chart, Insight, Audience) with JSON column for flexible schema
- `favorites`: Stores user favorite relationships

**Optimized Indexes:**
- Primary keys for fast lookups
- Composite indexes for common query patterns:
  - `idx_user_added_at`: Optimizes sorting favorites by date for a user
  - `idx_user_updated_at`: Optimizes sorting by update time
  - `idx_type`: Optimizes filtering by asset type
  - Foreign key constraints ensure data integrity

**Connection Pooling:**
- Max open connections: 25
- Max idle connections: 5
- Connection lifetime: 5 minutes

#### Redis Cache

Redis is used for high-performance caching with the following features:

**Cache-Aside Pattern:**
1. Check Redis cache first
2. On cache miss, fetch from MySQL
3. Store result in Redis for future requests
4. Invalidate cache on updates

**Cache Keys:**
- `favorites:user:{userReference}:list:{listReference}`: User's favorites list for a specific list (e.g., "favorites:user:user1:list:list_user1_default")
- `asset:{assetReference}`: Individual asset data (uses asset reference, e.g., "chart1")

**Cache TTL:** 15 minutes (configurable)

**Cache Invalidation:**
- Automatically invalidated on:
  - Adding a favorite
  - Removing a favorite
  - Updating asset description
- Ensures data consistency between cache and database

### Benefits of This Architecture

**Performance:**
- Redis provides sub-millisecond response times for cached data
- MySQL handles complex queries and data persistence
- Reduced database load through intelligent caching

**Scalability:**
- MySQL can handle large datasets with proper indexing
- Redis can be scaled horizontally with Redis Cluster
- Cache reduces database connection pressure

**Reliability:**
- Data persisted in MySQL (no data loss)
- Cache failures gracefully degrade to database queries
- Transaction support ensures data consistency

### Data Representation

The current MySQL data model uses:
- **Assets table**: 
  - `id` (INT AUTO_INCREMENT, PK) - Internal database ID
  - `reference` (VARCHAR, UNIQUE) - Asset reference used in API (e.g., "chart1", "insight1")
  - `type` (ENUM), `description` (TEXT), `data` (JSON), timestamps
  - API endpoints use `reference` instead of internal `id`
- **Lists table**: 
  - `id` (INT AUTO_INCREMENT, PK) - Internal database ID
  - `reference` (VARCHAR, UNIQUE) - List reference used in API (e.g., "list_user1_default")
  - `user_id` (INT, FK to `users.id`) - Internal user ID
  - `name` (VARCHAR) - List name (e.g., "default", "work", "personal")
  - Unique constraint: `(user_id, name)` - Prevents duplicate list names per user
  - `created_at`, `updated_at` timestamps
  - Foreign key with CASCADE delete
  - Each user can have multiple lists for organizing favorites
- **Favorites table**: 
  - `id` (INT AUTO_INCREMENT, PK) - Internal database ID
  - `reference` (VARCHAR, UNIQUE) - Favorite reference used in API (e.g., "fav_user1_favorite1")
  - `user_id` (INT, FK to `users.id`) - Internal user ID
  - `asset_id` (INT, FK to `assets.id`) - Internal asset ID
  - `list_id` (INT, FK to `lists.id`) - Internal list ID
  - Unique constraint: `(user_id, asset_id, list_id)` - Prevents duplicate assets in same list
  - `added_at`, `updated_at` timestamps
  - Foreign keys with CASCADE delete
  - Links users, assets, and lists together

**JSON Storage:**
- Asset-specific data (chart data points, insight text, audience characteristics) stored in JSON column
- Allows flexible schema while maintaining queryability
- MySQL 5.7+ provides JSON functions for querying nested data

**Index Strategy:**
- Primary keys ensure uniqueness and fast lookups
- Composite indexes optimize common query patterns
- Foreign keys maintain referential integrity

## Performance

The API includes comprehensive benchmark tests to measure and validate performance. See the [Benchmarking](#benchmarking) section for details on running benchmarks.

### Optimizations Implemented

1. **MySQL Connection Pooling**
   - Max 25 open connections
   - 5 idle connections maintained
   - Connection lifetime: 5 minutes
   - Prevents connection exhaustion

2. **Optimized Database Indexes**
   - Composite indexes for common query patterns
   - Primary keys for fast lookups
   - Foreign keys with CASCADE for data integrity

3. **Redis Caching**
   - Cache-aside pattern for high-performance reads
   - 15-minute TTL balances freshness and performance
   - Automatic cache invalidation on updates
   - Sub-millisecond response times for cached data

4. **Rate Limiting**
   - Token bucket algorithm (100 req/s, burst 200)
   - Prevents abuse and ensures fair resource usage

5. **Connection Management**
   - Configurable timeouts (read: 15s, write: 15s, idle: 60s)
   - Prevents resource exhaustion

### Scalability Considerations

For horizontal scaling:
- Use a shared data store (Redis, PostgreSQL, etc.)
- Implement distributed rate limiting (Redis-based)
- Add load balancing (nginx, HAProxy)
- Consider caching frequently accessed favorites

## Security Features

1. **JWT Authentication** (optional, currently disabled for testing)
   - Token-based authentication
   - Configurable expiration
   - Secure secret key generation

2. **Rate Limiting**
   - Prevents DDoS and abuse
   - Configurable limits per endpoint

3. **Security Headers**
   - `X-Content-Type-Options: nosniff`
   - `X-Frame-Options: DENY`
   - `X-XSS-Protection: 1; mode=block`
   - `Strict-Transport-Security`
   - `Content-Security-Policy`

4. **Input Validation**
   - Asset type validation
   - Required field checks
   - JSON schema validation

5. **CORS Support**
   - Configurable allowed origins
   - Preflight request handling

6. **Panic Recovery**
   - Graceful error handling
   - Prevents server crashes

## Project Structure

```
platform-go-challenge/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers.go         # HTTP request handlers
│   │   ├── handlers_test.go    # Handler tests
│   │   └── middleware.go       # Middleware (auth, logging, rate limiting)
│   ├── models/
│   │   └── models.go           # Data models (Chart, Insight, Audience)
│   ├── service/
│   │   ├── favorites.go        # Business logic
│   │   └── favorites_test.go  # Service tests
│   └── storage/
│       ├── memory.go           # In-memory storage implementation
│       └── memory_test.go     # Storage tests
├── pkg/
│   └── auth/
│       └── auth.go            # JWT authentication utilities
├── Dockerfile                 # Docker build configuration
├── .dockerignore              # Docker ignore patterns
├── .gitignore                # Git ignore patterns
├── go.mod                     # Go module definition
├── go.sum                     # Go module checksums
└── README.md                  # This file
```

## Postman Collection

A complete Postman collection is available with all endpoints and demo data:

**File:** `Favorites_API.postman_collection.json`

**Quick Import:**
1. Open Postman
2. Click **Import**
3. Select `Favorites_API.postman_collection.json`
4. Set `base_url` variable to `http://localhost:8080`

The collection includes:
- All API endpoints (GET, POST, DELETE, PATCH)
- Demo data for all asset types (Charts, Insights, Audiences)
- Examples for all 5 demo users
- Pre-configured requests with sample payloads

For detailed setup instructions, see [POSTMAN_SETUP.md](POSTMAN_SETUP.md).

## Example Usage

### Demo Data

The database includes pre-populated demo data for testing:

**Demo Users:**
- `user1` - Alice Johnson (alice.johnson@example.com)
- `user2` - Bob Smith (bob.smith@example.com)
- `user3` - Charlie Brown (charlie.brown@example.com)
- `user4` - Diana Prince (diana.prince@example.com)
- `user5` - Eve Wilson (eve.wilson@example.com)

**Demo Assets:**
- **Charts**: chart1 (Monthly Sales Revenue), chart2 (Social Media Usage), chart3 (E-commerce Conversion), chart4 (Customer Age Distribution)
- **Insights**: insight1-insight6 (various market insights)
- **Audiences**: audience1-audience6 (various demographic segments)

**Demo Favorites:** Each user has 3-5 pre-configured favorites.

## New Relic APM Integration

The application includes New Relic Application Performance Monitoring (APM) for tracking execution times and performance metrics.

### Setup

1. **Get New Relic License Key**:
   - Sign up for a New Relic account at https://newrelic.com
   - Navigate to API Keys section
   - Copy your license key

2. **Configure Environment Variables**:
   ```bash
   export NEW_RELIC_LICENSE_KEY="your-license-key-here"
   export NEW_RELIC_APP_NAME="Favorites API"
   ```
## License

This project is part of the GWI Platform Engineering Challenge.

## Author CBO

Built as a demonstration of Go engineering skills, focusing on:
- Clean architecture and code organization
- Small extra Business Logic (Favorite Lists, Asset Managment)
- Performance optimization
- Security best practices
- Production-ready features
- Comprehensive testing
