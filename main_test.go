package main

import (
	"fmt"
	"os"
	"testing"
)

func TestReadYAMLConfig(t *testing.T) {
	// Create a temporary test YAML file
	testYAML := `
repos:
  test-repo-1:
    repo: "org/test-repo-1"
    base: "main"
    head: "feature-branch"
    title: "Test PR"
    body: "This is a test PR."
`
	file, err := os.CreateTemp("", "test_config*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())

	_, err = file.WriteString(testYAML)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Read the YAML file
	config, err := readYAMLConfig(file.Name())
	if err != nil {
		t.Fatalf("Failed to read YAML config: %v", err)
	}

	// Assertions
	if len(config.Repos) != 1 {
		t.Errorf("Expected 1 repo, got %d", len(config.Repos))
	}
	if config.Repos["test-repo-1"].Repo != "org/test-repo-1" {
		t.Errorf("Repo name mismatch: expected org/test-repo-1, got %s", config.Repos["test-repo-1"].Repo)
	}
}

func TestCreatePullRequest(t *testing.T) {
	// Mock runCommand
	mockRunCommand = func(args ...string) error {
		if args[0] == "gh" && args[1] == "pr" && args[2] == "create" {
			return nil
		}
		return fmt.Errorf("Invalid command: %v", args)
	}

	config := &Config{
		Repos: map[string]Repo{
			"test-repo-1": {
				Repo:  "org/test-repo-1",
				Base:  "main",
				Head:  "feature-branch",
				Title: "Test PR",
				Body:  "This is a test PR.",
			},
		},
	}

	err := createPullRequest(config)
	if err != nil {
		t.Fatalf("Failed to create PR: %v", err)
	}
}

