package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings" // Required for strings.Join
	"sync"

	"gopkg.in/yaml.v3"
)

// Define package-level variables for log.Fatalf and os.Exit to allow mocking in tests
var (
	logFatalf = func(format string, v ...interface{}) { // Default to standard log.Fatalf behavior
		log.Printf(format, v...) // Use log.Printf to avoid direct recursion with log.Fatalf
		osExit(1)
	}
	osExit = os.Exit // Default to standard os.Exit
)

type Repo struct {
	Repo      string   `yaml:"repo"`
	Base      string   `yaml:"base"`
	Head      string   `yaml:"head"`
	Title     string   `yaml:"title"`
	Body      string   `yaml:"body"`
	Labels    []string `yaml:"labels,omitempty"`
	Assignees []string `yaml:"assignees,omitempty"`
	Reviewers []string `yaml:"reviewers,omitempty"`
	Draft     bool     `yaml:"draft,omitempty"`
}

type Config struct {
	Repos map[string]Repo `yaml:"repos"`
}

// readYAMLConfig reads YAML files and merges them into a single Config struct
func readYAMLConfig(filenames []string) (*Config, error) {
	mergedConfig := &Config{Repos: make(map[string]Repo)}

	for _, filename := range filenames {
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
		}

		var config Config
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal YAML from file %s: %w", filename, err)
		}

		for key, repo := range config.Repos {
			mergedConfig.Repos[key] = repo
		}
	}

	if len(mergedConfig.Repos) == 0 {
		return nil, fmt.Errorf("No repositories found in any configuration files")
	}

	return mergedConfig, nil
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
func createPullRequest(config *Config, dryRun bool) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(config.Repos))
	attemptedPRs := 0

	repoNames := make([]string, 0, len(config.Repos))
	for name := range config.Repos {
		repoNames = append(repoNames, name)
	}

	for _, repoName := range repoNames {
		details := config.Repos[repoName]

		bodyContent, err := os.ReadFile(details.Body)
		if err == nil {
			details.Body = string(bodyContent)
			config.Repos[repoName] = details
		} else {
			if !os.IsNotExist(err) {
				log.Printf("Warning: PR body file '%s' for repo '%s' exists but is unreadable: %v. Using raw string as body.", details.Body, repoName, err)
			}
		}

		if details.Repo == "" || details.Base == "" || details.Head == "" {
			log.Printf("Invalid repository configuration for %s, skipping\n", repoName)
			continue
		}
		attemptedPRs++

		wg.Add(1)
		go func(repoName string, currentDetails Repo) {
			defer wg.Done()

			log.Printf("Processing PR for %s (base: %s, head: %s)...\n", repoName, currentDetails.Base, currentDetails.Head)

			// Arguments for display in dry run (quoted)
			displayCmdParts := []string{"gh", "pr", "create"}
			if currentDetails.Draft {
				displayCmdParts = append(displayCmdParts, "--draft")
			}
			displayCmdParts = append(displayCmdParts, "--title", fmt.Sprintf("%q", currentDetails.Title))
			displayCmdParts = append(displayCmdParts, "--body", fmt.Sprintf("%q", currentDetails.Body))
			displayCmdParts = append(displayCmdParts, "--base", fmt.Sprintf("%q", currentDetails.Base))
			displayCmdParts = append(displayCmdParts, "--head", fmt.Sprintf("%q", currentDetails.Head))
			for _, label := range currentDetails.Labels {
				displayCmdParts = append(displayCmdParts, "--label", fmt.Sprintf("%q", label))
			}
			for _, assignee := range currentDetails.Assignees {
				displayCmdParts = append(displayCmdParts, "--assignee", fmt.Sprintf("%q", assignee))
			}
			for _, reviewer := range currentDetails.Reviewers {
				displayCmdParts = append(displayCmdParts, "--reviewer", fmt.Sprintf("%q", reviewer))
			}
			displayCmdParts = append(displayCmdParts, "--repo", fmt.Sprintf("%q", currentDetails.Repo))

			// Arguments for actual execution (not double-quoted)
			execCmdArgs := []string{"gh", "pr", "create"}
			if currentDetails.Draft {
				execCmdArgs = append(execCmdArgs, "--draft")
			}
			execCmdArgs = append(execCmdArgs,
				"--title", currentDetails.Title,
				"--body", currentDetails.Body,
				"--base", currentDetails.Base,
				"--head", currentDetails.Head)
			for _, label := range currentDetails.Labels {
				execCmdArgs = append(execCmdArgs, "--label", label)
			}
			for _, assignee := range currentDetails.Assignees {
				execCmdArgs = append(execCmdArgs, "--assignee", assignee)
			}
			for _, reviewer := range currentDetails.Reviewers {
				execCmdArgs = append(execCmdArgs, "--reviewer", reviewer)
			}
			execCmdArgs = append(execCmdArgs, "--repo", currentDetails.Repo)

			if dryRun {
				fmt.Printf("DRY RUN: Would execute: %s\n", strings.Join(displayCmdParts, " "))
				errChan <- nil
			} else {
				fmt.Printf("Creating PR for %s (base: %s, head: %s)...\n", repoName, currentDetails.Base, currentDetails.Head)
				err := runCommand(execCmdArgs...)
				if err != nil {
					log.Printf("Failed to create PR for %s: %v\n", repoName, err)
					errChan <- fmt.Errorf("failed to create PR for %s: %w", repoName, err)
				} else {
					fmt.Printf("PR created for %s successfully!\n", repoName)
					errChan <- nil
				}
			}
		}(repoName, details)
	}

	wg.Wait()
	close(errChan)

	if attemptedPRs == 0 && len(config.Repos) > 0 {
		return fmt.Errorf("no valid repository configurations found to attempt PR creation, though %d configurations were present", len(config.Repos))
	}

	var firstError error
	hasFailures := false
	for err := range errChan {
		if err != nil {
			if !hasFailures {
				firstError = err
			}
			hasFailures = true
		}
	}

	if hasFailures {
		return fmt.Errorf("one or more pull requests failed to process or create (first error: %w)", firstError)
	}

	return nil
}

func main() {
	help := flag.Bool("help", false, "Show help")
	version := flag.Bool("version", false, "Show version")
	dryRun := flag.Bool("dry-run", false, "Simulate PR creation without executing commands")

	flag.Parse()

	if *help {
		fmt.Println("Usage: gh bulkpr <config-file1> [config-file2] ...")
		fmt.Println("Create pull requests in multiple repositories using one or more configuration files.")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		osExit(0)
	}

	if *version {
		fmt.Println("gh-bulkpr v0.1.0")
		osExit(0)
	}

	if len(flag.Args()) < 1 {
		logFatalf("Usage: bulkpr <config-file1> [config-file2] ...")
	}
	configFiles := flag.Args()

	config, err := readYAMLConfig(configFiles)
	if err != nil {
		logFatalf("Error reading config files: %v", err)
	}

	err = createPullRequest(config, *dryRun)
	if err != nil {
		logFatalf("Error creating pull requests: %v", err)
	}
}
