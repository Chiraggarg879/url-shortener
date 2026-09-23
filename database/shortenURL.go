package database

import (
	"database/sql"
	"time"
)

// CreateURL stores a shortened URL and its expiration time.
func CreateURL(db *sql.DB, shortCode, originalURL string, expiresAt time.Time) error {
	_, err := db.Exec(`
		INSERT INTO url_mappings (short_code, original_url, expires_at, click_count)
		VALUES (?, ?, ?, 0)
	`, shortCode, originalURL, expiresAt)

	return err
}
