package main

import (
	"os"

	"github.com/sennadevos/kosa/cmd"

	// register backends
	_ "github.com/sennadevos/kosa/internal/backend/teable"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
