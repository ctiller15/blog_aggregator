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
	newCommands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	newCommands.register("feeds", handlerGetFeeds)
	newCommands.register("follow", middlewareLoggedIn(handlerFollow))
	newCommands.register("following", middlewareLoggedIn(handlerFollowing))
	newCommands.register("unfollow", middlewareLoggedIn(handlerUnfollow))

	return &newCommands
}

func NewState(cfg *config.Config, db *database.Queries) *state {
	newState := state{
		db:  db,
		cfg: cfg,
	}

	return &newState
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	ctx := context.Background()

	userFeeds, err := s.db.GetFeedFollowsForUser(ctx, user.Name)
	if err != nil {
		return err
	}

	for _, feed := range userFeeds {
		fmt.Printf("%s\n", feed.FeedName)
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	ctx := context.Background()

	if len(cmd.args) == 0 {
		return fmt.Errorf("must provide url")
	}

	url := cmd.args[0]

	current_time := time.Now()

	feed, err := s.db.GetFeedByUrl(ctx, url)
	if err != nil {
		return err
	}

	args := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: current_time,
		UpdatedAt: current_time,
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	feedFollow, err := s.db.CreateFeedFollow(ctx, args)
	if err != nil {
		return err
	}

	fmt.Printf("feed name: %s, feed user: %s", feedFollow.FeedName, feedFollow.UserName)
	return nil
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

func handlerUnfollow(s *state, cmd command, user database.User) error {
	ctx := context.Background()

	if len(cmd.args) < 1 {
		return fmt.Errorf("must provide a feed url")
	}

	feedUrl := cmd.args[0]
	deleteFeedFollowByUrlParams := database.DeleteFeedFollowByUrlParams{
		UserID: user.ID,
		Url:    feedUrl,
	}
	err := s.db.DeleteFeedFollowByUrl(ctx, deleteFeedFollowByUrlParams)
	if err != nil {
		return err
	}

	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	ctx := context.Background()

	if len(cmd.args) < 2 {
		return fmt.Errorf("must provide both a name and a url")
	}

	feedName := cmd.args[0]
	feedUrl := cmd.args[1]

	currentTime := time.Now()
	feedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		Name:      feedName,
		Url:       feedUrl,
	}

	feed, err := s.db.CreateFeed(ctx, feedParams)
	if err != nil {
		return err
	}

	feedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	_, err = s.db.CreateFeedFollow(ctx, feedFollowParams)
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

	err = s.db.DeleteFeeds(ctx)
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

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	ctx := context.Background()

	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(ctx, s.cfg.CurrentUserName)
		if err != nil {
			return err
		}

		return handler(s, cmd, user)
	}
}
