package main

import (
	"os"

	"github.com/openjny/dotgh/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
