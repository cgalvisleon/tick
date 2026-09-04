package tick

import (
	"fmt"

	"github.com/spf13/cobra"
)

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Lista los remotes configurados",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openStore()
		if err != nil {
			return err
		}
		remotes, err := db.Remote.List()
		if err != nil {
			return err
		}
		if len(remotes) == 0 {
			fmt.Println("(sin remotes)")
			return nil
		}
		for _, r := range remotes {
			fmt.Printf("%s\t%s\n", r.Name, r.Path)
		}
		return nil
	},
}

var remoteAddCmd = &cobra.Command{
	Use:   "add <nombre> <path>",
	Short: "Agrega o actualiza un remote (path del sistema de archivos)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openStore()
		if err != nil {
			return err
		}
		if err := db.Remote.Add(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("remote %s -> %s\n", args[0], args[1])
		return nil
	},
}

var remoteRemoveCmd = &cobra.Command{
	Use:   "remove <nombre>",
	Short: "Elimina un remote",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openStore()
		if err != nil {
			return err
		}
		if err := db.Remote.Remove(args[0]); err != nil {
			return err
		}
		fmt.Printf("remote %s eliminado\n", args[0])
		return nil
	},
}

func init() {
	remoteCmd.AddCommand(remoteAddCmd, remoteRemoveCmd)
	Root.AddCommand(remoteCmd)
}
