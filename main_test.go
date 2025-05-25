package main

import (
	"bytes" // Required for capturing stdout
	"flag"  // Required to reset flags for TestMainExecutionWithPartialSuccess
	"fmt"
	"io" // Required for io.ReadAll
	"os"
	"strings" // Required for string matching in error messages
	"testing"
)

// createTempYAMLFile is a helper function to create a temporary YAML file for testing.
func createTempYAMLFile(t *testing.T, content string) string {
	t.Helper()
	file, err := os.CreateTemp("", "test_config*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	_, err = file.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	return file.Name()
}

func TestReadYAMLConfig(t *testing.T) {
	testYAML := `
repos:
  test-repo-1:
    repo: "org/test-repo-1"
    base: "main"
    head: "feature-branch"
    title: "Test PR"
    body: "This is a test PR."
    labels: ["bug", "priority:high"]
    assignees: ["user1", "user2"]
    reviewers: ["user3", "org/team-review"]
    draft: true
`
	file := createTempYAMLFile(t, testYAML)
	defer os.Remove(file)

	config, err := readYAMLConfig([]string{file})
	if err != nil {
		t.Fatalf("Failed to read YAML config: %v", err)
	}

	if len(config.Repos) != 1 {
		t.Errorf("Expected 1 repo, got %d", len(config.Repos))
	}
	repoDetails := config.Repos["test-repo-1"]
	if repoDetails.Repo != "org/test-repo-1" {
		t.Errorf("Repo name mismatch: expected org/test-repo-1, got %s", repoDetails.Repo)
	}
	if !repoDetails.Draft {
		t.Errorf("Expected Draft to be true, got %v", repoDetails.Draft)
	}
}

func TestReadMultipleYAMLConfigs(t *testing.T) {
	t.Run("DistinctRepos", func(t *testing.T) {
		file1Content := `
repos:
  repo1: {repo: "org/repo1", base: "main", head: "dev", title: "R1", body: "B1", labels: ["l1"], assignees: ["a1"], reviewers: ["r1"], draft: true}
`
		file2Content := `
repos:
  repo2: {repo: "org/repo2", base: "master", head: "feat", title: "R2", body: "B2", labels: ["l2", "l3"], assignees: ["a2", "a3"], reviewers: ["r2", "r3"], draft: false}
`
		file1 := createTempYAMLFile(t, file1Content)
		defer os.Remove(file1)
		file2 := createTempYAMLFile(t, file2Content)
		defer os.Remove(file2)

		config, err := readYAMLConfig([]string{file1, file2})
		if err != nil {
			t.Fatalf("Error reading configs: %v", err)
		}
		if len(config.Repos) != 2 {
			t.Errorf("Expected 2 repos, got %d", len(config.Repos))
		}
		if !config.Repos["repo1"].Draft {
			t.Errorf("Repo1 draft flag mismatch, expected true, got %v", config.Repos["repo1"].Draft)
		}
		if config.Repos["repo2"].Draft {
			t.Errorf("Repo2 draft flag mismatch, expected false, got %v", config.Repos["repo2"].Draft)
		}
	})
}

func TestCreatePullRequest(t *testing.T) {
	originalMockRunCommand := mockRunCommand
	defer func() { mockRunCommand = originalMockRunCommand }()

	// Base test cases updated to include Draft: false
	t.Run("SingleRepoSuccess_NonDryRun", func(t *testing.T) {
		mockRunCommand = func(args ...string) error { return nil }
		configSingle := &Config{
			Repos: map[string]Repo{"test-repo-1": {Repo: "org/test-repo-1", Base: "main", Head: "feature", Title: "Test PR 1", Body: "Body", Labels: nil, Assignees: nil, Reviewers: nil, Draft: false}},
		}
		if err := createPullRequest(configSingle, false); err != nil {
			t.Errorf("createPullRequest with single repo (non-dry) failed: %v", err)
		}
	})

	t.Run("MultipleReposSuccess_NonDryRun", func(t *testing.T) {
		mockRunCommand = func(args ...string) error { return nil }
		configMultiple := &Config{
			Repos: map[string]Repo{
				"test-repo-1": {Repo: "org/test-repo-1", Base: "main", Head: "f1", Title: "T1", Body: "B1", Labels: []string{}, Assignees: []string{}, Reviewers: []string{}, Draft: false},
				"test-repo-2": {Repo: "org/test-repo-2", Base: "dev", Head: "f2", Title: "T2", Body: "B2", Labels: nil, Assignees: nil, Reviewers: nil, Draft: true},
			},
		}
		if err := createPullRequest(configMultiple, false); err != nil {
			t.Errorf("createPullRequest with multiple repos (non-dry) failed: %v", err)
		}
	})

	// Dry Run Tests updated for Draft
	t.Run("DryRunSuccess", func(t *testing.T) {
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = originalStdout; w.Close(); r.Close() }()
		mockRunCommand = func(args ...string) error { t.Error("gh pr create called in dry run"); return nil }
		configDryRun := &Config{Repos: map[string]Repo{"repo1-dry": {Repo: "org/repo1-dry", Base: "main", Head: "dev1", Title: "Dry PR1", Body: "Body1", Labels: nil, Assignees: nil, Reviewers: nil, Draft: false}}}
		err := createPullRequest(configDryRun, true)
		w.Close()
		os.Stdout = originalStdout
		var buf bytes.Buffer
		if _, e := io.Copy(&buf, r); e != nil {
			t.Fatalf("copy failed: %v", e)
		}
		if err != nil {
			t.Fatalf("Expected nil error, got %v. Output:\n%s", err, buf.String())
		}
	})

	// New Tests for Draft PRs
	t.Run("Draft_True_DryRun", func(t *testing.T) {
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = originalStdout; w.Close(); r.Close() }()
		config := &Config{Repos: map[string]Repo{"repo-draft1": {Repo: "org/draft1", Base: "b", Head: "h", Title: "T", Body: "B", Draft: true}}}
		err := createPullRequest(config, true)
		w.Close()
		os.Stdout = originalStdout
		var buf bytes.Buffer
		if _, e := io.Copy(&buf, r); e != nil {
			t.Fatalf("copy failed: %v", e)
		}
		output := buf.String()
		if err != nil {
			t.Fatalf("Expected nil error, got %v. Output:\n%s", err, output)
		}
		if !strings.Contains(output, "--draft") {
			t.Errorf("Expected output to contain '--draft'. Got: %s", output)
		}
	})

	t.Run("Draft_False_DryRun", func(t *testing.T) {
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = originalStdout; w.Close(); r.Close() }()
		config := &Config{Repos: map[string]Repo{"repo-draft2": {Repo: "org/draft2", Base: "b", Head: "h", Title: "T", Body: "B", Draft: false}}}
		err := createPullRequest(config, true)
		w.Close()
		os.Stdout = originalStdout
		var buf bytes.Buffer
		if _, e := io.Copy(&buf, r); e != nil {
			t.Fatalf("copy failed: %v", e)
		}
		output := buf.String()
		if err != nil {
			t.Fatalf("Expected nil error, got %v. Output:\n%s", err, output)
		}
		if strings.Contains(output, "--draft") {
			t.Errorf("Expected output NOT to contain '--draft'. Got: %s", output)
		}
	})

	t.Run("Draft_True_ActualRun_CaptureArgs", func(t *testing.T) {
		var capturedArgs []string
		mockRunCommand = func(args ...string) error {
			capturedArgs = make([]string, len(args))
			copy(capturedArgs, args)
			return nil
		}
		config := &Config{Repos: map[string]Repo{"repo-draft-actual1": {Repo: "org/draft-actual1", Base: "main", Head: "feature", Title: "Actual Draft Test", Body: "Body", Draft: true}}}
		err := createPullRequest(config, false)
		if err != nil {
			t.Fatalf("Expected nil error for actual run with draft, got %v", err)
		}

		foundDraft := false
		for _, arg := range capturedArgs {
			if arg == "--draft" {
				foundDraft = true
				break
			}
		}
		if !foundDraft {
			t.Errorf("Expected '--draft' in captured arguments. Got: %v", capturedArgs)
		}
	})

	t.Run("Draft_False_ActualRun_CaptureArgs", func(t *testing.T) {
		var capturedArgs []string
		mockRunCommand = func(args ...string) error {
			capturedArgs = make([]string, len(args))
			copy(capturedArgs, args)
			return nil
		}
		config := &Config{Repos: map[string]Repo{"repo-draft-actual2": {Repo: "org/draft-actual2", Base: "main", Head: "feature", Title: "Actual Non-Draft Test", Body: "Body", Draft: false}}}
		err := createPullRequest(config, false)
		if err != nil {
			t.Fatalf("Expected nil error for actual run non-draft, got %v", err)
		}

		foundDraft := false
		for _, arg := range capturedArgs {
			if arg == "--draft" {
				foundDraft = true
				break
			}
		}
		if foundDraft {
			t.Errorf("Expected NOT to find '--draft' in captured arguments for non-draft PR. Got: %v", capturedArgs)
		}
	})

	t.Run("AllFields_Including_Draft_DryRun", func(t *testing.T) {
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = originalStdout; w.Close(); r.Close() }()
		config := &Config{Repos: map[string]Repo{"repo-all-draft": {Repo: "org/all-draft", Base: "b", Head: "h", Title: "T", Body: "B", Labels: []string{"l1"}, Assignees: []string{"a1"}, Reviewers: []string{"r1"}, Draft: true}}}
		err := createPullRequest(config, true)
		w.Close()
		os.Stdout = originalStdout
		var buf bytes.Buffer
		if _, e := io.Copy(&buf, r); e != nil {
			t.Fatalf("copy failed: %v", e)
		}
		output := buf.String()
		if err != nil {
			t.Fatalf("Expected nil error, got %v. Output:\n%s", err, output)
		}
		if !strings.Contains(output, "--draft") || !strings.Contains(output, "--label \"l1\"") || !strings.Contains(output, "--assignee \"a1\"") || !strings.Contains(output, "--reviewer \"r1\"") {
			t.Errorf("Expected output to contain draft, label, assignee, and reviewer. Got: %s", output)
		}
	})

} // End of TestCreatePullRequest

// Helper function to compare two string slices
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestMainExecutionWithPartialSuccess(t *testing.T) {
	originalArgs := os.Args
	originalMockRunCommand := mockRunCommand
	origOSExit := osExit

	defer func() {
		os.Args = originalArgs
		mockRunCommand = originalMockRunCommand
		osExit = origOSExit
	}()

	var exitCode int
	osExit = func(code int) {
		exitCode = code
		panic("os.Exit called")
	}

	mockRunCommand = func(args ...string) error {
		repoFlagIndex := -1
		for i, arg := range args {
			if arg == "--repo" && i+1 < len(args) {
				repoFlagIndex = i + 1
				break
			}
		}
		if repoFlagIndex == -1 {
			return fmt.Errorf("mock: --repo not found: %v", args)
		}
		if args[repoFlagIndex] == "org/repo-fail" {
			return fmt.Errorf("simulated failure")
		}
		return nil
	}

	configFile1 := createTempYAMLFile(t, `repos: {repo-succeed: {repo: "org/repo-succeed", base: "main", head: "dev-s", title: "S", body: "B"}}`)
	defer os.Remove(configFile1)
	configFile2 := createTempYAMLFile(t, `repos: {repo-fail: {repo: "org/repo-fail", base: "main", head: "dev-f", title: "F", body: "B"}}`)
	defer os.Remove(configFile2)

	os.Args = []string{"bulkpr", configFile1, configFile2}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	recovered := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				if r == "os.Exit called" {
					recovered = true
				} else {
					panic(r)
				}
			}
		}()
		main()
	}()

	if !recovered {
		t.Fatal("os.Exit not called (via logFatalf) on partial success")
	}
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
}

func TestMainDryRunExecution(t *testing.T) {
	originalArgs := os.Args
	originalMockRunCommand := mockRunCommand
	origOSExit := osExit

	defer func() {
		os.Args = originalArgs
		mockRunCommand = originalMockRunCommand
		osExit = origOSExit
	}()

	var exitCode int = -1
	osExit = func(code int) {
		exitCode = code
		panic("os.Exit called")
	}

	ghCreateCallCount := 0
	mockRunCommand = func(args ...string) error {
		if len(args) > 0 && args[0] == "gh" && args[1] == "pr" && args[2] == "create" {
			ghCreateCallCount++
		}
		return nil
	}

	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = originalStdout; w.Close(); r.Close() }()

	configFile1 := createTempYAMLFile(t, `
repos:
  repo1-main-dry: {repo: "org/repo1-main-dry", base: "main", head: "dev-main1", title: "Main Dry PR1", body: "Body Main1", labels: ["docs"], assignees: ["dev1"], reviewers: ["rev1"], draft: true}
`)
	defer os.Remove(configFile1)
	configFile2 := createTempYAMLFile(t, `
repos:
  repo2-main-dry: {repo: "org/repo2-main-dry", base: "develop", head: "dev-main2", title: "Main Dry PR2", body: "Body Main2", labels: ["urgent", "bug"], assignees: ["dev2", "manager"], reviewers: ["rev2", "org/team-rev"], draft: false}
`)
	defer os.Remove(configFile2)

	os.Args = []string{"bulkpr", "--dry-run", configFile1, configFile2}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	main()

	w.Close()
	os.Stdout = originalStdout

	var buf bytes.Buffer
	if _, errCopy := io.Copy(&buf, r); errCopy != nil {
		t.Fatalf("Failed to copy stdout: %v", errCopy)
	}
	output := buf.String()

	if exitCode != -1 && exitCode != 0 {
		t.Errorf("Expected osExit not to be called through logFatalf, or to be 0 for help/version. Got %d. Output:\n%s", exitCode, output)
	}

	if ghCreateCallCount > 0 {
		t.Errorf("mockRunCommand for 'gh pr create' was called %d times during dry run, expected 0.", ghCreateCallCount)
	}

	expectedMsg1 := "DRY RUN: Would execute: gh pr create --draft --title \"Main Dry PR1\" --body \"Body Main1\" --base \"main\" --head \"dev-main1\" --label \"docs\" --assignee \"dev1\" --reviewer \"rev1\" --repo \"org/repo1-main-dry\""
	if !strings.Contains(output, expectedMsg1) {
		t.Errorf("Expected stdout for repo1 with draft. Got:\n%s", output)
	}
	expectedMsg2 := "DRY RUN: Would execute: gh pr create --title \"Main Dry PR2\" --body \"Body Main2\" --base \"develop\" --head \"dev-main2\" --label \"urgent\" --label \"bug\" --assignee \"dev2\" --assignee \"manager\" --reviewer \"rev2\" --reviewer \"org/team-rev\" --repo \"org/repo2-main-dry\""
	if !strings.Contains(output, expectedMsg2) { // This one should not have --draft
		t.Errorf("Expected stdout for repo2 without draft. Got:\n%s", output)
	}
	if strings.Contains(strings.Split(output, "\n")[1], "--draft") && strings.Contains(strings.Split(output, "\n")[1], "repo2-main-dry") { // Check specifically repo2's line
		t.Errorf("Repo2 (non-draft) dry run output unexpectedly contains --draft: %s", strings.Split(output, "\n")[1])
	}

}
