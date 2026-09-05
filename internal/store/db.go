package store

import (
	"fmt"
	"path/filepath"

	_ "github.com/cgalvisleon/et/jsql/drivers/sqlite"

	"github.com/cgalvisleon/et/jsql"
)

const schema = "main"

// DB wraps the jsql handle plus every model used by tick.
type DB struct {
	db      *jsql.DB
	Path    string
	Project *Project
	Task    *Task
	Remote  *Remote
	Config  *Config
}

/**
* Open: Connects to the SQLite database at dir/tick.db, creating the schema
* on first use, and returns a DB with all models defined and initialised.
* @param dir string
* @return *DB, error
**/
func Open(dir string) (*DB, error) {
	path := filepath.Join(dir, "tick.db")

	jdb, err := jsql.ConnectTo("tenant:tick", "local", jsql.DriverSqlite, path, false)
	if err != nil {
		return nil, fmt.Errorf("connecting to %s: %w", path, err)
	}

	result := &DB{db: jdb, Path: path}

	if result.Project, err = defineProject(jdb); err != nil {
		return nil, err
	}
	if result.Task, err = defineTask(jdb); err != nil {
		return nil, err
	}
	if result.Remote, err = defineRemote(jdb); err != nil {
		return nil, err
	}
	if result.Config, err = defineConfig(jdb); err != nil {
		return nil, err
	}

	return result, nil
}

/**
* Checkpoint: Flushes the WAL journal into the main database file so a plain
* file copy (used by push/pull) captures every committed write.
* @return error
**/
func (s *DB) Checkpoint() error {
	_, err := s.db.SqlTx(nil, "PRAGMA wal_checkpoint(TRUNCATE);")
	return err
}
