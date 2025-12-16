# Environment Variables Setup Guide

This guide explains how to configure the Favorites API using environment variables.

## Quick Start

1. **Copy the example environment file:**
   ```bash
   cp .env.example .env
   ```

2. **Edit `.env` with your configuration:**
   ```bash
   nano .env  # or use your preferred editor
   ```

3. **Run the application:**
   ```bash
   go run ./cmd/server
   ```

## Environment Variables

### Server Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PORT` | Port the server will listen on | `8080` | No |

### MySQL Database Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `MYSQL_DSN` | MySQL Data Source Name | `root:password@tcp(localhost:3306)/favorites_db?charset=utf8mb4&parseTime=True&loc=Local` | No |

**DSN Format:**
```
username:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local
```

**Examples:**
- Local MySQL: `root:password@tcp(localhost:3306)/favorites_db?charset=utf8mb4&parseTime=True&loc=Local`
- Remote MySQL: `user:pass@tcp(mysql.example.com:3306)/favorites_db?charset=utf8mb4&parseTime=True&loc=Local`
- With SSL: `user:pass@tcp(host:3306)/db?tls=true&charset=utf8mb4&parseTime=True&loc=Local`

### Redis Cache Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `REDIS_ADDR` | Redis server address (host:port) | `localhost:6379` | No |
| `REDIS_PASSWORD` | Redis password (leave empty if no password) | `` (empty) | No |
| `REDIS_DB` | Redis database number | `0` | No |

**Examples:**
- Local Redis: `localhost:6379`
- Remote Redis: `redis.example.com:6379`
- Redis with password: Set `REDIS_PASSWORD=yourpassword`

### New Relic APM Configuration (Optional)

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `NEW_RELIC_LICENSE_KEY` | New Relic license key | `` (empty) | No |
| `NEW_RELIC_APP_NAME` | Application name in New Relic dashboard | `Favorites API` | No |

**To enable New Relic:**
1. Sign up at https://newrelic.com
2. Get your license key from the API Keys section
3. Set both `NEW_RELIC_LICENSE_KEY` and `NEW_RELIC_APP_NAME`

**Note:** New Relic is optional. If not configured, the application works normally without monitoring.

### JWT Authentication (Optional)

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `JWT_SECRET` | Secret key for JWT token signing | `` (empty) | No |

**Note:** If not set, a random key will be generated (not recommended for production).

### Development/Testing

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DEBUG` | Enable debug logging | `false` | No |
| `ENVIRONMENT` | Environment name (development, staging, production) | `development` | No |

## Setting Environment Variables

### Method 1: Using .env File (Recommended for Development)

1. Copy `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your values

3. The application will automatically read from `.env` if you use a tool like `godotenv` (optional)

### Method 2: Export in Shell

```bash
export PORT=8080
export MYSQL_DSN="root:password@tcp(localhost:3306)/favorites_db?charset=utf8mb4&parseTime=True&loc=Local"
export REDIS_ADDR="localhost:6379"
go run ./cmd/server
```

### Method 3: Inline with Command

```bash
PORT=8080 MYSQL_DSN="root:password@tcp(localhost:3306)/favorites_db?charset=utf8mb4&parseTime=True&loc=Local" go run ./cmd/server
```

### Method 4: Docker Compose

In `docker-compose.yml`, set environment variables:

```yaml
services:
  api:
    environment:
      - PORT=8080
      - MYSQL_DSN=root:password@tcp(mysql:3306)/favorites_db?charset=utf8mb4&parseTime=True&loc=Local
      - REDIS_ADDR=redis:6379
```

## Production Deployment

For production, use your platform's environment variable configuration:

### Heroku
```bash
heroku config:set MYSQL_DSN="your-dsn"
heroku config:set REDIS_ADDR="your-redis"
```

### AWS (ECS/EC2)
Set environment variables in:
- ECS Task Definition
- EC2 User Data
- Systems Manager Parameter Store

### Kubernetes
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: favorites-api-config
data:
  PORT: "8080"
  REDIS_ADDR: "redis:6379"
---
apiVersion: v1
kind: Secret
metadata:
  name: favorites-api-secrets
type: Opaque
stringData:
  MYSQL_DSN: "root:password@tcp(mysql:3306)/favorites_db?charset=utf8mb4&parseTime=True&loc=Local"
  NEW_RELIC_LICENSE_KEY: "your-key"
```

## Security Best Practices

1. **Never commit `.env` files** - They're in `.gitignore`
2. **Use secrets management** in production (AWS Secrets Manager, HashiCorp Vault, etc.)
3. **Rotate credentials regularly**
4. **Use different credentials** for development, staging, and production
5. **Mask sensitive data** in logs (the config package does this automatically)

## Troubleshooting

### Configuration Not Loading

- Check that environment variables are set: `echo $MYSQL_DSN`
- Verify `.env` file exists and has correct format
- Check for typos in variable names (they're case-sensitive)

### Connection Errors

- Verify MySQL is running: `mysql -u root -p`
- Verify Redis is running: `redis-cli ping`
- Check DSN format matches examples above
- Ensure firewall allows connections

### New Relic Not Working

- Verify both `NEW_RELIC_LICENSE_KEY` and `NEW_RELIC_APP_NAME` are set
- Check license key is valid
- Wait a few minutes for data to appear in dashboard

## Example .env File

```bash
# Server
PORT=8080

# MySQL
MYSQL_DSN=root:password@tcp(localhost:3306)/favorites_db?charset=utf8mb4&parseTime=True&loc=Local

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# New Relic (Optional)
NEW_RELIC_LICENSE_KEY=your-license-key-here
NEW_RELIC_APP_NAME=Favorites API

# JWT (Optional)
JWT_SECRET=your-secret-key-here

# Development
DEBUG=true
ENVIRONMENT=development
```



