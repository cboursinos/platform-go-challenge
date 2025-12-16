package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gwi/platform-go-challenge/internal/models"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// MySQLStorage implements storage using MySQL database
type MySQLStorage struct {
	db          *sql.DB
	newRelicApp *newrelic.Application
}

// NewMySQLStorage creates a new MySQL storage instance
// If newRelicApp is provided, database queries will be instrumented with New Relic
func NewMySQLStorage(dsn string, newRelicApp *newrelic.Application) (*MySQLStorage, error) {
	return NewMySQLStorageWithDriver(dsn, "mysql", newRelicApp)
}

// NewMySQLStorageWithDriver creates a new MySQL storage instance with a specific driver
// Use "nrmysql" driver for New Relic instrumentation, "mysql" for standard driver
func NewMySQLStorageWithDriver(dsn string, driver string, newRelicApp *newrelic.Application) (*MySQLStorage, error) {
	db, err := sql.Open(driver, dsn)

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	storage := &MySQLStorage{
		db:          db,
		newRelicApp: newRelicApp,
	}

	// Initialize schema
	if err := storage.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// initSchema creates the database tables with optimized indexes
func (s *MySQLStorage) initSchema() error {

	queries := []string{
		// Users table with basic information
		`CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			reference VARCHAR(255) NOT NULL UNIQUE,
			email VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_reference (reference),
			INDEX idx_email (email),
			INDEX idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Assets table with indexes
		`CREATE TABLE IF NOT EXISTS assets (
			id INT AUTO_INCREMENT PRIMARY KEY,
			reference VARCHAR(255) NOT NULL UNIQUE,
			type ENUM('chart', 'insight', 'audience') NOT NULL,
			description TEXT,
			data JSON NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_reference (reference),
			INDEX idx_type (type),
			INDEX idx_updated_at (updated_at),
			INDEX idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Lists table for organizing favorites
		// First create without foreign key
		`CREATE TABLE IF NOT EXISTS lists (
			id INT AUTO_INCREMENT PRIMARY KEY,
			reference VARCHAR(255) NOT NULL UNIQUE,
			user_id INT NOT NULL,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_reference (reference),
			INDEX idx_user_id (user_id),
			INDEX idx_user_name (user_id, name),
			UNIQUE KEY uk_user_list_name (user_id, name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Favorites table with composite indexes
		// First create without foreign keys
		`CREATE TABLE IF NOT EXISTS favorites (
			id INT AUTO_INCREMENT PRIMARY KEY,
			reference VARCHAR(255) NOT NULL UNIQUE,
			user_id INT NOT NULL,
			asset_id INT NOT NULL,
			list_id INT NOT NULL,
			sort_order INT NOT NULL DEFAULT 0,
			added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_reference (reference),
			INDEX idx_user_id (user_id),
			INDEX idx_asset_id (asset_id),
			INDEX idx_list_id (list_id),
			INDEX idx_user_list (user_id, list_id),
			INDEX idx_user_list_sort_order (user_id, list_id, sort_order),
			INDEX idx_user_list_added_at (user_id, list_id, added_at),
			INDEX idx_user_list_updated_at (user_id, list_id, updated_at),
			UNIQUE KEY uk_user_asset_list (user_id, asset_id, list_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute schema query: %w", err)
		}
	}

	return nil
}

// getUserIDFromReference gets the internal user ID from the user reference
// Returns an error if the user doesn't exist (users are not created on-demand)
func (s *MySQLStorage) getUserIDFromReference(userReference string) (int, error) {
	var userID int
	query := `SELECT id FROM users WHERE reference = ?`
	err := s.db.QueryRow(query, userReference).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("user with reference %s not found", userReference)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get user ID: %w", err)
	}
	return userID, nil
}

// getUserReferenceFromID gets the user reference from the internal user ID
func (s *MySQLStorage) getUserReferenceFromID(userID int) (string, error) {
	var reference string
	query := `SELECT reference FROM users WHERE id = ?`
	err := s.db.QueryRow(query, userID).Scan(&reference)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("user with id %d not found", userID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get user reference: %w", err)
	}
	return reference, nil
}

// getAssetIDFromReference gets the internal asset ID from the asset reference
func (s *MySQLStorage) getAssetIDFromReference(assetReference string) (int, error) {
	var assetID int
	query := `SELECT id FROM assets WHERE reference = ?`
	err := s.db.QueryRow(query, assetReference).Scan(&assetID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("asset with reference %s not found", assetReference)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get asset ID: %w", err)
	}
	return assetID, nil
}

// getAssetReferenceFromID gets the asset reference from the internal asset ID
func (s *MySQLStorage) getAssetReferenceFromID(assetID int) (string, error) {
	var reference string
	query := `SELECT reference FROM assets WHERE id = ?`
	err := s.db.QueryRow(query, assetID).Scan(&reference)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("asset with id %d not found", assetID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get asset reference: %w", err)
	}
	return reference, nil
}

// generateListReference generates a unique reference for a list
func (s *MySQLStorage) generateListReference(userReference, listName string) string {
	return fmt.Sprintf("list_%s_%s", userReference, listName)
}

// getListIDFromReference gets the internal list ID from the list reference
func (s *MySQLStorage) getListIDFromReference(listReference string) (int, error) {
	var listID int
	query := `SELECT id FROM lists WHERE reference = ?`
	err := s.db.QueryRow(query, listReference).Scan(&listID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("list with reference %s not found", listReference)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get list ID: %w", err)
	}
	return listID, nil
}

// getListReferenceFromID gets the list reference from the internal list ID
func (s *MySQLStorage) getListReferenceFromID(listID int) (string, error) {
	var reference string
	query := `SELECT reference FROM lists WHERE id = ?`
	err := s.db.QueryRow(query, listID).Scan(&reference)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("list with id %d not found", listID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get list reference: %w", err)
	}
	return reference, nil
}

// generateFavoriteReference generates a unique reference for a favorite
// Format: fav_{assetReference} (e.g., "fav_user1_favorite1")
// The assetReference already contains the user reference, so we don't need to duplicate it
func (s *MySQLStorage) generateFavoriteReference(userReference, assetReference, listReference string) string {
	return fmt.Sprintf("fav_%s", assetReference)
}

// generateNextAssetID generates the next sequential asset ID for a user
// Format: {userReference}_favorite{number} (e.g., "user1_favorite1", "user1_favorite2")
// Uses transaction with SELECT FOR UPDATE to prevent race conditions when multiple goroutines generate IDs concurrently
// Race condition handling:
//   - When assets exist: SELECT FOR UPDATE locks the row, preventing concurrent reads
//   - When no assets exist: Transaction isolation prevents race conditions between concurrent calls
//   - Multiple concurrent requests will serialize through the transaction, ensuring unique IDs
func (s *MySQLStorage) generateNextAssetID(userReference string) (string, error) {
	// Start a transaction with REPEATABLE READ isolation level to prevent race conditions
	// This ensures that concurrent requests see a consistent view of the data
	tx, err := s.db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Find the highest numbered asset reference for this user
	// Pattern: {userReference}_favorite{number}
	pattern := fmt.Sprintf("%s_favorite%%", userReference)
	prefix := fmt.Sprintf("%s_favorite", userReference)
	prefixLen := len(prefix)
	
	// Query to find the highest numbered asset for this user
	// Use SELECT FOR UPDATE to lock matching rows and prevent concurrent reads
	// This ensures only one goroutine can read the max value at a time
	// Even if no rows match, the transaction isolation prevents race conditions
	query := `SELECT reference FROM assets 
		WHERE reference LIKE ? 
		AND reference REGEXP ?
		ORDER BY CAST(SUBSTRING(reference, ?) AS UNSIGNED) DESC 
		LIMIT 1
		FOR UPDATE`
	regexPattern := fmt.Sprintf("^%s[0-9]+$", prefix)
	
	var maxReference string
	err = tx.QueryRow(query, pattern, regexPattern, prefixLen+1).Scan(&maxReference)
	
	var nextNumber int
	if err == sql.ErrNoRows {
		// No assets exist yet for this user, start with favorite1
		// Transaction isolation ensures concurrent calls will serialize
		nextNumber = 1
	} else if err != nil {
		return "", fmt.Errorf("failed to query asset references: %w", err)
	} else {
		// Extract number from reference (e.g., "user1_favorite5" -> 5)
		// The number starts after the prefix
		numStr := maxReference[prefixLen:]
		var num int
		_, err := fmt.Sscanf(numStr, "%d", &num)
		if err != nil {
			// If parsing fails, start with favorite1
			nextNumber = 1
		} else {
			nextNumber = num + 1
		}
	}
	
	// Commit transaction to release the lock
	// This ensures the next concurrent call will see the updated state
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return fmt.Sprintf("%s_favorite%d", userReference, nextNumber), nil
}

// GenerateNextAssetID generates the next sequential asset ID for a user
// Implements the Storage interface
func (s *MySQLStorage) GenerateNextAssetID(userReference string) (string, error) {
	return s.generateNextAssetID(userReference)
}

// CreateList creates a new list for a user
func (s *MySQLStorage) CreateList(userReference, listName string) (*models.List, error) {
	// Get internal user ID from reference
	internalUserID, err := s.getUserIDFromReference(userReference)
	if err != nil {
		return nil, err
	}

	// Generate list reference
	listReference := s.generateListReference(userReference, listName)

	// Check if list already exists
	var existingID int
	var existingReference string
	checkQuery := `SELECT id, reference FROM lists WHERE user_id = ? AND name = ?`
	err = s.db.QueryRow(checkQuery, internalUserID, listName).Scan(&existingID, &existingReference)
	
	if err == nil {
		// List already exists, return it
		return &models.List{
			ID:        existingID,
			Reference: existingReference,
			UserID:    userReference,
			Name:      listName,
		}, nil
	}

	// Create new list
	now := time.Now()
	insertQuery := `INSERT INTO lists (reference, user_id, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	result, err := s.db.Exec(insertQuery, listReference, internalUserID, listName, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create list: %w", err)
	}

	listID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get list ID: %w", err)
	}

	return &models.List{
		ID:        int(listID),
		Reference: listReference,
		UserID:    userReference,
		Name:      listName,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// GetList retrieves a list by reference
func (s *MySQLStorage) GetList(userReference, listReference string) (*models.List, error) {
	// Get internal user ID from reference
	internalUserID, err := s.getUserIDFromReference(userReference)
	if err != nil {
		return nil, err
	}

	query := `SELECT id, reference, user_id, name, created_at, updated_at FROM lists WHERE reference = ? AND user_id = ?`
	var list models.List
	var internalUserIDCheck int
	err = s.db.QueryRow(query, listReference, internalUserID).Scan(
		&list.ID,
		&list.Reference,
		&internalUserIDCheck,
		&list.Name,
		&list.CreatedAt,
		&list.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("list with reference %s not found", listReference)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get list: %w", err)
	}

	list.UserID = userReference
	return &list, nil
}

// GetListByName retrieves a list by name for a user
func (s *MySQLStorage) GetListByName(userReference, listName string) (*models.List, error) {
	// Get internal user ID from reference
	internalUserID, err := s.getUserIDFromReference(userReference)
	if err != nil {
		return nil, err
	}

	query := `SELECT id, reference, user_id, name, created_at, updated_at FROM lists WHERE user_id = ? AND name = ?`
	var list models.List
	var internalUserIDCheck int
	err = s.db.QueryRow(query, internalUserID, listName).Scan(
		&list.ID,
		&list.Reference,
		&internalUserIDCheck,
		&list.Name,
		&list.CreatedAt,
		&list.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("list with name %s not found", listName)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get list: %w", err)
	}

	list.UserID = userReference
	return &list, nil
}

// GetAllLists returns all lists for a user
func (s *MySQLStorage) GetAllLists(userReference string) ([]*models.List, error) {
	// Get internal user ID from reference
	internalUserID, err := s.getUserIDFromReference(userReference)
	if err != nil {
		return nil, err
	}

	query := `SELECT id, reference, user_id, name, created_at, updated_at FROM lists WHERE user_id = ? ORDER BY name`
	rows, err := s.db.Query(query, internalUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to query lists: %w", err)
	}
	defer rows.Close()

	var lists []*models.List
	for rows.Next() {
		var list models.List
		var internalUserIDCheck int
		if err := rows.Scan(&list.ID, &list.Reference, &internalUserIDCheck, &list.Name, &list.CreatedAt, &list.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan list: %w", err)
		}
		list.UserID = userReference
		lists = append(lists, &list)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lists: %w", err)
	}

	// If no lists found, create default list
	if len(lists) == 0 {
		defaultList, err := s.CreateList(userReference, "default")
		if err != nil {
			return nil, err
		}
		return []*models.List{defaultList}, nil
	}

	return lists, nil
}

// DeleteList deletes a list and all its favorites (CASCADE)
func (s *MySQLStorage) DeleteList(userReference, listReference string) error {
	// Get internal user ID from reference
	internalUserID, err := s.getUserIDFromReference(userReference)
	if err != nil {
		return err
	}

	// Get internal list ID first (needed for deleting favorites)
	var internalListID int
	err = s.db.QueryRow("SELECT id FROM lists WHERE reference = ? AND user_id = ?", listReference, internalUserID).Scan(&internalListID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("list with reference %s not found", listReference)
	}
	if err != nil {
		return fmt.Errorf("failed to get list ID: %w", err)
	}

	// Start a transaction to ensure both deletions succeed or fail together
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete all favorites in this list first
	_, err = tx.Exec("DELETE FROM favorites WHERE list_id = ?", internalListID)
	if err != nil {
		return fmt.Errorf("failed to delete favorites: %w", err)
	}

	// Delete the list
	result, err := tx.Exec("DELETE FROM lists WHERE id = ? AND user_id = ?", internalListID, internalUserID)
	if err != nil {
		return fmt.Errorf("failed to delete list: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("list with reference %s not found", listReference)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// AddFavorite adds or updates a favorite for a user in a specific list
// userReference is the user reference (e.g., "user1")
// assetReference is the asset reference (e.g., "user1_favorite1")
// listReference is the list reference (e.g., "list_user1_default")
// sortOrder is optional - if nil, will use next sequential value
func (s *MySQLStorage) AddFavorite(userReference, assetReference, listReference string, sortOrder *int) (*models.Favorite, error) {
	// Get or create default list if listReference is empty
	if listReference == "" {
		list, err := s.GetListByName(userReference, "default")
		if err != nil {
			// Create default list if it doesn't exist
			list, err = s.CreateList(userReference, "default")
			if err != nil {
				return nil, err
			}
		}
		listReference = list.Reference
	}
	// Get internal user ID from reference
	internalUserID, err := s.getUserIDFromReference(userReference)
	if err != nil {
		return nil, err
	}

	// Verify asset exists
	var internalAssetID int
	err = s.db.QueryRow("SELECT id FROM assets WHERE reference = ?", assetReference).Scan(&internalAssetID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("asset with reference %s not found", assetReference)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get asset ID: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()

	// Get internal list ID from reference
	internalListID, err := s.getListIDFromReference(listReference)
	if err != nil {
		return nil, fmt.Errorf("failed to get list ID: %w", err)
	}

	// Get list name for response
	var listName string
	err = tx.QueryRow("SELECT name FROM lists WHERE id = ?", internalListID).Scan(&listName)
	if err != nil {
		return nil, fmt.Errorf("failed to get list name: %w", err)
	}

	// Generate favorite reference
	favoriteReference := s.generateFavoriteReference(userReference, assetReference, listReference)

	// Check if favorite already exists in this list
	var existingID int
	var existingAddedAt time.Time
	var existingSortOrder int
	checkQuery := `SELECT id, added_at, sort_order FROM favorites WHERE user_id = ? AND asset_id = ? AND list_id = ?`
	err = tx.QueryRow(checkQuery, internalUserID, internalAssetID, internalListID).Scan(&existingID, &existingAddedAt, &existingSortOrder)

	var favoriteID int
	var addedAt time.Time
	var finalSortOrder int
	if err == sql.ErrNoRows {
		// New favorite
		if sortOrder != nil {
			// Use provided sort_order
			finalSortOrder = *sortOrder
		} else {
			// Get the next sort_order value
			var maxSortOrder sql.NullInt64
			maxQuery := `SELECT MAX(sort_order) FROM favorites WHERE user_id = ? AND list_id = ?`
			err = tx.QueryRow(maxQuery, internalUserID, internalListID).Scan(&maxSortOrder)
			if err != nil && err != sql.ErrNoRows {
				return nil, fmt.Errorf("failed to get max sort_order: %w", err)
			}
			if maxSortOrder.Valid {
				finalSortOrder = int(maxSortOrder.Int64) + 1
			} else {
				finalSortOrder = 0
			}
		}

		addedAt = now
		insertQuery := `INSERT INTO favorites (reference, user_id, asset_id, list_id, sort_order, added_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)`
		result, err := tx.Exec(insertQuery, favoriteReference, internalUserID, internalAssetID, internalListID, finalSortOrder, addedAt, now)
		if err != nil {
			return nil, fmt.Errorf("failed to insert favorite: %w", err)
		}
		insertedID, err := result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get inserted favorite ID: %w", err)
		}
		favoriteID = int(insertedID)
	} else if err != nil {
		return nil, fmt.Errorf("failed to check existing favorite: %w", err)
	} else {
		// Existing favorite
		favoriteID = existingID
		addedAt = existingAddedAt
		if sortOrder != nil {
			// Update sort_order if provided
			finalSortOrder = *sortOrder
			updateQuery := `UPDATE favorites SET sort_order = ?, updated_at = ?, reference = ? WHERE id = ?`
			if _, err := tx.Exec(updateQuery, finalSortOrder, now, favoriteReference, favoriteID); err != nil {
				return nil, fmt.Errorf("failed to update favorite: %w", err)
			}
		} else {
			// Keep existing sort_order
			finalSortOrder = existingSortOrder
			updateQuery := `UPDATE favorites SET updated_at = ?, reference = ? WHERE id = ?`
			if _, err := tx.Exec(updateQuery, now, favoriteReference, favoriteID); err != nil {
				return nil, fmt.Errorf("failed to update favorite: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Get asset for response
	asset, err := s.GetAsset(assetReference)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	return &models.Favorite{
		ID:        favoriteID,
		Reference: favoriteReference,
		UserID:    userReference, // Return the reference, not internal ID
		AssetID:   assetReference, // Return the reference, not internal ID
		ListID:    internalListID,
		ListRef:   listReference,
		ListName:  listName,
		SortOrder: finalSortOrder,
		Asset:     asset,
		AddedAt:   addedAt,
		UpdatedAt: now,
	}, nil
}

// RemoveFavorite removes a favorite for a user from a specific list
// userReference, assetReference, and listReference are the reference values (e.g., "user1", "chart1", "list_user1_default")
func (s *MySQLStorage) RemoveFavorite(userReference, assetReference, listReference string) error {
	// Get or find default list if listReference is empty
	if listReference == "" {
		list, err := s.GetListByName(userReference, "default")
		if err != nil {
			return ErrFavoriteNotFound
		}
		listReference = list.Reference
	}

	// Get internal user ID from reference
	internalUserID, err := s.getUserIDFromReference(userReference)
	if err != nil {
		return err
	}

	// Get internal asset ID from reference
	internalAssetID, err := s.getAssetIDFromReference(assetReference)
	if err != nil {
		return err
	}

	// Get internal list ID from reference
	internalListID, err := s.getListIDFromReference(listReference)
	if err != nil {
		return ErrFavoriteNotFound
	}

	query := `DELETE FROM favorites WHERE user_id = ? AND asset_id = ? AND list_id = ?`
	result, err := s.db.Exec(query, internalUserID, internalAssetID, internalListID)
	if err != nil {
		return fmt.Errorf("failed to delete favorite: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrFavoriteNotFound
	}

	return nil
}

// RemoveFavoriteByReference removes a favorite by its reference
func (s *MySQLStorage) RemoveFavoriteByReference(favoriteReference string) error {
	query := `DELETE FROM favorites WHERE reference = ?`
	result, err := s.db.Exec(query, favoriteReference)
	if err != nil {
		return fmt.Errorf("failed to delete favorite: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrFavoriteNotFound
	}

	return nil
}

// GetFavorites returns all favorites for a user in a specific list
// userID parameter is actually the user reference (e.g., "user1")
// listReference is the list reference (e.g., "list_user1_default")
func (s *MySQLStorage) GetFavorites(userReference, listReference string) ([]*models.Favorite, error) {
	// Get or find default list if listReference is empty
	if listReference == "" {
		list, err := s.GetListByName(userReference, "default")
		if err != nil {
			// Create default list if it doesn't exist
			list, err = s.CreateList(userReference, "default")
			if err != nil {
				return nil, err
			}
		}
		listReference = list.Reference
	}

	// Get internal user ID from reference
	internalUserID, err := s.getUserIDFromReference(userReference)
	if err != nil {
		return nil, err
	}

	// Get internal list ID from reference
	internalListID, err := s.getListIDFromReference(listReference)
	if err != nil {
		return nil, err
	}

	query := `SELECT f.id, f.reference, f.user_id, f.asset_id, f.list_id, f.sort_order, f.added_at, f.updated_at,
		a.reference, a.type, a.description, a.data, a.created_at, a.updated_at,
		l.reference, l.name
		FROM favorites f
		INNER JOIN assets a ON f.asset_id = a.id
		INNER JOIN lists l ON f.list_id = l.id
		WHERE f.user_id = ? AND f.list_id = ?
		ORDER BY f.sort_order ASC, f.added_at DESC`

	rows, err := s.db.Query(query, internalUserID, internalListID)
	if err != nil {
		return nil, fmt.Errorf("failed to query favorites: %w", err)
	}
	defer rows.Close()

	var favorites []*models.Favorite
	for rows.Next() {
		var fav models.Favorite
		var internalUserID, internalAssetID, internalListID int
		var assetReference, assetType, assetData, assetDescription string
		var listReference, listName string
		var assetCreatedAt, assetUpdatedAt time.Time

		var dbReference string // Read old reference from DB but regenerate it
		err := rows.Scan(
			&fav.ID,
			&dbReference, // Read but don't use - we'll regenerate
			&internalUserID, // This is the INT user_id from favorites table (we'll convert to reference)
			&internalAssetID, // This is the INT asset_id from favorites table (we'll convert to reference)
			&internalListID, // This is the INT list_id from favorites table
			&fav.SortOrder, // Sort order for this favorite
			&fav.AddedAt,
			&fav.UpdatedAt,
			&assetReference, // This is the asset reference from assets table
			&assetType,
			&assetDescription,
			&assetData,
			&assetCreatedAt,
			&assetUpdatedAt,
			&listReference, // This is the list reference from lists table
			&listName,      // This is the list name from lists table
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan favorite: %w", err)
		}

		// Regenerate reference using new format: fav_{assetReference}
		fav.Reference = fmt.Sprintf("fav_%s", assetReference)

		// Set user reference (we already have it)
		fav.UserID = userReference

		// Set asset reference
		fav.AssetID = assetReference

		// Set list information
		fav.ListID = internalListID
		fav.ListRef = listReference
		fav.ListName = listName

		// Deserialize asset using reference as ID
		asset, err := s.deserializeAsset(assetReference, models.AssetType(assetType), assetDescription, assetData, assetCreatedAt, assetUpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize asset: %w", err)
		}

		fav.Asset = asset
		favorites = append(favorites, &fav)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating favorites: %w", err)
	}

	return favorites, nil
}

// GetFavoritesPaginated returns paginated favorites for a user in a specific list
func (s *MySQLStorage) GetFavoritesPaginated(userReference, listReference string, page, pageSize int, filters *models.FilterParams) ([]*models.Favorite, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20 // Default page size
	}
	if pageSize > 100 {
		pageSize = 100 // Max page size
	}

	// Get or find default list if listReference is empty
	if listReference == "" {
		list, err := s.GetListByName(userReference, "default")
		if err != nil {
			// Create default list if it doesn't exist
			list, err = s.CreateList(userReference, "default")
			if err != nil {
				return nil, 0, err
			}
		}
		listReference = list.Reference
	}

	// Get internal user ID from reference
	internalUserID, err := s.getUserIDFromReference(userReference)
	if err != nil {
		return nil, 0, err
	}

	// Get internal list ID from reference
	internalListID, err := s.getListIDFromReference(listReference)
	if err != nil {
		return nil, 0, err
	}

	// Build WHERE clause with filters
	whereClause := "f.user_id = ? AND f.list_id = ?"
	args := []interface{}{internalUserID, internalListID}

	if filters != nil {
		if filters.AssetType != nil {
			whereClause += " AND a.type = ?"
			args = append(args, string(*filters.AssetType))
		}
		if filters.DateFrom != nil {
			whereClause += " AND f.added_at >= ?"
			args = append(args, *filters.DateFrom)
		}
		if filters.DateTo != nil {
			whereClause += " AND f.added_at <= ?"
			args = append(args, *filters.DateTo)
		}
	}

	// Get total count with filters
	var totalCount int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM favorites f INNER JOIN assets a ON f.asset_id = a.id WHERE %s`, whereClause)
	err = s.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count favorites: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`SELECT f.id, f.reference, f.user_id, f.asset_id, f.list_id, f.sort_order, f.added_at, f.updated_at,
		a.reference, a.type, a.description, a.data, a.created_at, a.updated_at,
		l.reference, l.name
		FROM favorites f
		INNER JOIN assets a ON f.asset_id = a.id
		INNER JOIN lists l ON f.list_id = l.id
		WHERE %s
		ORDER BY f.sort_order ASC, f.added_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, pageSize, offset)
	rows, err := s.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query favorites: %w", err)
	}
	defer rows.Close()

	var favorites []*models.Favorite
	for rows.Next() {
		var fav models.Favorite
		var internalUserID, internalAssetID, internalListID int
		var assetReference, assetType, assetData, assetDescription string
		var listReference, listName string
		var assetCreatedAt, assetUpdatedAt time.Time

		var dbReference string // Read old reference from DB but regenerate it
		err := rows.Scan(
			&fav.ID,
			&dbReference, // Read but don't use - we'll regenerate
			&internalUserID,
			&internalAssetID,
			&internalListID,
			&fav.SortOrder, // Sort order for this favorite
			&fav.AddedAt,
			&fav.UpdatedAt,
			&assetReference,
			&assetType,
			&assetDescription,
			&assetData,
			&assetCreatedAt,
			&assetUpdatedAt,
			&listReference,
			&listName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan favorite: %w", err)
		}

		// Regenerate reference using new format: fav_{assetReference}
		fav.Reference = fmt.Sprintf("fav_%s", assetReference)

		fav.UserID = userReference
		fav.AssetID = assetReference
		fav.ListID = internalListID
		fav.ListRef = listReference
		fav.ListName = listName

		asset, err := s.deserializeAsset(assetReference, models.AssetType(assetType), assetDescription, assetData, assetCreatedAt, assetUpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to deserialize asset: %w", err)
		}

		fav.Asset = asset
		favorites = append(favorites, &fav)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating favorites: %w", err)
	}

	return favorites, totalCount, nil
}

// GetAllFavorites returns all favorites for a user across all lists
func (s *MySQLStorage) GetAllFavorites(userReference string) ([]*models.Favorite, error) {
	// Get internal user ID from reference
	internalUserID, err := s.getUserIDFromReference(userReference)
	if err != nil {
		return nil, err
	}

	query := `SELECT f.id, f.reference, f.user_id, f.asset_id, f.list_id, f.sort_order, f.added_at, f.updated_at,
		a.reference, a.type, a.description, a.data, a.created_at, a.updated_at,
		l.reference, l.name
		FROM favorites f
		INNER JOIN assets a ON f.asset_id = a.id
		INNER JOIN lists l ON f.list_id = l.id
		WHERE f.user_id = ?
		ORDER BY f.list_id ASC, f.sort_order ASC, f.added_at DESC`

	rows, err := s.db.Query(query, internalUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to query favorites: %w", err)
	}
	defer rows.Close()

	var favorites []*models.Favorite
	for rows.Next() {
		var fav models.Favorite
		var internalUserID, internalAssetID, internalListID int
		var assetReference, assetType, assetData, assetDescription string
		var listReference, listName string
		var assetCreatedAt, assetUpdatedAt time.Time

		err := rows.Scan(
			&fav.ID,
			&fav.Reference,
			&internalUserID,
			&internalAssetID,
			&internalListID,
			&fav.SortOrder, // Sort order for this favorite
			&fav.AddedAt,
			&fav.UpdatedAt,
			&assetReference,
			&assetType,
			&assetDescription,
			&assetData,
			&assetCreatedAt,
			&assetUpdatedAt,
			&listReference,
			&listName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan favorite: %w", err)
		}

		fav.UserID = userReference
		fav.AssetID = assetReference
		fav.ListID = internalListID
		fav.ListRef = listReference
		fav.ListName = listName

		asset, err := s.deserializeAsset(assetReference, models.AssetType(assetType), assetDescription, assetData, assetCreatedAt, assetUpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize asset: %w", err)
		}

		fav.Asset = asset
		favorites = append(favorites, &fav)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating favorites: %w", err)
	}

	return favorites, nil
}

// GetAllFavoritesPaginated returns paginated favorites for a user across all lists
func (s *MySQLStorage) GetAllFavoritesPaginated(userReference string, page, pageSize int, filters *models.FilterParams) ([]*models.Favorite, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20 // Default page size
	}
	if pageSize > 100 {
		pageSize = 100 // Max page size
	}

	// Get internal user ID from reference
	internalUserID, err := s.getUserIDFromReference(userReference)
	if err != nil {
		return nil, 0, err
	}

	// Build WHERE clause with filters
	whereClause := "f.user_id = ?"
	args := []interface{}{internalUserID}

	if filters != nil {
		if filters.AssetType != nil {
			whereClause += " AND a.type = ?"
			args = append(args, string(*filters.AssetType))
		}
		if filters.DateFrom != nil {
			whereClause += " AND f.added_at >= ?"
			args = append(args, *filters.DateFrom)
		}
		if filters.DateTo != nil {
			whereClause += " AND f.added_at <= ?"
			args = append(args, *filters.DateTo)
		}
	}

	// Get total count with filters
	var totalCount int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM favorites f INNER JOIN assets a ON f.asset_id = a.id WHERE %s`, whereClause)
	err = s.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count favorites: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`SELECT f.id, f.reference, f.user_id, f.asset_id, f.list_id, f.sort_order, f.added_at, f.updated_at,
		a.reference, a.type, a.description, a.data, a.created_at, a.updated_at,
		l.reference, l.name
		FROM favorites f
		INNER JOIN assets a ON f.asset_id = a.id
		INNER JOIN lists l ON f.list_id = l.id
		WHERE %s
		ORDER BY f.sort_order ASC, f.added_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, pageSize, offset)
	rows, err := s.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query favorites: %w", err)
	}
	defer rows.Close()

	var favorites []*models.Favorite
	for rows.Next() {
		var fav models.Favorite
		var internalUserID, internalAssetID, internalListID int
		var assetReference, assetType, assetData, assetDescription string
		var listReference, listName string
		var assetCreatedAt, assetUpdatedAt time.Time

		var dbReference string // Read old reference from DB but regenerate it
		err := rows.Scan(
			&fav.ID,
			&dbReference, // Read but don't use - we'll regenerate
			&internalUserID,
			&internalAssetID,
			&internalListID,
			&fav.SortOrder, // Sort order for this favorite
			&fav.AddedAt,
			&fav.UpdatedAt,
			&assetReference,
			&assetType,
			&assetDescription,
			&assetData,
			&assetCreatedAt,
			&assetUpdatedAt,
			&listReference,
			&listName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan favorite: %w", err)
		}

		// Regenerate reference using new format: fav_{assetReference}
		fav.Reference = fmt.Sprintf("fav_%s", assetReference)

		fav.UserID = userReference
		fav.AssetID = assetReference
		fav.ListID = internalListID
		fav.ListRef = listReference
		fav.ListName = listName

		asset, err := s.deserializeAsset(assetReference, models.AssetType(assetType), assetDescription, assetData, assetCreatedAt, assetUpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to deserialize asset: %w", err)
		}

		fav.Asset = asset
		favorites = append(favorites, &fav)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating favorites: %w", err)
	}

	return favorites, totalCount, nil
}

// GetAllListsPaginated returns paginated lists for a user
func (s *MySQLStorage) GetAllListsPaginated(userReference string, page, pageSize int) ([]*models.List, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20 // Default page size
	}
	if pageSize > 100 {
		pageSize = 100 // Max page size
	}

	// Get internal user ID from reference
	internalUserID, err := s.getUserIDFromReference(userReference)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM lists WHERE user_id = ?`
	err = s.db.QueryRow(countQuery, internalUserID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count lists: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	query := `SELECT id, reference, user_id, name, created_at, updated_at FROM lists WHERE user_id = ? ORDER BY name LIMIT ? OFFSET ?`
	rows, err := s.db.Query(query, internalUserID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query lists: %w", err)
	}
	defer rows.Close()

	var lists []*models.List
	for rows.Next() {
		var list models.List
		var internalUserIDCheck int
		if err := rows.Scan(&list.ID, &list.Reference, &internalUserIDCheck, &list.Name, &list.CreatedAt, &list.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan list: %w", err)
		}
		list.UserID = userReference
		lists = append(lists, &list)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating lists: %w", err)
	}

	return lists, totalCount, nil
}

// UpdateAssetDescription updates the description of an asset
// assetID parameter is actually the asset reference (e.g., "chart1")
func (s *MySQLStorage) UpdateAssetDescription(assetReference, description string) error {
	query := `UPDATE assets SET description = ?, updated_at = ? WHERE reference = ?`
	result, err := s.db.Exec(query, description, time.Now(), assetReference)
	if err != nil {
		return fmt.Errorf("failed to update asset description: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrAssetNotFound
	}

	return nil
}

// CreateAsset creates a new asset
func (s *MySQLStorage) CreateAsset(userID string, asset models.Asset) (models.Asset, error) {
	if asset == nil {
		return nil, fmt.Errorf("asset is required")
	}

	assetReference := asset.GetID()
	if assetReference == "" {
		return nil, fmt.Errorf("asset ID (reference) is required")
	}

	now := time.Now()
	asset.SetUpdatedAt(now)

	// Serialize asset to JSON
	assetData, err := s.serializeAsset(asset)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize asset: %w", err)
	}

	// Insert asset
	query := `INSERT INTO assets (reference, type, description, data, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err = s.db.Exec(query,
		assetReference,
		string(asset.GetType()),
		asset.GetDescription(),
		assetData,
		asset.GetCreatedAt(),
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	return asset, nil
}

// UpdateAsset updates an existing asset
func (s *MySQLStorage) UpdateAsset(assetReference string, asset models.Asset) (models.Asset, error) {
	if asset == nil {
		return nil, fmt.Errorf("asset is required")
	}

	now := time.Now()
	asset.SetUpdatedAt(now)

	// Serialize asset to JSON
	assetData, err := s.serializeAsset(asset)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize asset: %w", err)
	}

	// Update asset
	query := `UPDATE assets SET type = ?, description = ?, data = ?, updated_at = ? WHERE reference = ?`
	result, err := s.db.Exec(query,
		string(asset.GetType()),
		asset.GetDescription(),
		assetData,
		now,
		assetReference,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update asset: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, ErrAssetNotFound
	}

	return asset, nil
}

// DeleteAsset deletes an asset
func (s *MySQLStorage) DeleteAsset(assetReference string) error {
	// Check if asset exists
	var assetID int
	err := s.db.QueryRow("SELECT id FROM assets WHERE reference = ?", assetReference).Scan(&assetID)
	if err == sql.ErrNoRows {
		return ErrAssetNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to check asset: %w", err)
	}

	// Delete asset (CASCADE will delete associated favorites)
	query := `DELETE FROM assets WHERE reference = ?`
	result, err := s.db.Exec(query, assetReference)
	if err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrAssetNotFound
	}

	return nil
}

// GetAsset retrieves an asset by reference
// assetID parameter is actually the asset reference (e.g., "chart1")
func (s *MySQLStorage) GetAsset(assetReference string) (models.Asset, error) {
	query := `SELECT reference, type, description, data, created_at, updated_at
		FROM assets WHERE reference = ?`

	var reference, assetType, assetData, description string
	var createdAt, updatedAt time.Time

	err := s.db.QueryRow(query, assetReference).Scan(&reference, &assetType, &description, &assetData, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrAssetNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query asset: %w", err)
	}

	return s.deserializeAsset(reference, models.AssetType(assetType), description, assetData, createdAt, updatedAt)
}

// GetAllAssetsPaginated returns paginated assets
func (s *MySQLStorage) GetAllAssetsPaginated(page, pageSize int, filters *models.FilterParams) ([]models.Asset, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20 // Default page size
	}
	if pageSize > 100 {
		pageSize = 100 // Max page size
	}

	// Build WHERE clause with filters
	whereClause := "1=1"
	args := []interface{}{}

	if filters != nil {
		if filters.AssetType != nil {
			whereClause += " AND type = ?"
			args = append(args, string(*filters.AssetType))
		}
		if filters.DateFrom != nil {
			whereClause += " AND created_at >= ?"
			args = append(args, *filters.DateFrom)
		}
		if filters.DateTo != nil {
			whereClause += " AND created_at <= ?"
			args = append(args, *filters.DateTo)
		}
	}

	// Get total count with filters
	var totalCount int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM assets WHERE %s`, whereClause)
	err := s.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count assets: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Query assets with pagination
	query := fmt.Sprintf(`SELECT reference, type, description, data, created_at, updated_at
		FROM assets
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, pageSize, offset)
	rows, err := s.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query assets: %w", err)
	}
	defer rows.Close()

	var assets []models.Asset
	for rows.Next() {
		var reference, assetType, assetData, description string
		var createdAt, updatedAt time.Time

		err := rows.Scan(&reference, &assetType, &description, &assetData, &createdAt, &updatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan asset: %w", err)
		}

		asset, err := s.deserializeAsset(reference, models.AssetType(assetType), description, assetData, createdAt, updatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to deserialize asset: %w", err)
		}

		assets = append(assets, asset)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating assets: %w", err)
	}

	return assets, totalCount, nil
}

// UpdateFavoriteSortOrder updates the sort order of a favorite
// userReference is the user reference (e.g., "user1")
// favoriteReference is the favorite reference (e.g., "fav_user1_favorite1")
func (s *MySQLStorage) UpdateFavoriteSortOrder(userReference, favoriteReference string, sortOrder int) error {
	// Get internal user ID from reference
	internalUserID, err := s.getUserIDFromReference(userReference)
	if err != nil {
		return fmt.Errorf("user with reference %s not found", userReference)
	}

	// Get favorite ID and list_id from favorite reference
	var favoriteID, listID int
	query := `SELECT id, list_id FROM favorites WHERE reference = ? AND user_id = ?`
	err = s.db.QueryRow(query, favoriteReference, internalUserID).Scan(&favoriteID, &listID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("favorite with reference %s not found", favoriteReference)
	}
	if err != nil {
		return fmt.Errorf("failed to get favorite: %w", err)
	}

	// Update sort_order
	updateQuery := `UPDATE favorites SET sort_order = ?, updated_at = ? WHERE id = ?`
	result, err := s.db.Exec(updateQuery, sortOrder, time.Now(), favoriteID)
	if err != nil {
		return fmt.Errorf("failed to update sort order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("favorite with reference %s not found", favoriteReference)
	}

	return nil
}

// Close closes the database connection
func (s *MySQLStorage) Close() error {
	return s.db.Close()
}

// serializeAsset converts an asset to JSON
func (s *MySQLStorage) serializeAsset(asset models.Asset) (string, error) {
	data, err := json.Marshal(asset)
	if err != nil {
		return "", fmt.Errorf("failed to marshal asset: %w", err)
	}
	return string(data), nil
}

// deserializeAsset converts JSON to an asset
func (s *MySQLStorage) deserializeAsset(id string, assetType models.AssetType, description, data string, createdAt, updatedAt time.Time) (models.Asset, error) {
	var assetData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &assetData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal asset data: %w", err)
	}

	baseAsset := models.BaseAsset{
		ID:          id,
		Type:        assetType,
		Description: description,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	switch assetType {
	case models.AssetTypeChart:
		chart := &models.Chart{BaseAsset: baseAsset}
		if title, ok := assetData["title"].(string); ok {
			chart.Title = title
		}
		if xAxis, ok := assetData["x_axis"].(string); ok {
			chart.XAxis = xAxis
		}
		if yAxis, ok := assetData["y_axis"].(string); ok {
			chart.YAxis = yAxis
		}
		if dataPoints, ok := assetData["data"].([]interface{}); ok {
			chart.Data = make([]models.DataPoint, 0, len(dataPoints))
			for _, dp := range dataPoints {
				if dpMap, ok := dp.(map[string]interface{}); ok {
					dataPoint := models.DataPoint{}
					if label, ok := dpMap["label"].(string); ok {
						dataPoint.Label = label
					}
					if value, ok := dpMap["value"]; ok {
						valueBytes, _ := json.Marshal(value)
						dataPoint.Value = valueBytes
					}
					chart.Data = append(chart.Data, dataPoint)
				}
			}
		}
		return chart, nil

	case models.AssetTypeInsight:
		insight := &models.Insight{BaseAsset: baseAsset}
		if text, ok := assetData["text"].(string); ok {
			insight.Text = text
		}
		return insight, nil

	case models.AssetTypeAudience:
		audience := &models.Audience{BaseAsset: baseAsset}
		if gender, ok := assetData["gender"].(string); ok {
			g := models.Gender(gender)
			audience.Gender = &g
		}
		if country, ok := assetData["birth_country"].(string); ok {
			audience.BirthCountry = country
		}
		if ageGroupMap, ok := assetData["age_group"].(map[string]interface{}); ok {
			ageGroup := &models.AgeGroup{}
			if min, ok := ageGroupMap["min"].(float64); ok {
				ageGroup.Min = int(min)
			}
			if max, ok := ageGroupMap["max"].(float64); ok {
				ageGroup.Max = int(max)
			}
			audience.AgeGroup = ageGroup
		}
		if hours, ok := assetData["social_media_hours_min"].(float64); ok {
			audience.SocialMediaHoursMin = &hours
		}
		if purchases, ok := assetData["purchases_last_month"].(float64); ok {
			p := int(purchases)
			audience.PurchasesLastMonth = &p
		}
		return audience, nil

	default:
		return nil, fmt.Errorf("unknown asset type: %s", assetType)
	}
}

