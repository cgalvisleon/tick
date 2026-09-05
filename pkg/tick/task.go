package tick

import (
	"fmt"
	"sort"

	"github.com/cgalvisleon/tick/internal/store"
	"github.com/cgalvisleon/tick/internal/ui"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task [ID:id|code:codigo] [campo:valor ...] [tag ...] [status]",
	Short: "Lista, crea o actualiza tareas",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openStore()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return listTasks(db)
		}

		kv, bare := parseKV(args)
		id := kv["ID"]
		code := kv["code"]
		delete(kv, "ID")
		delete(kv, "id")

		hasIdentifier := id != "" || code != ""

		if hasIdentifier && len(bare) == 1 && bare[0] == "status" {
			return showTaskHistory(db, id, code)
		}

		if hasIdentifier && len(bare) >= 1 && bare[0] == "tag" {
			info, exists, err := db.Task.Find(id, code)
			if err != nil {
				return err
			}
			if !exists {
				return fmt.Errorf("tarea no encontrada")
			}
			return handleTaskTag(db, info.ID, bare[1:])
		}

		delete(kv, "code")

		if hasIdentifier {
			info, exists, err := db.Task.Find(id, code)
			if err != nil {
				return err
			}
			if exists {
				if err := db.Task.Update(info.ID, kv); err != nil {
					return err
				}
				return showTask(db, info.ID, "")
			}
			if id != "" {
				return fmt.Errorf("tarea no encontrada")
			}
			// code given but no task has it yet: fall through to create it below.
		}

		newCode := code
		if newCode == "" {
			return fmt.Errorf("se requiere code:<codigo> para crear una tarea nueva")
		}
		if kv["name"] == "" {
			return fmt.Errorf("se requiere name:<nombre> para crear una tarea nueva")
		}
		if _, exists, _ := db.Task.Find("", newCode); exists {
			return fmt.Errorf("ya existe una tarea con code %s", newCode)
		}
		created, err := db.Task.Create(newCode, kv)
		if err != nil {
			return err
		}
		return showTask(db, created.ID, "")
	},
}

func listTasks(db *store.DB) error {
	tasks, err := db.Task.List()
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		fmt.Println("(sin tareas)")
		return nil
	}

	for _, t := range tasks {
		fmt.Printf("%-12s %-24s %-12s %s\n", t.Code, t.Name, t.Type, ui.StatusLabel(t.Status))
	}

	byType, err := db.Task.AveragesByType()
	if err != nil {
		return err
	}
	if len(byType) > 0 {
		sort.Slice(byType, func(i, j int) bool { return byType[i].Type < byType[j].Type })
		fmt.Println("\npromedio de duracion por tipo (tareas done):")
		for _, t := range byType {
			fmt.Printf("  %-15s %5.1f min  (%d tareas)\n", t.Type, t.AvgMinutes, t.Count)
		}
	}
	return nil
}

func showTask(db *store.DB, id, code string) error {
	info, exists, err := db.Task.Find(id, code)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("tarea no encontrada")
	}

	fmt.Printf("id:            %s\n", info.ID)
	fmt.Printf("code:          %s\n", info.Code)
	fmt.Printf("name:          %s\n", info.Name)
	fmt.Printf("description:   %s\n", info.Description)
	fmt.Printf("type:          %s\n", info.Type)
	fmt.Printf("assignee:      %s\n", info.Assignee)
	fmt.Printf("status:        %s\n", ui.StatusLabel(info.Status))
	fmt.Printf("planned_start: %s\n", info.PlannedStart)
	fmt.Printf("planned_end:   %s\n", info.PlannedEnd)
	if info.ActualStart != nil {
		fmt.Printf("actual_start:  %s\n", info.ActualStart.Format("2006-01-02 15:04"))
	}
	if info.ActualEnd != nil {
		fmt.Printf("actual_end:    %s\n", info.ActualEnd.Format("2006-01-02 15:04"))
	}
	if info.ActualMinutes > 0 {
		fmt.Printf("actual_minutes: %d\n", info.ActualMinutes)
	}

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
	return nil
}

func showTaskHistory(db *store.DB, id, code string) error {
	info, exists, err := db.Task.Find(id, code)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("tarea no encontrada")
	}

	history, err := db.Task.History(info.ID)
	if err != nil {
		return err
	}
	if len(history) == 0 {
		fmt.Println("(sin historial de estado)")
		return nil
	}

	fmt.Printf("historial de %s (%s):\n\n", info.Code, info.Name)
	for _, h := range history {
		fmt.Printf("%s  %-24s %s  %s\n",
			h.CreatedAt.Format("2006-01-02 15:04"), ui.StatusLabel(h.Status), ui.PercentBar(h.Status, h.Percent), h.Description)
	}

	if info.ActualStart != nil {
		fmt.Printf("\nplaneado: %s -> %s\n", info.PlannedStart, info.PlannedEnd)
		if info.ActualEnd != nil {
			fmt.Printf("real:     %s -> %s (%d min pausados, %d min reales)\n",
				info.ActualStart.Format("2006-01-02 15:04"), info.ActualEnd.Format("2006-01-02 15:04"),
				info.PausedMinutes, info.ActualMinutes)
		} else {
			fmt.Printf("real:     %s -> en curso (%d min pausados hasta ahora)\n",
				info.ActualStart.Format("2006-01-02 15:04"), info.PausedMinutes)
		}
	}
	return nil
}

func handleTaskTag(db *store.DB, taskID string, args []string) error {
	if len(args) >= 1 && args[0] == "remove" {
		if len(args) != 2 {
			return fmt.Errorf("uso: tick task ID:<id> tag remove <nombre>")
		}
		return db.Task.RemoveTag(taskID, args[1])
	}
	if len(args) != 2 {
		return fmt.Errorf("uso: tick task ID:<id> tag <nombre> <valor>")
	}
	return db.Task.SetTag(taskID, args[0], args[1])
}

func init() {
	Root.AddCommand(taskCmd)
}
