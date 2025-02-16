package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/ctiller15/gator/internal/config"
	"github.com/ctiller15/gator/internal/database"
	"github.com/google/uuid"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commandMap map[string]func(*state, command) error
}

func NewCommand(name string, args []string) command {
	newCommand := command{
		name: name,
		args: args,
	}

	return newCommand
}

func NewCommands() *commands {
	newCommands := commands{
		commandMap: make(map[string]func(*state, command) error),
	}

	newCommands.register("login", handlerLogin)
	newCommands.register("register", handlerRegister)

	return &newCommands
}

func NewState(cfg *config.Config, db *database.Queries) *state {
	newState := state{
		db:  db,
		cfg: cfg,
	}

	return &newState
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("must provide username")
	}

	ctx := context.Background()

	user, err := s.db.GetUser(ctx, cmd.args[0])
	if err != nil {
		return err
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("user %s has been set\n", cmd.args[0])

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("must provide username")
	}

	ctx := context.Background()

	currentTime := time.Now()
	args := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		Name:      cmd.args[0],
	}

	user, err := s.db.CreateUser(ctx, args)

	if err != nil {
		return err
	}

	err = s.cfg.SetUser(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Printf("user %s has been created\n", cmd.args[0])
	fmt.Printf("%+v\n", user)

	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commandMap[name] = f
}

func (c *commands) Run(s *state, cmd command) error {
	commandFunc, ok := c.commandMap[cmd.name]

	if !ok {
		return fmt.Errorf("command %s not found", cmd.name)
	}

	err := commandFunc(s, cmd)
	if err != nil {
		return err
	}

	return nil
}
