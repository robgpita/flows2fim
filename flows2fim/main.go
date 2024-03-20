package main

import (
	"fmt"
	"os"

	"flows2fim/cmd/controls"
	"flows2fim/cmd/fim"
	"flows2fim/internal/config"
)

func main() {

	config.LoadConfig()

	if len(os.Args) < 2 {
		fmt.Println("Please provide a subcommand: controls | fim")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "controls":
		controls.Run(os.Args[2:])
	case "fim":
		fim.Run(os.Args[2:])
	default:
		fmt.Println("Unknown subcommand")
		os.Exit(1)
	}
}
