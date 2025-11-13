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

type source struct{ projects []Project }

func (s source) Len() int            { return len(s.projects) }
func (s source) String(i int) string { return s.projects[i].Name }

func isComposeFileName(n string) bool {
	switch n {
	case "compose.yaml", "compose.yml", "docker-compose.yaml", "docker-compose.yml":
		return true
	}
	return false
}

func isBlacklisted(path string, cfg config.Config) bool {
	for _, b := range cfg.Blacklist {
		if path == b {
			return true
		}
	}
	return false
}

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
				return filepath.SkipDir
			}

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

			if d.Type().IsRegular() && isComposeFileName(d.Name()) {
				dir := filepath.Dir(path)
				rel, _ := filepath.Rel(root, dir)
				name := strings.Trim(rel, "/")
				if name == "" {
					name = filepath.Base(dir)
				}
				projectMap[root+"::"+name] = Project{Path: dir, Name: name}
			}
			return nil
		})
	}

	var list []Project
	for _, p := range projectMap {
		list = append(list, p)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list
}

func Search(cfg config.Config, ps []Project, pat string) (Project, error) {
	for _, p := range ps {
		if p.Name == pat {
			return p, nil
		}
	}

	m := fuzzy.FindFrom(pat, source{ps})
	if len(m) == 0 {
		return Project{}, fmt.Errorf("no project matching %q", pat)
	}
	return ps[m[0].Index], nil
}

func PrintList(ps []Project) {
	if len(ps) == 0 {
		fmt.Println("No projects found.")
		return
	}
	fmt.Println("Discovered projects:")
	for _, p := range ps {
		fmt.Printf("  %-30s %s\n", p.Name, p.Path)
	}
}
