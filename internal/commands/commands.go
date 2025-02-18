package commands

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/ctiller15/gator/internal/config"
	"github.com/ctiller15/gator/internal/database"
	"github.com/ctiller15/gator/internal/rss"
	"github.com/google/uuid"
	"github.com/lib/pq"
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
	newCommands.register("browse", middlewareLoggedIn(handlerBrowseFeeds))

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

	if len(cmd.args) < 0 {
		return fmt.Errorf("must provide a scrape duration")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Collecting feeds every %s\n", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		err = scrapeFeeds(ctx, s)
		if err != nil {
			return err
		}
	}
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

func handlerBrowseFeeds(s *state, cmd command, user database.User) error {
	ctx := context.Background()

	postLimit := 2
	if len(cmd.args) > 0 {
		limit, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return err
		}

		postLimit = limit
	}

	getPostsForUserParams := database.GetPostsForUserParams{
		ID:    user.ID,
		Limit: int32(postLimit),
	}
	results, err := s.db.GetPostsForUser(ctx, getPostsForUserParams)
	if err != nil {
		return err
	}

	fmt.Println("browsing!")
	fmt.Println(len(results))
	for _, result := range results {
		fmt.Printf("%+v\n", result)
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

func scrapeFeeds(ctx context.Context, s *state) error {
	fmt.Println("visiting next feed...\n")
	feed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		return err
	}

	markFeedFetchedResult, err := s.db.MarkFeedFetched(ctx, feed.ID)
	if err != nil {
		return err
	}

	feedResults, err := rss.FetchFeed(ctx, markFeedFetchedResult.Url)
	if err != nil {
		return err
	}

	for _, feedResult := range feedResults.Channel.Item {

		currentTime := time.Now()
		parsedPubTime, err := parsePubTime(feedResult.PubDate)
		if err != nil {
			return err
		}
		savePostParams := database.CreatePostParams{
			ID:        uuid.New(),
			CreatedAt: currentTime,
			UpdatedAt: currentTime,
			Title:     feedResult.Title,
			Url:       feedResult.Link,
			Description: sql.NullString{
				String: feedResult.Description,
				Valid:  true,
			},
			PublishedAt: sql.NullTime{
				Time:  parsedPubTime,
				Valid: true,
			},
			FeedID: markFeedFetchedResult.ID,
		}
		_, err = s.db.CreatePost(ctx, savePostParams)
		if err != nil {
			pgErr, ok := err.(*pq.Error)
			if ok {
				if pgErr.Code == "23505" && pgErr.Code.Name() == "unique_violation" {
					continue
				}
			}
			return err
		}
	}

	return nil
}

var timeFormats = []string{
	time.RFC1123Z,
	time.RFC3339Nano,
	time.RFC3339,
	time.RFC1123Z,
	time.RFC1123,
	time.RFC850,
	time.RFC822Z,
	time.RFC822,
	time.RubyDate,
	time.UnixDate,
	time.ANSIC,
	time.Layout,
}

func parsePubTime(timeStr string) (time.Time, error) {
	var err error
	for _, format := range timeFormats {
		parsedTime, err := time.Parse(format, timeStr)
		if err != nil {
			continue
		}

		return parsedTime, nil
	}

	return time.Time{}, err

}
