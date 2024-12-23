# BulkPR

BulkPR is a tool designed to automate the creation of pull requests across multiple GitHub repositories. It streamlines the release management process by allowing you to create bulk PRs with customizable titles, descriptions, and branches.

## Features

- Bulk creation of pull requests across multiple repositories
- Customizable PR titles, descriptions, and branch names
- Easy configuration using YAML files

## Prerequisites

- GitHub CLI installed
- A GitHub account

## Installation

### As a Standalone Executable

To install BulkPR, download the latest executable from the [releases](https://github.com/l-lumin/gh-bulkpr/releases) page.

### As a Github CLI Extension

If you already have Github CLI installed, you can install BulkPR as a Github CLI extension for easy integration:

```sh
gh extension install l-lumin/gh-bulkpr
```

## Configuration

BulkPR uses a YAML configuration file to manage repositories and PR details. Create a `config.yaml` file with the following structure.

```yaml
repos:
  bulkpr:
    base: main
    head: develop
    title: "Test PR"
    body: "Test"
    repo: "l-lumin/gh-bulkpr"
```

## Usage

### Using the Standalone Executable

Once the configuration file is set up, run the tool using the following command:

```shell
bulkpr config.yaml
```

### Using the Github CLI Extension

```sh
gh bulkpr config.yaml
```

## Command Flags

- `--help`: Display help for the command
- `--version`: Show the version of the `gh-bulkpr` extension

## Future Plans

- Support for multiple config files
- Integration with CI/CD tools
