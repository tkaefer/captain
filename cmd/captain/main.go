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

	// Global config in Closure "captured"
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
		{
			Name:  "up",
			Usage: "Start a project (docker compose up -d)",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "env-file",
					Usage: "Additional env file(s) to load",
				},
			},
			Action: func(c *cli.Context) error {
				if c.NArg() < 1 {
					println("Usage: captain up <project> [--env-file file...]")
					return nil
				}
				pattern := c.Args().Get(0)
				ps := projects.Collect(cfg)
				proj, err := projects.Search(cfg, ps, pattern)
				if err != nil {
					println(err.Error())
					return nil
				}
				println("Starting", proj.Name)

				return compose.Run(cfg, proj, c.StringSlice("env-file"), "up", "-d")
			},
		},
		{
			Name:  "down",
			Usage: "Stop a project (docker compose down)",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "env-file",
					Usage: "Additional env file(s) to load",
				},
			},
			Action: func(c *cli.Context) error {
				if c.NArg() < 1 {
					println("Usage: captain down <project> [--env-file file...]")
					return nil
				}
				pattern := c.Args().Get(0)
				ps := projects.Collect(cfg)
				proj, err := projects.Search(cfg, ps, pattern)
				if err != nil {
					println(err.Error())
					return nil
				}
				println("Stopping", proj.Name)

				return compose.Run(cfg, proj, c.StringSlice("env-file"), "down")
			},
		},
		{
			Name:  "logs",
			Usage: "Show logs (docker compose logs)",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "env-file",
					Usage: "Additional env file(s) to load",
				},
			},
			Action: func(c *cli.Context) error {
				if c.NArg() < 1 {
					println("Usage: captain logs <project> [service] [--env-file file...]")
					return nil
				}
				pattern := c.Args().Get(0)
				ps := projects.Collect(cfg)
				proj, err := projects.Search(cfg, ps, pattern)
				if err != nil {
					println(err.Error())
					return nil
				}
				args := []string{"logs"}
				if c.NArg() > 1 {
					args = append(args, c.Args().Get(1))
				}
				return compose.Run(cfg, proj, c.StringSlice("env-file"), args...)
			},
		},
		{
			Name:  "ps",
			Usage: "Show containers (docker compose ps)",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "env-file",
					Usage: "Additional env file(s) to load",
				},
			},
			Action: func(c *cli.Context) error {
				if c.NArg() < 1 {
					println("Usage: captain ps <project> [--env-file file...]")
					return nil
				}
				pattern := c.Args().Get(0)
				ps := projects.Collect(cfg)
				proj, err := projects.Search(cfg, ps, pattern)
				if err != nil {
					println(err.Error())
					return nil
				}
				return compose.Run(cfg, proj, c.StringSlice("env-file"), "ps")
			},
		},
		{
			Name:  "exec",
			Usage: "Execute command in a running service container",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "env-file",
					Usage: "Additional env file(s) to load",
				},
			},
			Action: func(c *cli.Context) error {
				if c.NArg() < 2 {
					println("Usage: captain exec <project> <service> [command...] [--env-file file...]")
					return nil
				}
				pattern := c.Args().Get(0)
				service := c.Args().Get(1)
				extra := []string{}
				if c.NArg() > 2 {
					extra = c.Args()[2:]
				}

				ps := projects.Collect(cfg)
				proj, err := projects.Search(cfg, ps, pattern)
				if err != nil {
					println(err.Error())
					return nil
				}

				args := append([]string{"exec", service}, extra...)
				return compose.Run(cfg, proj, c.StringSlice("env-file"), args...)
			},
		},
		{
			Name:  "run",
			Usage: "Run a one-off command (docker compose run)",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "env-file",
					Usage: "Additional env file(s) to load",
				},
			},
			Action: func(c *cli.Context) error {
				if c.NArg() < 2 {
					println("Usage: captain run <project> <service> [command...] [--env-file file...]")
					return nil
				}
				pattern := c.Args().Get(0)
				service := c.Args().Get(1)
				extra := []string{}
				if c.NArg() > 2 {
					extra = c.Args()[2:]
				}

				ps := projects.Collect(cfg)
				proj, err := projects.Search(cfg, ps, pattern)
				if err != nil {
					println(err.Error())
					return nil
				}

				args := append([]string{"run", "--rm", service}, extra...)
				return compose.Run(cfg, proj, c.StringSlice("env-file"), args...)
			},
		},
		{
			Name:  "restart",
			Usage: "Restart a project (down + up)",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "env-file",
					Usage: "Additional env file(s) to load",
				},
			},
			Action: func(c *cli.Context) error {
				if c.NArg() < 1 {
					println("Usage: captain restart <project> [--env-file file...]")
					return nil
				}
				pattern := c.Args().Get(0)
				ps := projects.Collect(cfg)
				proj, err := projects.Search(cfg, ps, pattern)
				if err != nil {
					println(err.Error())
					return nil
				}
				envFiles := c.StringSlice("env-file")
				if err := compose.Run(cfg, proj, envFiles, "down"); err != nil {
					return err
				}
				return compose.Run(cfg, proj, envFiles, "up", "-d")
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
