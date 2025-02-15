package commands

import (
	"fmt"

	"github.com/ctiller15/gator/internal/config"
)

type state struct {
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

	return &newCommands
}

func NewState(cfg *config.Config) *state {
	newState := state{
		cfg: cfg,
	}

	return &newState
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("must provide username")
	}

	err := s.cfg.SetUser(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Printf("user %s has been set\n", cmd.args[0])

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
