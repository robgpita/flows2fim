package main

import (
	"fmt"
	"os"

	"flows2fim/cmd/calculator"
	"flows2fim/cmd/hello"
	"flows2fim/internal/config"
)

func main() {

	config.LoadConfig()

	if len(os.Args) < 2 {
		fmt.Println("Please provide a subcommand: hello or calculator")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "hello":
		hello.Run(os.Args[2])
	case "calculator":
		calculator.Run(os.Args[2:])
	default:
		fmt.Println("Unknown subcommand")
		os.Exit(1)
	}
}
