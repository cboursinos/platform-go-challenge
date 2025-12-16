# Docker Commands Guide

This guide provides commands to build and run the Favorites API using Docker.

## Option 1: Using Docker Compose (Recommended)

Docker Compose automatically sets up MySQL, Redis, and the API together.

### Build and Start All Services

```bash
# Build and start all services (MySQL, Redis, API)
docker-compose up -d

# Or build and start with logs visible
docker-compose up
```

### View Logs

```bash
# View logs from all services
docker-compose logs -f

# View logs from specific service
docker-compose logs -f api
docker-compose logs -f mysql
docker-compose logs -f redis
```

### Stop Services

```bash
# Stop all services (keeps containers)
docker-compose stop

# Stop and remove containers (keeps volumes)
docker-compose down

# Stop and remove containers and volumes (clean slate)
docker-compose down -v
```

### Restart Services

```bash
# Restart all services
docker-compose restart

# Restart specific service
docker-compose restart api
```

### Rebuild After Code Changes

```bash
# Rebuild and restart
docker-compose up -d --build

# Rebuild specific service
docker-compose build api
docker-compose up -d api
```

### Check Service Status

```bash
# List running services
docker-compose ps

# Check health status
docker-compose ps
```

### Access Services

```bash
# Execute command in API container
docker-compose exec api sh

# Execute command in MySQL container
docker-compose exec mysql mysql -uroot -ppassword favorites_db

# Execute command in Redis container
docker-compose exec redis redis-cli
```

## Option 2: Using Docker Directly

### Build the Docker Image

```bash
# Build the image
docker build -t favorites-api .

# Build with specific tag
docker build -t favorites-api:latest .

# Build without cache (fresh build)
docker build --no-cache -t favorites-api .
```

### Run the Container

**Note:** You'll need MySQL and Redis running separately for this option.

```bash
# Run with environment variables
docker run -d \
  --name favorites-api \
  -p 8080:8080 \
  -e PORT=8080 \
  -e MYSQL_DSN="root:password@tcp(host.docker.internal:3306)/favorites_db?charset=utf8mb4&parseTime=True&loc=Local" \
  -e REDIS_ADDR="host.docker.internal:6379" \
  favorites-api

# Run with .env file
docker run -d \
  --name favorites-api \
  -p 8080:8080 \
  --env-file .env \
  favorites-api

# Run with interactive mode (see logs)
docker run -it \
  --name favorites-api \
  -p 8080:8080 \
  --env-file .env \
  favorites-api
```

### View Logs

```bash
# View logs
docker logs favorites-api

# Follow logs (like tail -f)
docker logs -f favorites-api
```

### Stop and Remove Container

```bash
# Stop container
docker stop favorites-api

# Remove container
docker rm favorites-api

# Stop and remove in one command
docker rm -f favorites-api
```

### Access Container Shell

```bash
# Execute command in running container
docker exec -it favorites-api sh
```

## Quick Start Commands

### Full Setup with Docker Compose

```bash
# 1. Build and start everything
docker-compose up -d --build

# 2. Check status
docker-compose ps

# 3. View API logs
docker-compose logs -f api

# 4. Test the API
curl http://localhost:8080/api/v1/health
curl http://localhost:8080/api/v1/users/user1/favorites

# 5. Stop everything
docker-compose down
```

### Development Workflow

```bash
# Start services
docker-compose up -d mysql redis

# Run API locally (connects to Docker MySQL/Redis)
go run ./cmd/server

# Or run API in Docker
docker-compose up -d --build api
```

## Environment Variables in Docker

### Using docker-compose.yml

Environment variables are already configured in `docker-compose.yml`. To customize:

1. Edit `docker-compose.yml` directly, or
2. Use environment variable substitution:

```yaml
environment:
  - MYSQL_DSN=${MYSQL_DSN:-root:password@tcp(mysql:3306)/favorites_db?charset=utf8mb4&parseTime=True&loc=Local}
```

### Using .env File with Docker Compose

Docker Compose automatically reads `.env` file:

```bash
# Create .env file
cp .env.example .env

# Edit .env with your values
nano .env

# Docker Compose will use .env automatically
docker-compose up -d
```

### Passing Environment Variables

```bash
# With docker run
docker run -d \
  -e PORT=8080 \
  -e MYSQL_DSN="your-dsn" \
  favorites-api

# With docker-compose
docker-compose run -e PORT=3000 api
```

## Troubleshooting

### Check Container Logs

```bash
# All services
docker-compose logs

# Specific service
docker-compose logs api
docker-compose logs mysql
docker-compose logs redis
```

### Check Container Status

```bash
# List all containers
docker ps -a

# Check specific container
docker inspect favorites-api
```

### Restart After Changes

```bash
# Rebuild and restart
docker-compose up -d --build

# Force recreate containers
docker-compose up -d --force-recreate
```

### Database Connection Issues

```bash
# Check MySQL is running
docker-compose ps mysql

# Test MySQL connection
docker-compose exec mysql mysql -uroot -ppassword -e "SELECT 1"

# Check MySQL logs
docker-compose logs mysql
```

### Redis Connection Issues

```bash
# Check Redis is running
docker-compose ps redis

# Test Redis connection
docker-compose exec redis redis-cli ping

# Check Redis logs
docker-compose logs redis
```

### Port Conflicts

If ports are already in use:

```bash
# Check what's using the port
lsof -i :8080
lsof -i :3306
lsof -i :6379

# Change ports in docker-compose.yml
ports:
  - "8081:8080"  # Use 8081 instead of 8080
```

### Clean Everything

```bash
# Stop and remove all containers, networks, and volumes
docker-compose down -v

# Remove all images
docker-compose down --rmi all

# Complete cleanup (be careful!)
docker system prune -a --volumes
```

## Production Deployment

### Build for Production

```bash
# Build optimized image
docker build -t favorites-api:production .

# Tag for registry
docker tag favorites-api:production your-registry/favorites-api:v1.0.0

# Push to registry
docker push your-registry/favorites-api:v1.0.0
```

### Run in Production

```bash
# Use production environment variables
docker run -d \
  --name favorites-api \
  -p 8080:8080 \
  --env-file .env.production \
  --restart unless-stopped \
  favorites-api:production
```

## Useful Docker Commands Reference

```bash
# List images
docker images

# Remove image
docker rmi favorites-api

# List containers
docker ps -a

# Remove stopped containers
docker container prune

# View resource usage
docker stats

# Inspect container
docker inspect favorites-api

# Copy file to/from container
docker cp file.txt favorites-api:/path/
docker cp favorites-api:/path/file.txt ./
```

