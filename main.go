package main

import (
	"log"
	"os"

	"github.com/ctiller15/gator/internal/commands"
	"github.com/ctiller15/gator/internal/config"
)

func main() {
	configStruct, err := config.Read()

	if err != nil {
		log.Fatalf("error occurred: %v", err)
	}

	newState := commands.NewState(&configStruct)

	newCommands := commands.NewCommands()

	args := os.Args
	if len(args) < 2 {
		log.Fatal("requires at least 2 arguments")
	}

	commandName := args[1]
	commandArgs := args[2:]

	command := commands.NewCommand(commandName, commandArgs)
	err = newCommands.Run(newState, command)
	if err != nil {
		log.Fatalf("error occurred: %v", err)
	}
}
