# Postman Collection Setup Guide

This guide explains how to import and use the Postman collection for the Favorites API.

## Importing the Collection

### Method 1: Import from File

1. Open Postman
2. Click **Import** button (top left)
3. Select **File** tab
4. Choose `Favorites_API.postman_collection.json`
5. Click **Import**

### Method 2: Import from URL (if hosted)

1. Open Postman
2. Click **Import** button
3. Select **Link** tab
4. Paste the collection URL
5. Click **Import**

## Collection Structure

The collection is organized into the following folders:

### 1. Health Check
- **Health Check**: Verify API is running

### 2. Users
- **Get Favorites - user1** through **user5**: Get favorites for each demo user

### 3. Add Favorite
- **Add Chart Favorite**: Add a new chart asset
- **Add Insight Favorite**: Add a new insight asset
- **Add Audience Favorite**: Add a new audience asset
- **Add Chart/Insight/Audience with Demo Data**: Examples using pre-populated demo data

### 4. Remove Favorite
- Examples for removing different asset types from favorites

### 5. Update Description
- Examples for updating asset descriptions

### 6. Demo Data Examples
- **All Chart Examples**: chart1, chart2, chart3, chart4
- **All Insight Examples**: insight1 through insight6
- **All Audience Examples**: audience1 through audience6

## Configuration

### Setting the Base URL

1. Click on the collection name: **Favorites API**
2. Go to the **Variables** tab
3. Set `base_url` to your API endpoint:
   - Local: `http://localhost:8080`
   - Docker: `http://localhost:8080`
   - Production: `https://your-api-domain.com`

### Environment Variables (Optional)

You can create a Postman Environment for different configurations:

1. Click **Environments** (left sidebar)
2. Click **+** to create new environment
3. Add variables:
   - `base_url`: `http://localhost:8080`
   - `user_id`: `user1` (for quick testing)
4. Save and select the environment

## Demo Users

The collection uses these demo users (pre-populated in database):

| User Reference | Name | Email | Internal ID |
|----------------|------|-------|-------------|
| `user1` | Alice Johnson | alice.johnson@example.com | 1 (auto-increment) |
| `user2` | Bob Smith | bob.smith@example.com | 2 (auto-increment) |
| `user3` | Charlie Brown | charlie.brown@example.com | 3 (auto-increment) |
| `user4` | Diana Prince | diana.prince@example.com | 4 (auto-increment) |
| `user5` | Eve Wilson | eve.wilson@example.com | 5 (auto-increment) |

**Note:** The API uses user "reference" values (e.g., "user1") in URLs, not internal database IDs. The reference field provides stable, human-readable identifiers while the database uses efficient INT primary keys internally.

## Demo Assets

### Charts
- `chart1`: Monthly Sales Revenue Overview
- `chart2`: Social Media Usage by Platform
- `chart3`: E-commerce Conversion Rates
- `chart4`: Customer Age Distribution

### Insights
- `insight1`: Social Media Usage Statistics
- `insight2`: E-commerce Trends
- `insight3`: Remote Work Impact
- `insight4`: Consumer Behavior
- `insight5`: Digital Marketing
- `insight6`: Technology Adoption

### Audiences
- `audience1`: Tech-Savvy Millennials
- `audience2`: Social Media Enthusiasts
- `audience3`: Professional Networkers
- `audience4`: Young Digital Natives
- `audience5`: International Shoppers
- `audience6`: Mature Consumers

## Quick Start

1. **Start the API:**
   ```bash
   # Using Docker Compose
   docker-compose up -d
   
   # Or locally
   go run ./cmd/server
   ```

2. **Import the collection** into Postman

3. **Set the base URL** in collection variables

4. **Test Health Check:**
   - Run "Health Check" request
   - Should return `200 OK` with status "healthy"

5. **Get Demo User Favorites:**
   - Run "Get Favorites - user1"
   - Should return favorites for Alice Johnson

6. **Add a Favorite:**
   - Run "Add Chart Favorite"
   - Note: Asset IDs are automatically generated in the format `{userReference}_favorite{number}` (e.g., `user1_favorite1`, `user1_favorite2`)
   - Any `id` field in the request body will be ignored
   - Then run "Get Favorites - user1" again to see the new favorite

## Testing Workflow

### Complete CRUD Workflow

1. **Create (Add Favorite):**
   - Use "Add Chart Favorite" or any Add request
   - Verify with "Get Favorites"

2. **Read (Get Favorites):**
   - Use "Get Favorites - user1" (or any user)
   - Check response contains the favorite

3. **Update (Update Description):**
   - Use "Update Description - chart1"
   - Verify with "Get Favorites" to see updated description

4. **Delete (Remove Favorite):**
   - Use "Remove Favorite - chart1 from user1"
   - Verify with "Get Favorites" to confirm removal

## Tips

### Using Collection Runner

1. Select the collection
2. Click **Run** button
3. Select requests to run
4. Click **Run Favorites API**
5. View results and test results

### Creating Tests

You can add tests to requests:

```javascript
// Example test in Postman
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Response has favorites array", function () {
    var jsonData = pm.response.json();
    pm.expect(jsonData).to.have.property('favorites');
    pm.expect(jsonData.favorites).to.be.an('array');
});
```

### Using Pre-request Scripts

Set variables dynamically:

```javascript
// Set user_id from environment
pm.variables.set("user_id", pm.environment.get("user_id") || "user1");
```

## Troubleshooting

### Connection Refused

- Verify API is running: `curl http://localhost:8080/api/v1/health`
- Check `base_url` variable in collection
- Verify port matches your server configuration

### 404 Not Found

- Check the URL path matches your API routes
- Verify collection variable `base_url` is set correctly

### 500 Internal Server Error

- Check API logs
- Verify MySQL and Redis are running
- Check database connection in environment variables

### Empty Favorites Response

- Verify demo data is loaded in database
- Check user ID exists (user1-user5)
- Run migration script if needed

## Collection Updates

To update the collection:

1. Export current collection from Postman
2. Compare with `Favorites_API.postman_collection.json`
3. Merge changes or replace if needed
4. Re-import into Postman

## Sharing the Collection

The collection can be shared with your team:

1. Export from Postman: **Collection** → **Export**
2. Commit to version control
3. Team members import the file

Or use Postman's sharing features:
- Postman Workspace
- Postman API Network
- Direct file sharing

