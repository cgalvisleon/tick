package tick

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cgalvisleon/tick/internal/findroot"
	"github.com/cgalvisleon/tick/internal/store"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicializa un proyecto tick en el directorio actual (.tick/)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		dotDir := filepath.Join(cwd, findroot.DirName)
		if err := os.MkdirAll(dotDir, 0o755); err != nil {
			return err
		}

		db, err := store.Open(dotDir)
		if err != nil {
			return err
		}

		if err := db.Project.Init(); err != nil {
			return err
		}

		fmt.Printf("Proyecto tick inicializado en %s\n", dotDir)
		return nil
	},
}

func init() {
	Root.AddCommand(initCmd)
}
