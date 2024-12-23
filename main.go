package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"gopkg.in/yaml.v3"
)

type Repo struct {
	Repo  string `yaml:"repo"`
	Base  string `yaml:"base"`
	Head  string `yaml:"head"`
	Title string `yaml:"title"`
	Body  string `yaml:"body"`
}

type Config struct {
	Repos map[string]Repo `yaml:"repos"`
}

// readYAMLConfig read the YAML file and parses it into the Config struct
func readYAMLConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	if len(config.Repos) == 0 {
		return nil, fmt.Errorf("No repositories found in configuration")
	}

	return &config, nil
}

var mockRunCommand func(args ...string) error

// runCommand executes a shell command in a given directory
func runCommand(args ...string) error {
	if mockRunCommand != nil {
		return mockRunCommand(args...)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error running command %s: %v", args, err)
	}

	return nil
}

// createPullRequest generates a PR for each repository in the YAML file
func createPullRequest(config *Config) error {
	var wg sync.WaitGroup
	for repoName, details := range config.Repos {
		if details.Repo == "" || details.Base == "" || details.Head == "" {
			log.Printf("Invalid repository configuration for %s/n", repoName)
			continue
		}

		wg.Add(1)
		go func(repoName string, details Repo) {
			defer wg.Done()

			fmt.Printf("Creating PR for %s (base: %s, head: %s)...\n", repoName, details.Base, details.Head)

			err := runCommand("gh", "pr", "create", "--title", details.Title, "--body", details.Body, "--base", details.Base, "--head", details.Head, "--repo", details.Repo)

			if err != nil {
				log.Printf("Failed to create PR for %s: %v\n", repoName, err)
			} else {
				fmt.Printf("PR create for %s successfully!\n", repoName)
			}
		}(repoName, details)

	}
	wg.Wait()

	return nil
}
func main() {
	// Define flags
	help := flag.Bool("help", false, "Show help")
	version := flag.Bool("version", false, "Show version")

	// Parse flags
	flag.Parse()

	if *help {
		fmt.Println("Usage: gh bulkpr <config-file>")
		fmt.Println("Create pull requests in multiple repositories using a single command")
	}

	if *version {
		fmt.Println("gh-bulkpr v0.1.0")
	}

	if len(flag.Args()) < 1 {
		log.Fatal("Usage: bulkpr <config-file>")
	}
	configFile := flag.Args()[0]

	// Read the configuration file
	config, err := readYAMLConfig(configFile)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Create pull requests
	err = createPullRequest(config)
	if err != nil {
		log.Fatalf("Error creating pull requests: %v", err)
	}
}
