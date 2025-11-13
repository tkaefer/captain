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

func loadEnvFiles(cfg config.Config, proj projects.Project, extra []string) ([]string, error) {
	env := os.Environ()
	all := append(append([]string{}, cfg.EnvFiles...), extra...)
	seen := map[string]bool{}

	for _, f := range all {
		if f == "" || seen[f] {
			continue
		}
		seen[f] = true

		path := f
		if !filepath.IsAbs(path) {
			path = filepath.Join(proj.Path, f)
		}

		file, err := os.Open(path)
		if err != nil {
			continue
		}
		defer file.Close()

		sc := bufio.NewScanner(file)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				env = append(env, line)
			}
		}
	}
	return env, nil
}

func Run(cfg config.Config, proj projects.Project, extraEnv []string, args ...string) error {
	cmdName := cfg.ComposeCommand[0]
	cmdArgs := append(cfg.ComposeCommand[1:], args...)

	env, _ := loadEnvFiles(cfg, proj, extraEnv)

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
