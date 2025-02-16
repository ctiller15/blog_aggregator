package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/ctiller15/gator/internal/config"
	"github.com/ctiller15/gator/internal/database"
	"github.com/ctiller15/gator/internal/rss"
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
	newCommands.register("reset", handlerReset)
	newCommands.register("users", handlerGetUsers)
	newCommands.register("agg", handlerAggregation)
	newCommands.register("addfeed", handlerAddFeed)
	newCommands.register("feeds", handlerGetFeeds)

	return &newCommands
}

func NewState(cfg *config.Config, db *database.Queries) *state {
	newState := state{
		db:  db,
		cfg: cfg,
	}

	return &newState
}

func handlerGetFeeds(s *state, cmd command) error {
	ctx := context.Background()
	feedData, err := s.db.GetFeeds(ctx)
	if err != nil {
		return err
	}

	for _, feed := range feedData {
		fmt.Printf("%+v\n", feed)
	}

	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("must provide both a name and a url")
	}

	feedName := cmd.args[0]
	feedUrl := cmd.args[1]

	ctx := context.Background()

	username := s.cfg.CurrentUserName
	userData, err := s.db.GetUser(ctx, username)
	if err != nil {
		return err
	}

	currentTime := time.Now()
	feedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		Name:      feedName,
		Url:       feedUrl,
		UserID:    userData.ID,
	}

	feed, err := s.db.CreateFeed(ctx, feedParams)
	if err != nil {
		return err
	}

	fmt.Printf("feed: %+v\n", feed)
	return nil
}

func handlerAggregation(s *state, cmd command) error {
	ctx := context.Background()
	feedUrl := "https://www.wagslane.dev/index.xml"
	feed, err := rss.FetchFeed(ctx, feedUrl)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", feed)
	return nil
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

func handlerReset(s *state, cmd command) error {
	ctx := context.Background()

	err := s.db.DeleteUser(ctx)

	if err != nil {
		return err
	}

	fmt.Println("Deletion successful")
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	ctx := context.Background()

	users, err := s.db.GetUsers(ctx)

	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}

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
