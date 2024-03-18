package calculator

import (
	"fmt"
	"go-cli-app/pkg/utils"
	"strconv"
)

func Run(args []string) {
	defer utils.PrintSeparator()
	if len(args) != 3 {
		fmt.Println("Usage: calculator [add|sub] num1 num2")
		return
	}

	num1, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("Invalid number:", args[1])
		return
	}

	num2, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Println("Invalid number:", args[2])
		return
	}

	switch args[0] {
	case "add":
		fmt.Println("Result:", add(num1, num2))
	case "sub":
		fmt.Println("Result:", subtract(num1, num2))
	default:
		fmt.Println("Unknown operation:", args[0])
	}
}
