package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	EnvRoots      = "CAPTAIN_ROOTS"
	EnvRoot       = "CAPTAIN_ROOT"
	EnvDepth      = "CAPTAIN_DEPTH"
	EnvComposeCmd = "CAPTAIN_COMPOSE_CMD"
	EnvDebug      = "CAPTAIN_DEBUG"
	EnvEnvFiles   = "CAPTAIN_ENV_FILES"
)

type Config struct {
	Roots          []string
	Blacklist      []string
	Depth          int
	ComposeCommand []string
	Debug          bool
	EnvFiles       []string
}

func Init() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	rootsEnv := os.Getenv(EnvRoots)
	if rootsEnv == "" {
		if single := os.Getenv(EnvRoot); single != "" {
			rootsEnv = single
		} else {
			rootsEnv = home
		}
	}

	rawRoots := strings.Split(rootsEnv, ":")
	var roots []string
	for _, r := range rawRoots {
		r = strings.TrimSpace(r)
		if r != "" {
			roots = append(roots, r)
		}
	}
	if len(roots) == 0 {
		roots = []string{home}
	}

	depth := 5
	if d := os.Getenv(EnvDepth); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			depth = v
		}
	}

	blacklist := []string{
		filepath.Join(home, "Library"),
		filepath.Join(home, "Applications"),
	}

	composeCmd, err := detectComposeCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v, falling back to docker-compose\n", err)
		composeCmd = []string{"docker-compose"}
	}

	debug := os.Getenv(EnvDebug) != ""

	envFilesEnv := os.Getenv(EnvEnvFiles)
	var envFiles []string
	if envFilesEnv != "" {
		for _, p := range strings.Split(envFilesEnv, ":") {
			p = strings.TrimSpace(p)
			if p != "" {
				envFiles = append(envFiles, p)
			}
		}
	}

	cfg := Config{
		Roots:          roots,
		Blacklist:      blacklist,
		Depth:          depth,
		ComposeCommand: composeCmd,
		Debug:          debug,
		EnvFiles:       envFiles,
	}

	return cfg, nil
}

func detectComposeCommand() ([]string, error) {
	if cmd := os.Getenv(EnvComposeCmd); cmd != "" {
		fields := strings.Fields(cmd)
		if len(fields) == 0 {
			return nil, fmt.Errorf("%s empty", EnvComposeCmd)
		}
		return fields, nil
	}

	if _, err := exec.LookPath("docker-compose"); err == nil {
		return []string{"docker-compose"}, nil
	}
	if _, err := exec.LookPath("docker"); err == nil {
		return []string{"docker", "compose"}, nil
	}

	return nil, fmt.Errorf("no docker-compose or docker binary found")
}
