package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/ctiller15/gator/internal/commands"
	"github.com/ctiller15/gator/internal/config"
	"github.com/ctiller15/gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	configStruct, err := config.Read()
	if err != nil {
		log.Fatalf("error occurred: %v", err)
	}

	db, err := sql.Open("postgres", configStruct.DB_URL)
	if err != nil {
		log.Fatalf("error occurred: %v", err)
	}

	dbQueries := database.New(db)

	newState := commands.NewState(&configStruct, dbQueries)

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
		os.Exit(1)
	}
}
