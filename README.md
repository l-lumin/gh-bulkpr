# BulkPR

BulkPR is a tool designed to automate the creation of pull requests across multiple GitHub repositories. It streamlines the release management process by allowing you to create bulk PRs with customizable titles, descriptions, and branches.

## Features

- Bulk createtion of pull requests across multiple repositories
- Customizable PR titles, descritons, and branch names
- Easy configuration using YAML files

## Prerequisites

- GitHub CLI installed
- A GitHub account

## Installation

To install BulkPR, download the latest executable from the [releases](https://github.com/l-melon/bulkpr/releases) page.

## Configuration

BulkPR uses a YAML configuration file to manage repositories and PR details. Create a `config.yaml` file with the following structure.

```yaml
repos:
  bulkpr:
    base: main
    head: develop
    title: "Test PR"
    body: "Test"
    repo: "l-melon/bulkpr"
```

## Usage

Once the configuration file is set up, run the tool using the following command:

```shell
bulkpr config.yaml
```

## Future Plans
- Support for multiple config files
- Integration with CI/CD tools
