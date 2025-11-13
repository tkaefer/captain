package projects

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sahilm/fuzzy"
	"github.com/tkaefer/captain/internal/config"
)

type Project struct {
	Path string
	Name string
}

type source struct {
	projects []Project
}

func (s source) Len() int            { return len(s.projects) }
func (s source) String(i int) string { return s.projects[i].Name }

func isComposeFileName(name string) bool {
	switch name {
	case "compose.yaml", "compose.yml", "docker-compose.yaml", "docker-compose.yml":
		return true
	default:
		return false
	}
}

func isBlacklisted(path string, cfg config.Config) bool {
	for _, b := range cfg.Blacklist {
		if path == b {
			return true
		}
	}
	return false
}

// Collect scans all configured roots for projects
func Collect(cfg config.Config) []Project {
	projectMap := make(map[string]Project)

	for _, root := range cfg.Roots {
		root = filepath.Clean(root)

		filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if cfg.Debug {
					fmt.Fprintf(os.Stderr, "skip %s: %v\n", path, err)
				}
				if d != nil && d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if d.IsDir() && isBlacklisted(path, cfg) {
				if cfg.Debug {
					fmt.Fprintf(os.Stderr, "skip blacklisted dir %s\n", path)
				}
				return filepath.SkipDir
			}

			// Tiefe begrenzen
			if rel, err := filepath.Rel(root, path); err == nil {
				depth := 0
				if rel != "." {
					depth = len(strings.Split(rel, string(os.PathSeparator)))
				}
				if depth > cfg.Depth {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}

			if !d.Type().IsRegular() {
				return nil
			}

			if !isComposeFileName(d.Name()) {
				return nil
			}

			projDir := filepath.Dir(path)
			rel, _ := filepath.Rel(root, projDir)
			name := strings.Trim(rel, string(os.PathSeparator))
			if name == "" {
				name = filepath.Base(projDir)
			}

			key := root + "::" + name
			projectMap[key] = Project{
				Path: projDir,
				Name: name,
			}

			return nil
		})
	}

	projects := make([]Project, 0, len(projectMap))
	for _, p := range projectMap {
		projects = append(projects, p)
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	if cfg.Debug {
		fmt.Fprintf(os.Stderr, "found %d projects\n", len(projects))
	}

	return projects
}

func Search(cfg config.Config, projects []Project, pattern string) (Project, error) {
	if len(projects) == 0 {
		return Project{}, fmt.Errorf("no projects found")
	}

	// exakter Treffer
	for _, p := range projects {
		if p.Name == pattern {
			return p, nil
		}
	}

	src := source{projects: projects}
	matches := fuzzy.FindFrom(pattern, src)

	if len(matches) == 0 {
		return Project{}, fmt.Errorf("no project matching %q found", pattern)
	}
	if len(matches) > 1 && cfg.Debug {
		fmt.Fprintf(os.Stderr, "multiple matches for %q, choosing %s\n",
			pattern, matches[0].Str)
	}

	return projects[matches[0].Index], nil
}

func PrintList(projects []Project) {
	if len(projects) == 0 {
		fmt.Println("No projects found.")
		return
	}
	fmt.Println("Discovered projects:")
	for _, p := range projects {
		fmt.Printf("  %-30s  %s\n", p.Name, p.Path)
	}
}
