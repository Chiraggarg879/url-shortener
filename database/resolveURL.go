package database

import "database/sql"

// ResolveURL returns the destination URL for an active short code.
func ResolveURL(db *sql.DB, shortCode string) (string, error) {
	var originalURL string

	err := db.QueryRow(`
		SELECT original_url
		FROM url_mappings
		WHERE short_code = ?
		  AND (expires_at IS NULL OR expires_at > NOW())
	`, shortCode).Scan(&originalURL)
	if err != nil {
		return "", err
	}

	_, err = db.Exec(`
		UPDATE url_mappings
		SET click_count = click_count + 1
		WHERE short_code = ?
	`, shortCode)
	if err != nil {
		return "", err
	}

	return originalURL, nil
}
