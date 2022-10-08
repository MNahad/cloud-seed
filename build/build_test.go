package build

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildExamples(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	examplesPath := filepath.Join(cwd, "../examples")
	entries, err := os.ReadDir(examplesPath)
	if err != nil {
		t.Fatal(err)
	}
	projects := make([]string, 0, len(entries))
	for i := range entries {
		if entries[i].IsDir() {
			projects = append(projects, entries[i].Name())
		}
	}
	for i := range projects {
		t.Run(projects[i], func(t *testing.T) { Build("", filepath.Join(examplesPath, projects[i])) })
	}
}
