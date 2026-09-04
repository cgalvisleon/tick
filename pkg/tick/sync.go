package tick

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cgalvisleon/tick/internal/store"
)

// resolveRemote picks the named remote, or "origin" when name is empty.
func resolveRemote(db *store.DB, name string) (store.RemoteInfo, error) {
	if name == "" {
		name = "origin"
	}
	info, exists, err := db.Remote.Get(name)
	if err != nil {
		return store.RemoteInfo{}, err
	}
	if !exists {
		return store.RemoteInfo{}, fmt.Errorf("remote %q no existe; usa 'tick remote add %s <path>'", name, name)
	}
	return info, nil
}

// removeWALSidecars deletes the -wal/-shm files next to a SQLite database path,
// if present. Called before overwriting that path with a plain file copy, so a
// stale WAL from a previous session (which checksums against the old main file)
// can't make the freshly-copied database look corrupt on next open.
func removeWALSidecars(dbPath string) {
	os.Remove(dbPath + "-wal")
	os.Remove(dbPath + "-shm")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
