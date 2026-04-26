package aifilter

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/XDwanj/gx/internal/index"

	"github.com/jmoiron/sqlx"

	// Register the SQLite driver used by the AI selection cache.
	_ "modernc.org/sqlite"
)

const sqliteBusyTimeoutMillis = 5_000

type Cache struct {
	db *sqlx.DB
}

type cacheRow struct {
	SelectedRaw []byte `db:"selected_json"`
}

func OpenCache(root string) (*Cache, error) {
	path := index.CachePathFor(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := sqlx.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.Exec(fmt.Sprintf("PRAGMA busy_timeout = %d", sqliteBusyTimeoutMillis)); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS ai_cache (
			key TEXT PRIMARY KEY,
			created_at INTEGER NOT NULL,
			last_used_at INTEGER NOT NULL,
			prompt_version INTEGER NOT NULL,
			provider_hash TEXT NOT NULL,
			selected_json BLOB NOT NULL
		)
	`); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Cache{db: db}, nil
}

func (cache *Cache) Close() error {
	return cache.db.Close()
}

func (cache *Cache) Get(key string) ([]string, bool, error) {
	var row cacheRow
	if err := cache.db.Get(&row, `
		SELECT selected_json
		FROM ai_cache
		WHERE key = ?
	`, key); err != nil {
		if isNoRows(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	var selected []string
	if err := json.Unmarshal(row.SelectedRaw, &selected); err != nil {
		return nil, false, err
	}
	if _, err := cache.db.Exec(`
		UPDATE ai_cache
		SET last_used_at = ?
		WHERE key = ?
	`, time.Now().Unix(), key); err != nil {
		return nil, false, err
	}
	return selected, true, nil
}

func (cache *Cache) Put(key string, providerID string, selected []string) error {
	encoded, err := json.Marshal(selected)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	_, err = cache.db.Exec(`
		INSERT INTO ai_cache (key, created_at, last_used_at, prompt_version, provider_hash, selected_json)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			last_used_at = excluded.last_used_at,
			prompt_version = excluded.prompt_version,
			provider_hash = excluded.provider_hash,
			selected_json = excluded.selected_json
	`, key, now, now, promptVersion, providerID, encoded)
	return err
}

func isNoRows(err error) bool {
	return err == sql.ErrNoRows
}
