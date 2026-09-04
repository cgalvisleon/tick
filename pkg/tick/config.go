package tick

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config <key> [value]",
	Short: "Lee o escribe la configuracion del proyecto (guardada en .tick/tick.db)",
	Long: `Lee o escribe la configuracion del proyecto (guardada en .tick/tick.db).

Sin value: imprime el valor actual de <key>.
Con value: crea o actualiza <key>.

Claves conocidas:
  user.name   nombre del usuario, usado como autor en este proyecto
  user.email  correo del usuario, usado como autor en este proyecto
  token       token de autenticacion, usado por el futuro comando 'login'

Tambien acepta cualquier otra clave libre (no hay una lista cerrada).`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openStore()
		if err != nil {
			return err
		}

		key := args[0]
		if len(args) == 1 {
			value, exists, err := db.Config.Get(key)
			if err != nil {
				return err
			}
			if !exists {
				return fmt.Errorf("%s no esta configurado", key)
			}
			fmt.Println(value)
			return nil
		}

		if err := db.Config.Set(key, args[1]); err != nil {
			return err
		}
		fmt.Printf("%s = %s\n", key, args[1])
		return nil
	},
}

func init() {
	Root.AddCommand(configCmd)
}
