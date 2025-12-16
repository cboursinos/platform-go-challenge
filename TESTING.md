# Testing Guide

This document describes how to run the test suite for the Favorites API.

## Test Structure

The test suite includes:

1. **Unit Tests** (`memory_test.go`): Tests for in-memory storage (no external dependencies)
2. **MySQL Integration Tests** (`mysql_test.go`): Tests for MySQL storage (requires MySQL)
3. **Redis Integration Tests** (`redis_cache_test.go`): Tests for Redis cache (requires Redis)
4. **Cached Storage Integration Tests** (`cached_storage_test.go`): Tests for MySQL + Redis combined (requires both)

## Running Tests

### All Tests (Unit Tests Only)

Run all tests that don't require external dependencies:

```bash
go test ./... -v
```

### With Integration Tests

To run integration tests, you need MySQL and Redis running. You can use Docker Compose:

```bash
# Start MySQL and Redis
docker-compose up -d mysql redis

# Run all tests including integration tests
go test ./... -v

# Or run specific test suites
go test ./internal/storage -v
```

### Skip Integration Tests

To skip integration tests (useful for CI/CD when databases aren't available):

```bash
go test ./... -short
```

Integration tests are marked with `if testing.Short() { t.Skip(...) }` and will be skipped when using the `-short` flag.

## Environment Variables

You can configure test database connections using environment variables:

### MySQL

```bash
export TEST_MYSQL_DSN="user:password@tcp(localhost:3306)/favorites_test?charset=utf8mb4&parseTime=True&loc=Local"
```

Default: `root:password@tcp(localhost:3306)/favorites_test?charset=utf8mb4&parseTime=True&loc=Local`

### Redis

```bash
export TEST_REDIS_ADDR="localhost:6379"
export TEST_REDIS_PASSWORD=""  # Optional, leave empty if no password
```

Default: `localhost:6379` with no password

## Test Coverage

Run tests with coverage:

```bash
go test ./... -cover
```

Generate HTML coverage report:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Running Specific Tests

### MySQL Tests Only

```bash
go test ./internal/storage -v -run TestMySQLStorage
```

### Redis Tests Only

```bash
go test ./internal/storage -v -run TestRedisCache
```

### Cached Storage Tests

```bash
go test ./internal/storage -v -run TestCachedStorage
```

## Test Database Setup

The MySQL tests automatically create the required schema. Make sure:

1. MySQL is running
2. The test database exists (or can be created)
3. The user has CREATE/DROP permissions

The tests use a separate database (`favorites_test`) to avoid interfering with production data.

## Continuous Integration

For CI/CD pipelines, you can:

1. Use Docker Compose to start test databases
2. Run tests with `-short` flag to skip integration tests
3. Use testcontainers-go for isolated test environments (future enhancement)

Example CI script:

```bash
#!/bin/bash
set -e

# Start test databases
docker-compose up -d mysql redis

# Wait for databases to be ready
sleep 5

# Run tests
go test ./... -v

# Cleanup
docker-compose down
```

## Troubleshooting

### MySQL Connection Errors

- Verify MySQL is running: `docker ps` or `systemctl status mysql`
- Check connection string format
- Ensure database exists or user has CREATE DATABASE permission

### Redis Connection Errors

- Verify Redis is running: `redis-cli ping`
- Check firewall settings
- Verify Redis address and port

### Test Failures

- Ensure test databases are clean (tests may fail if data from previous runs exists)
- Check for port conflicts
- Verify environment variables are set correctly

## Writing New Tests

When adding new tests:

1. Use `testing.Short()` to skip integration tests when appropriate
2. Use helper functions from `test_helpers.go` for database connections
3. Clean up test data after tests (or use a separate test database)
4. Use descriptive test names following the pattern: `Test<Component>_<Scenario>`

Example:

```go
func TestMySQLStorage_NewFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping MySQL integration test in short mode")
    }
    
    storage, err := NewMySQLStorage(getTestMySQLDSN())
    if err != nil {
        t.Fatalf("Failed to create storage: %v", err)
    }
    defer storage.Close()
    
    // Test implementation
}
```



