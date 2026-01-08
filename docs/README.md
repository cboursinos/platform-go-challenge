# Swagger Documentation

This directory contains the Swagger/OpenAPI documentation for the Favorites API.

## Accessing Swagger Documentation

### Option 1: Standalone HTML File (Recommended)

Open `swagger.html` directly in your browser:
```bash
# On Linux/Mac
open docs/swagger.html
# or
xdg-open docs/swagger.html

# On Windows
start docs/swagger.html
```

The standalone HTML file contains the complete Swagger UI and can be opened without running the server.

### Option 2: Via Server (Interactive Testing)

1. Start the server:
   ```bash
   go run cmd/server/main.go
   ```

2. Access Swagger UI at: http://localhost:8080/swagger/index.html

   This version allows you to test API endpoints directly from the browser.

## Documentation Files

- `swagger.html` - Standalone HTML file with embedded Swagger UI (open in browser)
- `swagger.json` - OpenAPI 2.0 specification in JSON format
- `swagger.yaml` - OpenAPI 2.0 specification in YAML format
- `docs.go` - Generated Go code containing the Swagger annotations
