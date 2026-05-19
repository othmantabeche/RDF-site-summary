package main

import (
	"context"
	"database/sql"
	"fmt"
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
