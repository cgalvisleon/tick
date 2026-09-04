package tick

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull [remote]",
	Short: "Copia la base de datos del remote (por defecto 'origin') hacia el proyecto local. Sobreescribe lo local por completo.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openStore()
		if err != nil {
			return err
		}

		name := ""
		if len(args) == 1 {
			name = args[0]
		}
		remote, err := resolveRemote(db, name)
		if err != nil {
			return err
		}

		src := filepath.Join(remote.Path, "tick.db")
		if _, err := os.Stat(src); err != nil {
			return fmt.Errorf("no se encontro %s: %w", src, err)
		}

		removeWALSidecars(db.Path)
		if err := copyFile(src, db.Path); err != nil {
			return err
		}

		fmt.Printf("pull de %s (%s) completo\n", remote.Name, src)
		return nil
	},
}

func init() {
	Root.AddCommand(pullCmd)
}
