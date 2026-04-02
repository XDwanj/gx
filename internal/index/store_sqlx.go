package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
)

type storeMetaRow struct {
	Version int    `db:"version"`
	Root    string `db:"root"`
}

type fileEntryRow struct {
	Path       string `db:"path"`
	MTimeSecs  int64  `db:"mtime_secs"`
	MTimeNanos int64  `db:"mtime_nanos"`
	Language   string `db:"language"`
	SymbolsRaw []byte `db:"symbols_json"`
}

func loadEntries(path string) (map[string]FileData, error) {
	if !fileExists(path) {
		return nil, os.ErrNotExist
	}

	db, err := openStore(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	version, _, err := loadMeta(db)
	if err != nil {
		return nil, err
	}
	if version != IndexVersion {
		return nil, fmt.Errorf("version mismatch")
	}

	var rows []fileEntryRow
	if err := db.Select(&rows, `
		SELECT path, mtime_secs, mtime_nanos, language, symbols_json
		FROM file_entries
	`); err != nil {
		return nil, err
	}

	entries := make(map[string]FileData, len(rows))
	for _, row := range rows {
		symbols, err := decodeSymbols(row.SymbolsRaw)
		if err != nil {
			return nil, err
		}
		entries[row.Path] = FileData{
			Meta: FileEntry{
				MTimeSecs:  row.MTimeSecs,
				MTimeNanos: row.MTimeNanos,
				Language:   row.Language,
			},
			Symbols: symbols,
		}
	}
	return entries, nil
}

func saveStore(root string, entries map[string]FileData) error {
	path := CachePathFor(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	db, err := openStore(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, execErr := tx.Exec(`DELETE FROM store_meta`); execErr != nil {
		return execErr
	}
	if _, execErr := tx.Exec(
		`INSERT INTO store_meta (id, version, root) VALUES (1, ?, ?)`,
		IndexVersion,
		root,
	); execErr != nil {
		return execErr
	}
	if _, execErr := tx.Exec(`DELETE FROM file_entries`); execErr != nil {
		return execErr
	}

	statement, err := tx.Preparex(`
		INSERT INTO file_entries (path, mtime_secs, mtime_nanos, language, symbols_json)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer func() {
		_ = statement.Close()
	}()

	for path, data := range entries {
		symbolsJSON, err := encodeSymbols(data.Symbols)
		if err != nil {
			return err
		}
		if _, err := statement.Exec(
			path,
			data.Meta.MTimeSecs,
			data.Meta.MTimeNanos,
			data.Meta.Language,
			symbolsJSON,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func openStore(path string) (*sqlx.DB, error) {
	db, err := sqlx.Open(sqliteDriverName, path)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.Exec(fmt.Sprintf("PRAGMA busy_timeout = %d", sqliteBusyTimeoutMillis)); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := initStore(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func initStore(db *sqlx.DB) error {
	for _, statement := range storeSchemaStatements() {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

func loadMeta(db *sqlx.DB) (int, string, error) {
	var row storeMetaRow
	if err := db.Get(&row, `
		SELECT version, root
		FROM store_meta
		WHERE id = 1
	`); err != nil {
		return 0, "", err
	}
	return row.Version, row.Root, nil
}

func storeSchemaStatements() []string {
	return []string{
		`
			CREATE TABLE IF NOT EXISTS store_meta (
				id INTEGER PRIMARY KEY CHECK (id = 1),
				version INTEGER NOT NULL,
				root TEXT NOT NULL
			)
		`,
		`
			CREATE TABLE IF NOT EXISTS file_entries (
				path TEXT PRIMARY KEY,
				mtime_secs INTEGER NOT NULL,
				mtime_nanos INTEGER NOT NULL,
				language TEXT NOT NULL,
				symbols_json BLOB NOT NULL
			)
		`,
	}
}

func encodeSymbols(symbols []Symbol) ([]byte, error) {
	symbolsJSON, err := json.Marshal(symbols)
	if err != nil {
		return nil, fmt.Errorf("gx: failed to encode index: %w", err)
	}
	return symbolsJSON, nil
}

func decodeSymbols(symbolsRaw []byte) ([]Symbol, error) {
	var symbols []Symbol
	if err := json.Unmarshal(symbolsRaw, &symbols); err != nil {
		return nil, err
	}
	return symbols, nil
}
