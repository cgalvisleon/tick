package tick

import (
	"fmt"

	"github.com/cgalvisleon/tick/internal/findroot"
	"github.com/cgalvisleon/tick/internal/store"
	"github.com/spf13/cobra"
)

var Root = &cobra.Command{
	Use:   "tick",
	Short: "Seguimiento simple de proyectos y tareas, con ergonomia tipo git",
}

/**
* openStore: Locates the nearest .tick directory and opens its database.
* Used by every command except init, which creates the project instead of
* locating one.
* @return *store.DB, error
**/
func openStore() (*store.DB, error) {
	root, err := findroot.Find()
	if err != nil {
		return nil, err
	}
	db, err := store.Open(findroot.DotDir(root))
	if err != nil {
		return nil, fmt.Errorf("opening project database: %w", err)
	}
	return db, nil
}
