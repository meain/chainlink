This file provides guidance to agents when working with code in this repository.

## Project Overview

`chainlink` is a Go CLI tool for managing GitHub PR chains. It allows developers to visualize, open, and rebase pull request dependency chains efficiently. The tool uses GitHub's GraphQL API to fetch PR data and provides three main commands: `log`, `open`, and `rebase`.

## Development Commands

### Building
```bash
go build -o chainlink .
```

### Running
```bash
go run . <command> [args]
# or after building:
./chainlink <command> [args]
```

### Testing
```bash
go test
go test -v  # verbose output
```

### Code Formatting
```bash
go fmt ./...
```

### Module Management
```bash
go mod tidy     # clean up dependencies
go mod download # download dependencies
```

## Code Architecture

### Core Components

- **main.go**: Entry point with CLI command parsing using Kong framework. Defines CLI struct with three main commands (Log, Open, Rebase) and their filtering options.

- **fetch.go**: GitHub GraphQL API integration. Contains data fetching logic, caching mechanism (1-minute cache in `/tmp/chainlink`), and PR data processing. Embeds GraphQL query from `request.graphql`.

- **types.go**: GraphQL response structures and internal data types. Defines `Response` struct matching GitHub's GraphQL schema and internal `pr`, `mapping`, and `data` structs.

- **filters.go**: PR filtering logic supporting multiple filter types (author, review status, labels, reviewers, draft status, age, size). Includes duration parsing for age filters.

- **log.go**: Chain visualization and logging functionality. Implements hierarchical display of PR chains with color coding for approved PRs.

- **open.go**: Browser opening functionality for PR chains. Handles URL generation and browser launching.

- **rebase.go**: Git rebase command generation. Creates shell scripts for rebasing PR chains in correct dependency order.

- **utils.go**: Utility functions for common operations.

### Data Flow

1. **Fetch Phase**: Uses embedded GraphQL query to fetch PR data from GitHub API
2. **Parse Phase**: Converts GraphQL response into internal data structures
3. **Map Phase**: Builds dependency mappings between PRs based on base/head branch relationships
4. **Filter Phase**: Applies user-specified filters to PR data
5. **Action Phase**: Executes requested command (log, open, or rebase)

### Key Data Structures

- `data`: Main container holding URL, default branch, PRs map, branch-to-PR mapping, and dependency mappings
- `pr`: Individual PR representation with metadata (number, branches, title, author, reviews, etc.)
- `mapping`: Dependency relationship showing base PR and following PRs in chain

## Configuration

Requires `CHAINLINK_TOKEN` environment variable with GitHub personal access token having repo scope access.

## Testing

The project includes unit tests in `open_test.go` focusing on PR chain filtering logic. Tests use table-driven approach to verify chain traversal algorithms.
