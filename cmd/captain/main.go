package main

import (
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/tkaefer/captain/internal/compose"
	"github.com/tkaefer/captain/internal/config"
	"github.com/tkaefer/captain/internal/projects"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("init config: %v", err)
	}

	app := cli.NewApp()
	app.Name = "captain"
	app.Usage = "Manage multiple docker compose projects from any directory"
	app.Version = "0.5.0-fork"

	app.Commands = []cli.Command{
		{
			Name:    "ls",
			Aliases: []string{"list"},
			Usage:   "List discovered projects",
			Action: func(c *cli.Context) error {
				ps := projects.Collect(cfg)
				projects.PrintList(ps)
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
