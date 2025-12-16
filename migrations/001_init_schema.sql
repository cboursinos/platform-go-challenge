-- Database schema for Favorites API

-- Create database if it doesn't exist
CREATE DATABASE IF NOT EXISTS favorites_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE favorites_db;

-- ============================================
-- TABLES
-- ============================================

-- Users table with basic information
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    reference VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_reference (reference),
    INDEX idx_email (email),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Assets table with optimized indexes
CREATE TABLE IF NOT EXISTS assets (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Lists table for organizing favorites
-- Each user can have multiple lists (e.g., "default", "work", "personal")
CREATE TABLE IF NOT EXISTS lists (
    id INT AUTO_INCREMENT PRIMARY KEY,
    reference VARCHAR(255) NOT NULL UNIQUE,
    user_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_reference (reference),
    INDEX idx_user_id (user_id),
    INDEX idx_user_name (user_id, name),
    UNIQUE KEY uk_user_list_name (user_id, name),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Favorites table with composite indexes for optimized queries
-- Links users, assets, and lists together
CREATE TABLE IF NOT EXISTS favorites (
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
    UNIQUE KEY uk_user_asset_list (user_id, asset_id, list_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE,
    FOREIGN KEY (list_id) REFERENCES lists(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- DEMO DATA
-- ============================================

-- Insert demo users
INSERT IGNORE INTO users (reference, email, name, created_at, updated_at) VALUES
('user1', 'alice.johnson@example.com', 'Alice Johnson', NOW(), NOW()),
('user2', 'bob.smith@example.com', 'Bob Smith', NOW(), NOW()),
('user3', 'charlie.brown@example.com', 'Charlie Brown', NOW(), NOW()),
('user4', 'diana.prince@example.com', 'Diana Prince', NOW(), NOW()),
('user5', 'eve.wilson@example.com', 'Eve Wilson', NOW(), NOW());

-- Insert demo Chart assets
INSERT IGNORE INTO assets (reference, type, description, data, created_at, updated_at) VALUES
('chart1', 'chart', 'Monthly Sales Revenue Overview', 
 '{"title":"Monthly Sales Revenue","x_axis":"Month","y_axis":"Revenue (USD)","data":[{"label":"Jan","value":125000},{"label":"Feb","value":145000},{"label":"Mar","value":138000},{"label":"Apr","value":162000},{"label":"May","value":178000},{"label":"Jun","value":195000}]}',
 NOW(), NOW()),
('chart2', 'chart', 'Social Media Usage by Platform',
 '{"title":"Social Media Usage by Platform","x_axis":"Platform","y_axis":"Daily Active Users (Millions)","data":[{"label":"Facebook","value":2900},{"label":"Instagram","value":2000},{"label":"Twitter","value":450},{"label":"TikTok","value":1000},{"label":"LinkedIn","value":310}]}',
 NOW(), NOW()),
('chart3', 'chart', 'E-commerce Conversion Rates',
 '{"title":"E-commerce Conversion Rates","x_axis":"Device Type","y_axis":"Conversion Rate (%)","data":[{"label":"Desktop","value":3.2},{"label":"Mobile","value":2.1},{"label":"Tablet","value":2.8}]}',
 NOW(), NOW()),
('chart4', 'chart', 'Customer Age Distribution',
 '{"title":"Customer Age Distribution","x_axis":"Age Group","y_axis":"Number of Customers","data":[{"label":"18-24","value":12500},{"label":"25-34","value":34200},{"label":"35-44","value":28900},{"label":"45-54","value":15600},{"label":"55+","value":8900}]}',
 NOW(), NOW());

-- Insert demo Insight assets
INSERT IGNORE INTO assets (reference, type, description, data, created_at, updated_at) VALUES
('insight1', 'insight', 'Social Media Usage Statistics',
 '{"text":"40% of millennials spend more than 3 hours on social media daily"}',
 NOW(), NOW()),
('insight2', 'insight', 'E-commerce Trends',
 '{"text":"Mobile commerce accounts for 72% of all e-commerce transactions in 2024"}',
 NOW(), NOW()),
('insight3', 'insight', 'Remote Work Impact',
 '{"text":"65% of employees prefer hybrid work models combining remote and office work"}',
 NOW(), NOW()),
('insight4', 'insight', 'Consumer Behavior',
 '{"text":"78% of consumers research products online before making in-store purchases"}',
 NOW(), NOW()),
('insight5', 'insight', 'Digital Marketing',
 '{"text":"Video content generates 3x more engagement than text-based content on social platforms"}',
 NOW(), NOW()),
('insight6', 'insight', 'Technology Adoption',
 '{"text":"92% of Gen Z consumers use mobile payment apps for everyday transactions"}',
 NOW(), NOW());

-- Insert demo Audience assets
INSERT IGNORE INTO assets (reference, type, description, data, created_at, updated_at) VALUES
('audience1', 'audience', 'Tech-Savvy Millennials',
 '{"gender":"male","birth_country":"USA","age_group":{"min":25,"max":35},"social_media_hours_min":3.5,"purchases_last_month":8}',
 NOW(), NOW()),
('audience2', 'audience', 'Social Media Enthusiasts',
 '{"gender":"female","birth_country":"UK","age_group":{"min":18,"max":30},"social_media_hours_min":4.0,"purchases_last_month":12}',
 NOW(), NOW()),
('audience3', 'audience', 'Professional Networkers',
 '{"gender":null,"birth_country":"USA","age_group":{"min":30,"max":45},"social_media_hours_min":2.0,"purchases_last_month":5}',
 NOW(), NOW()),
('audience4', 'audience', 'Young Digital Natives',
 '{"gender":"male","birth_country":"Canada","age_group":{"min":18,"max":25},"social_media_hours_min":5.0,"purchases_last_month":15}',
 NOW(), NOW()),
('audience5', 'audience', 'International Shoppers',
 '{"gender":"female","birth_country":"Germany","age_group":{"min":25,"max":40},"social_media_hours_min":2.5,"purchases_last_month":6}',
 NOW(), NOW()),
('audience6', 'audience', 'Mature Consumers',
 '{"gender":null,"birth_country":"USA","age_group":{"min":45,"max":60},"social_media_hours_min":1.0,"purchases_last_month":3}',
 NOW(), NOW());

-- Insert demo Lists
-- Each user gets a "default" list, user1 also gets a "work" list
INSERT IGNORE INTO lists (reference, user_id, name, created_at, updated_at) VALUES
-- User 1 lists
('list_user1_default', (SELECT id FROM users WHERE reference = 'user1'), 'default', NOW(), NOW()),
('list_user1_work', (SELECT id FROM users WHERE reference = 'user1'), 'work', NOW(), NOW()),
-- User 2 lists
('list_user2_default', (SELECT id FROM users WHERE reference = 'user2'), 'default', NOW(), NOW()),
-- User 3 lists
('list_user3_default', (SELECT id FROM users WHERE reference = 'user3'), 'default', NOW(), NOW()),
-- User 4 lists
('list_user4_default', (SELECT id FROM users WHERE reference = 'user4'), 'default', NOW(), NOW()),
-- User 5 lists
('list_user5_default', (SELECT id FROM users WHERE reference = 'user5'), 'default', NOW(), NOW());

-- Insert demo Favorites (linking users to assets via lists)
-- Note: All IDs reference auto-increment INT ids, not the references
-- We use subqueries to get the id from the reference
-- sort_order is set sequentially (0, 1, 2, ...) for each list
INSERT IGNORE INTO favorites (reference, user_id, asset_id, list_id, sort_order, added_at, updated_at) VALUES
-- User 1 favorites - default list
('fav_user1_chart1_default', (SELECT id FROM users WHERE reference = 'user1'), (SELECT id FROM assets WHERE reference = 'chart1'), (SELECT id FROM lists WHERE reference = 'list_user1_default'), 0, NOW(), NOW()),
('fav_user1_insight1_default', (SELECT id FROM users WHERE reference = 'user1'), (SELECT id FROM assets WHERE reference = 'insight1'), (SELECT id FROM lists WHERE reference = 'list_user1_default'), 1, NOW(), NOW()),
('fav_user1_audience1_default', (SELECT id FROM users WHERE reference = 'user1'), (SELECT id FROM assets WHERE reference = 'audience1'), (SELECT id FROM lists WHERE reference = 'list_user1_default'), 2, NOW(), NOW()),
('fav_user1_chart2_default', (SELECT id FROM users WHERE reference = 'user1'), (SELECT id FROM assets WHERE reference = 'chart2'), (SELECT id FROM lists WHERE reference = 'list_user1_default'), 3, NOW(), NOW()),
-- User 1 favorites - work list
('fav_user1_chart3_work', (SELECT id FROM users WHERE reference = 'user1'), (SELECT id FROM assets WHERE reference = 'chart3'), (SELECT id FROM lists WHERE reference = 'list_user1_work'), 0, NOW(), NOW()),
-- User 2 favorites - default list
('fav_user2_chart3_default', (SELECT id FROM users WHERE reference = 'user2'), (SELECT id FROM assets WHERE reference = 'chart3'), (SELECT id FROM lists WHERE reference = 'list_user2_default'), 0, NOW(), NOW()),
('fav_user2_insight2_default', (SELECT id FROM users WHERE reference = 'user2'), (SELECT id FROM assets WHERE reference = 'insight2'), (SELECT id FROM lists WHERE reference = 'list_user2_default'), 1, NOW(), NOW()),
('fav_user2_audience2_default', (SELECT id FROM users WHERE reference = 'user2'), (SELECT id FROM assets WHERE reference = 'audience2'), (SELECT id FROM lists WHERE reference = 'list_user2_default'), 2, NOW(), NOW()),
('fav_user2_insight3_default', (SELECT id FROM users WHERE reference = 'user2'), (SELECT id FROM assets WHERE reference = 'insight3'), (SELECT id FROM lists WHERE reference = 'list_user2_default'), 3, NOW(), NOW()),
-- User 3 favorites - default list
('fav_user3_chart4_default', (SELECT id FROM users WHERE reference = 'user3'), (SELECT id FROM assets WHERE reference = 'chart4'), (SELECT id FROM lists WHERE reference = 'list_user3_default'), 0, NOW(), NOW()),
('fav_user3_insight4_default', (SELECT id FROM users WHERE reference = 'user3'), (SELECT id FROM assets WHERE reference = 'insight4'), (SELECT id FROM lists WHERE reference = 'list_user3_default'), 1, NOW(), NOW()),
('fav_user3_audience3_default', (SELECT id FROM users WHERE reference = 'user3'), (SELECT id FROM assets WHERE reference = 'audience3'), (SELECT id FROM lists WHERE reference = 'list_user3_default'), 2, NOW(), NOW()),
('fav_user3_chart1_default', (SELECT id FROM users WHERE reference = 'user3'), (SELECT id FROM assets WHERE reference = 'chart1'), (SELECT id FROM lists WHERE reference = 'list_user3_default'), 3, NOW(), NOW()),
('fav_user3_insight5_default', (SELECT id FROM users WHERE reference = 'user3'), (SELECT id FROM assets WHERE reference = 'insight5'), (SELECT id FROM lists WHERE reference = 'list_user3_default'), 4, NOW(), NOW()),
-- User 4 favorites - default list
('fav_user4_audience4_default', (SELECT id FROM users WHERE reference = 'user4'), (SELECT id FROM assets WHERE reference = 'audience4'), (SELECT id FROM lists WHERE reference = 'list_user4_default'), 0, NOW(), NOW()),
('fav_user4_insight6_default', (SELECT id FROM users WHERE reference = 'user4'), (SELECT id FROM assets WHERE reference = 'insight6'), (SELECT id FROM lists WHERE reference = 'list_user4_default'), 1, NOW(), NOW()),
('fav_user4_chart2_default', (SELECT id FROM users WHERE reference = 'user4'), (SELECT id FROM assets WHERE reference = 'chart2'), (SELECT id FROM lists WHERE reference = 'list_user4_default'), 2, NOW(), NOW()),
('fav_user4_audience5_default', (SELECT id FROM users WHERE reference = 'user4'), (SELECT id FROM assets WHERE reference = 'audience5'), (SELECT id FROM lists WHERE reference = 'list_user4_default'), 3, NOW(), NOW()),
-- User 5 favorites - default list
('fav_user5_insight1_default', (SELECT id FROM users WHERE reference = 'user5'), (SELECT id FROM assets WHERE reference = 'insight1'), (SELECT id FROM lists WHERE reference = 'list_user5_default'), 0, NOW(), NOW()),
('fav_user5_chart3_default', (SELECT id FROM users WHERE reference = 'user5'), (SELECT id FROM assets WHERE reference = 'chart3'), (SELECT id FROM lists WHERE reference = 'list_user5_default'), 1, NOW(), NOW()),
('fav_user5_audience6_default', (SELECT id FROM users WHERE reference = 'user5'), (SELECT id FROM assets WHERE reference = 'audience6'), (SELECT id FROM lists WHERE reference = 'list_user5_default'), 2, NOW(), NOW());

-- ============================================
-- INDEX OPTIMIZATION NOTES
-- ============================================

-- Users table:
-- 1. PRIMARY KEY (id): INT AUTO_INCREMENT ensures uniqueness and provides fast lookups
-- 2. idx_reference: Optimizes queries filtering by reference (used in API)
-- 3. idx_email: Optimizes queries filtering by email (unique constraint)
-- 4. idx_created_at: Optimizes queries filtering by creation time
--
-- Assets table:
-- 5. PRIMARY KEY (id): INT AUTO_INCREMENT ensures uniqueness and provides fast lookups
-- 6. idx_reference: Optimizes queries filtering by reference (used in API)
-- 7. idx_type: Optimizes filtering assets by type
-- 8. idx_updated_at: Optimizes queries filtering by update time
-- 9. idx_created_at: Optimizes queries filtering by creation time
--
-- Lists table:
-- 10. PRIMARY KEY (id): INT AUTO_INCREMENT ensures uniqueness and provides fast lookups
-- 11. idx_reference: Optimizes queries filtering by reference (used in API)
-- 12. idx_user_id: Optimizes queries filtering by user_id (INT foreign key)
-- 13. idx_user_name: Composite index for user + name queries
-- 14. uk_user_list_name: Unique constraint ensures one list per name per user
--
-- Favorites table:
-- 15. PRIMARY KEY (id): INT AUTO_INCREMENT ensures uniqueness and provides fast lookups
-- 16. idx_reference: Optimizes queries filtering by reference (used in API)
-- 17. idx_user_id: Optimizes queries filtering by user_id (INT foreign key)
-- 18. idx_asset_id: Optimizes queries filtering by asset_id (INT foreign key)
-- 19. idx_list_id: Optimizes queries filtering by list_id (INT foreign key)
-- 20. idx_user_list: Composite index for user + list queries
-- 21. idx_user_list_sort_order: Optimizes sorting by sort_order for a specific user and list
-- 22. idx_user_list_added_at: Optimizes sorting by added_at for a specific user and list
-- 23. idx_user_list_updated_at: Optimizes sorting by updated_at for a specific user and list
-- 24. uk_user_asset_list: Unique constraint ensures one asset per user per list
--     Note: All foreign keys are INT (efficient), not VARCHAR
--     Note: sort_order allows custom ordering of favorites within each list (0-based, lower numbers appear first)
--
-- API Usage:
-- The API uses "reference" (VARCHAR) in URLs for users, assets, lists, and favorites:
--   - Users: /users/user1/lists (user1 is the reference)
--   - Lists: /users/user1/lists/list_user1_default/favorites (list_user1_default is the list reference)
--   - Assets: /users/user1/lists/list_user1_default/favorites/chart1 (chart1 is the asset reference)
--   - Favorites: /users/user1/favorites/fav_user1_chart1_default (favorite reference)
-- Internally, the system maps references to INT IDs for database operations
-- This provides stable API identifiers while using efficient INT foreign keys
-- Multiple lists per user are supported via the lists table (each list has a unique reference)
