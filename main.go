package main

import (
	"fmt"
	"log"
	"os"

	"github.com/othmantabeche/RDF-site-summary.git/internal/config"
)

type state struct {
	cfg *config.Config
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
		return
	}

	programState := &state{cfg: &cfg}

	cmds := commands{
		registeredCommands: make(map[string]func(*state, command) error)}

	cmds.register("login", handlerLogin)

	if len(os.Args) < 2 {
		log.Fatal("Usage: cli <command> [args...]")
	}

	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]

	err = cmds.run(programState, command{Name: cmdName, Args: cmdArgs})
	if err != nil {
		log.Fatal(err)
	}
}
