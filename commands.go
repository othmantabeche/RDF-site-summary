package main

import (
	"fmt"
)

type command struct {
	Name string
	Args []string
}

type commands struct {
	registeredCommands map[string]func(*state, command) error
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("the login handler expects a single argument, the username")
	}

	err := s.cfg.SetUser(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("Error traying to set the user name")
	}
	fmt.Println("Username set successfully")
	return nil
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
