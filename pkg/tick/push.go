package tick

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [remote]",
	Short: "Copia la base de datos local hacia el remote (por defecto 'origin'). Sobreescribe el remote por completo.",
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

		if err := db.Checkpoint(); err != nil {
			return err
		}

		dst := filepath.Join(remote.Path, "tick.db")
		removeWALSidecars(dst)
		if err := copyFile(db.Path, dst); err != nil {
			return err
		}

		fmt.Printf("push a %s (%s) completo\n", remote.Name, dst)
		return nil
	},
}

func init() {
	Root.AddCommand(pushCmd)
}
