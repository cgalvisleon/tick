package tick

import (
	"fmt"
	"sort"

	"github.com/cgalvisleon/tick/internal/store"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project [campo:valor ...] [tag <nombre> <valor> | tag remove <nombre>]",
	Short: "Muestra o actualiza los datos del proyecto actual",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openStore()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return showProject(db)
		}

		if args[0] == "tag" {
			return handleProjectTag(db, args[1:])
		}

		kv, _ := parseKV(args)
		delete(kv, "id")
		delete(kv, "ID")
		if err := db.Project.Set(kv); err != nil {
			return err
		}
		return showProject(db)
	},
}

func showProject(db *store.DB) error {
	info, exists, err := db.Project.Get()
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("no hay proyecto inicializado; corre 'tick init'")
	}

	fmt.Printf("id:          %s\n", info.ID)
	fmt.Printf("code:        %s\n", info.Code)
	fmt.Printf("name:        %s\n", info.Name)
	fmt.Printf("description: %s\n", info.Description)
	fmt.Printf("created_at:  %s\n", info.CreatedAt.Format("2006-01-02 15:04"))

	if len(info.Tags) > 0 {
		fmt.Println("tags:")
		keys := make([]string, 0, len(info.Tags))
		for k := range info.Tags {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("  %s: %s\n", k, info.Tags[k])
		}
	}

	byType, err := db.Task.AveragesByType()
	if err != nil {
		return err
	}
	if len(byType) > 0 {
		fmt.Println("\npromedio de duracion por tipo (tareas done):")
		for _, t := range byType {
			fmt.Printf("  %-15s %5.1f min  (%d tareas)\n", t.Type, t.AvgMinutes, t.Count)
		}
	}

	return nil
}

func handleProjectTag(db *store.DB, args []string) error {
	if len(args) >= 1 && args[0] == "remove" {
		if len(args) != 2 {
			return fmt.Errorf("uso: tick project tag remove <nombre>")
		}
		return db.Project.RemoveTag(args[1])
	}
	if len(args) != 2 {
		return fmt.Errorf("uso: tick project tag <nombre> <valor>")
	}
	return db.Project.SetTag(args[0], args[1])
}

func init() {
	Root.AddCommand(projectCmd)
}
