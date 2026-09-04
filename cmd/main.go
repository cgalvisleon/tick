package main

import (
	"fmt"
	"os"

	"github.com/cgalvisleon/tick/pkg/tick"
)

func main() {
	if err := tick.Root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
