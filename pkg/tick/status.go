package tick

import (
	"fmt"
	"strconv"

	"github.com/cgalvisleon/tick/internal/store"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status ID:<id>|code:<codigo> status:<pending|in_process|stop|await|done> description:<texto> percent:<0-100>",
	Short: "Registra un avance de estado para una tarea",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openStore()
		if err != nil {
			return err
		}

		kv, _ := parseKV(args)

		id := kv["ID"]
		code := kv["code"]
		if id == "" && code == "" {
			return fmt.Errorf("se requiere ID:<id> o code:<codigo>")
		}

		info, exists, err := db.Task.Find(id, code)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("tarea no encontrada")
		}

		rawStatus := kv["status"]
		if rawStatus == "" {
			return fmt.Errorf("se requiere status:<%v>", validStatusList())
		}
		normalized, ok := store.NormalizeStatus(rawStatus)
		if !ok {
			return fmt.Errorf("status invalido %q, valores validos: %v", rawStatus, validStatusList())
		}

		percent := 0
		if raw, hasPercent := kv["percent"]; hasPercent {
			percent, err = strconv.Atoi(raw)
			if err != nil {
				return fmt.Errorf("percent invalido: %w", err)
			}
		}

		if _, err := db.Task.SetStatus(info.ID, normalized, kv["description"], percent); err != nil {
			return err
		}

		return showTaskHistory(db, info.ID, "")
	},
}

func validStatusList() []string {
	return store.ValidStatuses
}

func init() {
	Root.AddCommand(statusCmd)
}
