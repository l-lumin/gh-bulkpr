# BulkPR

BulkPR is a tool designed to automate the creation of pull requests across multiple GitHub repositories. It streamlines the release management process by allowing you to create bulk PRs with customizable titles, descriptions, branches, labels, assignees, reviewers, and draft status. It supports using one or more YAML configuration files.

## Features

- Bulk creation of pull requests across multiple repositories
- Customizable PR titles, descriptions (direct string or from file), branch names, labels, assignees, and reviewers
- Option to create PRs as drafts
- Support for multiple YAML configuration files (configurations are merged, with later files overriding earlier ones for the same repository key)
- Dry run mode to preview actions without making any changes
- Designed for non-interactive execution and CI/CD pipelines

## Prerequisites

- GitHub CLI (`gh`) installed and authenticated
- A GitHub account with necessary permissions for the target repositories

## Installation

### As a Standalone Executable

To install BulkPR, download the latest executable from the [releases](https://github.com/l-lumin/gh-bulkpr/releases) page for your operating system.

### As a Github CLI Extension

If you already have Github CLI installed, you can install BulkPR as a Github CLI extension for easy integration:

```sh
gh extension install l-lumin/gh-bulkpr
```

## Configuration

BulkPR uses YAML configuration files to manage repositories and PR details. You can specify one or more configuration files when running the tool.

Create a `config.yaml` (or any other name) file with the following structure:

```yaml
repos:
  # Example for a specific repository, identified by a unique key (e.g., "my-service-pr")
  my-service-pr:
    repo: "owner/repository-name"  # Required: The GitHub repository (e.g., "my-org/my-service")
    base: "main"                   # Required: The base branch for the PR
    head: "feature-branch"         # Required: The head branch (topic branch) for the PR
    title: "Feature: Implement New X Functionality"
    body: "./path/to/pr-body-template.md" # Can be a string or a path to a file for the PR body
    labels:
      - "enhancement"
      - "needs-review"
    assignees:
      - "username1"
      - "username2"
    reviewers:
      - "reviewer-username"
      - "github-org/team-slug" # For team reviewers
    draft: true                  # Optional: Set to true to create the PR as a draft (defaults to false)

  another-repo-hotfix:
    repo: "owner/another-repo"
    base: "release-v1.0"
    head: "hotfix/issue-123"
    title: "Hotfix: Critical Issue #123"
    body: "This PR addresses a critical issue found in production (Issue #123)."
    labels:
      - "bug"
      - "critical"
    # Assignees, reviewers, and draft can be omitted if not needed (draft defaults to false)
```

**Configuration Fields per Repository:**

-   `repo` (string, required): The full name of the repository, including the owner (e.g., `owner/repo-name`).
-   `base` (string, required): The name of the branch you want to merge your changes into.
-   `head` (string, required): The name of the branch containing the changes you want to merge.
-   `title` (string, required): The title of the pull request.
-   `body` (string, required): The content of the pull request. This can be a direct string or a path to a markdown file (e.g., `./pr_description.md`). If the path points to a readable file, its content will be used as the PR body. Otherwise, the string itself is used.
-   `labels` (array of strings, optional): A list of labels to add to the pull request.
-   `assignees` (array of strings, optional): A list of GitHub usernames to assign to the pull request.
-   `reviewers` (array of strings, optional): A list of GitHub usernames or team slugs (e.g., `github-org/team-slug`) to request reviews from.
-   `draft` (boolean, optional): Set to `true` to create the pull request as a draft. Defaults to `false` if omitted.

If multiple configuration files are provided, their `repos` sections are merged. If the same repository key appears in multiple files, the configuration from the last specified file takes precedence.

## Usage

### Using the Standalone Executable

Once the configuration file(s) are set up, run the tool using the following command:

```shell
bulkpr config1.yaml [config2.yaml ...]
```

### Using the Github CLI Extension

```sh
gh bulkpr config1.yaml [config2.yaml ...]
```

## Command Flags

-   `--dry-run`: Simulate PR creation without executing any `gh pr create` commands. Instead, it prints the command that would be executed for each PR. This is useful for verifying your configuration.
-   `--help`: Display help for the command.
-   `--version`: Show the version of the `gh-bulkpr` extension.

Example with dry run:
```shell
bulkpr --dry-run config.yaml
```

## CI/CD Integration and Automation

BulkPR is designed for non-interactive execution, making it suitable for CI/CD pipelines and other automation scripts.

**Exit Codes:**

The tool uses specific exit codes to indicate the outcome of its execution:
-   **`0`**: Success. All operations (pull request creations or dry run simulations) were completed successfully.
-   **`1`**: Failure. Indicates an error occurred. This can be due to:
    -   Invalid command-line arguments.
    -   Errors reading or parsing configuration files.
    -   No valid repository configurations found in the provided file(s).
    -   One or more pull requests failed to be created during an actual run (not a dry run).
    -   Other runtime errors.

Check the standard error output for specific error messages when a non-zero exit code is encountered. The `--dry-run` flag is particularly useful for testing configurations in a CI environment before making actual API calls.

## Future Plans

- Integration with CI/CD tools (further enhancements beyond current suitability)
```
