package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/othmantabeche/RDF-site-summary.git/internal/database"
)

type command struct {
	Name string
	Args []string
}

type commands struct {
	registeredCommands map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	f, ok := c.registeredCommands[cmd.Name]
	if !ok {
		return fmt.Errorf("command not found")
	}
	return f(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.registeredCommands[name] = f
}

func reset(s *state, cmd command) error {
	if len(cmd.Args) > 1 {
		return fmt.Errorf("No arguments needed")
	}

	ctx := context.Background()
	err := s.db.DeletUsers(ctx)
	if err != nil {
		return fmt.Errorf("Error cleaning the database")
	}

	fmt.Println("Database cleaned suceesfully")
	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("Need more arguemnts")
	}

	name := cmd.Args[0]
	if name == "" {
		return fmt.Errorf("username cannot be empty")
	}

	ctx := context.Background()
	lookup := sql.NullString{String: name, Valid: true}

	if _, err := s.db.GetUser(ctx, lookup); err != nil {
		return fmt.Errorf("Error user do not exists: %v", err)
	}

	err := s.cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("Error traying to set the user name")
	}
	fmt.Println("Login set successfully")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("Need more arguemnts")
	}

	name := cmd.Args[0]
	if name == "" {
		return fmt.Errorf("username cannot be empty")
	}

	ctx := context.Background()
	lookup := sql.NullString{String: name, Valid: true}

	if _, err := s.db.GetUser(ctx, lookup); err == nil {
		return fmt.Errorf("username already exists")
	} else if err != sql.ErrNoRows {
		return fmt.Errorf("error checking user: %v", err)
	}

	user := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      sql.NullString{String: name, Valid: true},
	}

	data, err := s.db.CreateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("Error creating user: %v", err)
	}

	if err := s.cfg.SetUser(name); err != nil {
		return fmt.Errorf("error saving user: %v", err)
	}
	fmt.Println(data)
	fmt.Println("User added suceesfully")

	return nil
}

func getUsers(s *state, cmd command) error {
	if len(cmd.Args) > 1 {
		return fmt.Errorf("No arguments needed")
	}

	ctx := context.Background()
	data, err := s.db.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("Error traying to get users")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("Error converting to JSON: %v", err)
	}

	var users []database.User
	if err := json.Unmarshal(jsonData, &users); err != nil {
		return fmt.Errorf("error deserializing: %w", err)
	}

	for _, u := range users {
		if u.Name.String == s.cfg.Current_user_name {
			fmt.Printf("* %v (current)\n", u.Name.String)
		} else {
			fmt.Printf("* %v\n", u.Name.String)
		}
	}

	return nil
}

func agg(s *state, cmd command) error {
	if len(cmd.Args) > 1 {
		return fmt.Errorf("No arguments needed")
	}

	feedUrl := "https://www.wagslane.dev/index.xml"
	ctx := context.Background()

	res, err := fetchFeed(ctx, feedUrl)
	if err != nil {
		return fmt.Errorf("Error fetching feed: %v", err)
	}

	fmt.Println(res)

	return nil
}

func addfeed(s *state, cmd command) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("Need more arguemnts")
	}
	name := cmd.Args[0]
	url := cmd.Args[1]

	ctx := context.Background()
	user, err := s.db.GetUser(ctx, sql.NullString{String: s.cfg.Current_user_name, Valid: true})
	if err != nil {
		return fmt.Errorf("Error getting the user: %v", err)
	}

	feedPar := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       sql.NullString{String: url, Valid: true},
		UserID:    user.ID,
	}

	feed, err := s.db.CreateFeed(ctx, feedPar)
	if err != nil {
		return fmt.Errorf("Error creating feed: %v", err)
	}

	feedFollow := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	followFeed, err := s.db.CreateFeedFollow(ctx, feedFollow)
	if err != nil {
		return fmt.Errorf("couldn't create feed (%v) follow: %w", followFeed, err)
	}

	fmt.Println(feed)

	return nil
}

func feeds(s *state, cmd command) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("No arguments needed")
	}

	ctx := context.Background()

	feeds, err := s.db.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("error getting feeds: %v", err)
	}

	if len(feeds) == 0 {
		fmt.Println("No feeds found.")
		return nil
	}

	for _, feed := range feeds {
		fmt.Println(feed.Name)
		fmt.Println(feed.Url.String)
		user, err := s.db.GetUserById(ctx, feed.UserID)
		if err != nil {
			return fmt.Errorf("Error serching for user")
		}
		fmt.Println(user.Name.String)

	}

	return nil
}

func follow(s *state, cmd command) error {
	ctx := context.Background()
	user, err := s.db.GetUser(ctx, sql.NullString{String: s.cfg.Current_user_name, Valid: true})
	if err != nil {
		return fmt.Errorf("Error serching for user")
	}

	baseUrl := cmd.Args[0]
	u, err := url.Parse(baseUrl)
	if err != nil {
		return fmt.Errorf("Invalid url (%v): %v", u, err)
	}

	feed, err := s.db.GetFeedByURL(ctx, sql.NullString{String: baseUrl, Valid: true})
	if err != nil {
		return fmt.Errorf("couldn't get feed: %w", err)
	}

	feedFollow := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	ffRow, err := s.db.CreateFeedFollow(ctx, feedFollow)
	if err != nil {
		return fmt.Errorf("couldn't create feed follow: %w", err)
	}

	fmt.Println(ffRow.FeedName)
	fmt.Println(s.cfg.Current_user_name)

	return nil
}

func following(s *state, cmd command) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("No arguments needed")
	}

	ctx := context.Background()
	user, err := s.db.GetUser(ctx, sql.NullString{String: s.cfg.Current_user_name, Valid: true})
	if err != nil {
		return fmt.Errorf("Error user do not exists: %v", err)
	}

	feedFollows, err := s.db.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("Error serching for the feeds you follow")
	}

	for _, ff := range feedFollows {
		fmt.Println(ff.FeedName)
	}

	return nil
}
