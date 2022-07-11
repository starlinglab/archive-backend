package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/starlinglab/archive-backend/providers"
	"github.com/starlinglab/archive-backend/types"
)

var db *sql.DB

var ErrNotFound = errors.New("not found in database")

// queueMut protects queue operations
var queueMut sync.Mutex

type QueueItem struct {
	RowID          int64
	FileID         string
	StorageRequest *types.StorageRequest
	Provider       string
	Status         types.UploadState
}

func Init() error {
	var err error
	db, err = sql.Open("sqlite3", filepath.Join(os.Getenv("AB_DATA_DIR"), "data.db"))
	if err != nil {
		return err
	}

	// Create column for each provider
	q := `CREATE TABLE IF NOT EXISTS files
    (
        file_id         TEXT PRIMARY KEY,
		storage_request TEXT NOT NULL,
	`
	for i, prov := range providers.Providers {
		if i+1 == len(providers.Providers) {
			// Last one
			q += fmt.Sprintf("%s TEXT\n)", prov.Name())
		} else {
			q += fmt.Sprintf("%s TEXT,\n", prov.Name())
		}
	}
	_, err = db.Exec(q)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS queue
	(
		rowid           INTEGER PRIMARY KEY
		file_id         TEXT    NOT NULL,
		provider        TEXT    NOT NULL,
		status          INTEGER NOT NULL,
		taken           BOOLEAN NOT NULL
	)
	`)
	if err != nil {
		return err
	}

	// No threads are running yet, set all taken vars to false
	// They might have been set to true before due to an unexpected shutdown
	_, err = db.Exec(`
	UPDATE queue
	SET taken = 0
	WHERE taken = 1
	`)
	if err != nil {
		return err
	}

	return nil
}

// NextInQueue returns the next item in the database queue.
// It assumes the caller is going to operate on this item and eventually remove
// it from the queue, so it sets 'taken' to true.
//
// 'status' is not set to in-progress, the caller must do that.
//
// If ErrNotFound is returned that means the queue is empty.
func NextInQueue() (*QueueItem, error) {
	queueMut.Lock()
	defer queueMut.Unlock()

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// First read queue item

	var (
		qi QueueItem
		sr []byte // storage request JSON
	)

	row := tx.QueryRow(`
	SELECT rowid, file_id, provider, status
	FROM queue
	WHERE taken = 0
	LIMIT 1
	`)
	err = row.Scan(&qi.RowID, &qi.FileID, &qi.Provider, &qi.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	row = tx.QueryRow(`SELECT storage_request FROM files WHERE file_id = ?`,
		qi.FileID)
	err = row.Scan(&sr)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(sr, &qi.StorageRequest); err != nil {
		return nil, err
	}

	// Set queue item as taken so no one else retrieves it
	_, err = tx.Exec(`UPDATE queue SET taken = 1 WHERE rowid = ?`, qi.RowID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &qi, nil
}

// AddToQueue takes information about a file and creates queue items for each
// provider.
func AddToQueue(fileID string, providers []string, sr *types.StorageRequest) error {
	queueMut.Lock()
	defer queueMut.Unlock()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	srJSON, err := json.Marshal(sr)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO files (file_id, storage_request) VALUES (?,?)`,
		fileID, srJSON)
	if err != nil {
		return err
	}

	for _, provider := range providers {
		_, err = tx.Exec(`
		INSERT INTO queue (file_id, provider, status, taken)
		VALUES (?,?,?,0)
		`, fileID, provider, types.Pending)
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
