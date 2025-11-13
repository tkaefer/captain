package compose

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/tkaefer/captain/internal/config"
	"github.com/tkaefer/captain/internal/projects"
)

// loadEnvFiles liest zusätzliche Env-Files und baut ein Env-Array auf
func loadEnvFiles(cfg config.Config, proj projects.Project, extraFiles []string) ([]string, error) {
	env := os.Environ()

	allFiles := append([]string{}, cfg.EnvFiles...)
	allFiles = append(allFiles, extraFiles...)

	seen := make(map[string]bool)

	for _, f := range allFiles {
		if f == "" {
			continue
		}
		if seen[f] {
			continue
		}
		seen[f] = true

		path := f
		if !filepath.IsAbs(path) {
			path = filepath.Join(proj.Path, f)
		}

		file, err := os.Open(path)
		if err != nil {
			if cfg.Debug {
				fmt.Fprintf(os.Stderr, "env file %s not found or unreadable: %v\n", path, err)
			}
			continue
		}
		if cfg.Debug {
			fmt.Fprintf(os.Stderr, "loading env file %s\n", path)
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			// simple KEY=VALUE
			// wir machen keine komplexe Shell-Parsing-Magie
			env = append(env, line)
		}
		file.Close()

		if err := scanner.Err(); err != nil && cfg.Debug {
			fmt.Fprintf(os.Stderr, "error reading env file %s: %v\n", path, err)
		}
	}

	return env, nil
}

// Run führt docker compose/docker-compose im Projekt aus
// extraEnvFiles kann z.B. von CLI-Flags kommen (--env-file)
func Run(cfg config.Config, proj projects.Project, extraEnvFiles []string, args ...string) error {
	cmdName := cfg.ComposeCommand[0]
	cmdArgs := append(cfg.ComposeCommand[1:], args...)

	if cfg.Debug {
		fmt.Fprintf(os.Stderr, "running %s %s in %s\n",
			cmdName, strings.Join(cmdArgs, " "), proj.Path)
	}

	env, err := loadEnvFiles(cfg, proj, extraEnvFiles)
	if err != nil {
		return err
	}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = proj.Path
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return err
	}

	signal.Ignore(os.Interrupt)
	defer signal.Reset(os.Interrupt)

	return cmd.Wait()
}
