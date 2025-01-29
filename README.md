# chainlink

> Manage GitHub PR chains with ease

`chainlink` simplifies the management of PR chains on GitHub, allowing you to efficiently log, open, and rebase them.

## Features

- **Log PR chains**: Visualize your PR chains and their dependencies in a clear, hierarchical format.
  - Use filters to narrow down the list:
    - `--author <author>`:  See PRs by a specific author.
    - `--review-status <status>`: Filter by review status (`approved`, `pending`, `all`).
    - `--labels <label1>,<label2>`: Find PRs with specific labels.
    - ... (and other filters like `--reviewer`, `--draft-status`, `--age`, `--size`)
  - Example:
    ```
    $ chainlink log --repo alcionai/corso --author ashmrtn --review-status approved
    ```

- **Open PR chains**: Open an entire PR chain in your browser, in the correct dependency order, with a single command.
  - Select a chain by branch name or PR number.
  - Supports the same filters as `log` to help you find the right chain.
  - Example:
    ```
    $ chainlink open --repo alcionai/corso group-cli
    ```

- **Rebase PR chains**:  Rebase a complete PR chain onto `main` (or another branch) to keep your branches up-to-date.
  - Automatically handles the correct rebase order based on PR dependencies.
  - `--push` flag to automatically push rebased branches.
  - Example:
    ```
    $ chainlink rebase 3217-model-mod-time --push
    ```
    This will output and optionally run shell commands to rebase the chain.

## Configuration

You need a GitHub personal access token to use `chainlink`. Set it as an environment variable: `CHAINLINK_TOKEN`.

### How to get `CHAINLINK_TOKEN`

1. Go to [GitHub Personal Access Tokens](https://github.com/settings/tokens).
2. Click "Generate new token" > "Generate new token (classic)".
3. Give it a descriptive "Note" and set an "Expiration" if desired.
4. In "Select scopes", check the "repo" scope.
5. Click "Generate token".
6. Copy the generated token and set it as an environment variable:
   ```bash
   export CHAINLINK_TOKEN="gh_your_token_here"
   ```
   For public repositories, consider using a [fine-grained token](https://github.com/settings/tokens?type=beta) with read-only access for enhanced security.

## Examples

### Log PR chains

_Approved PRs are highlighted in green._

```
$ chainlink log --repo alcionai/corso

#4051 Basic code for backup cleanup (ashmrtn) [3217-incomplete-backup-cleanup]
 #4065 Add and populate mod time for BaseModel (ashmrtn) [3217-model-mod-time]
  #4066 Exclude recently created models from garbage collection (ashmrtn) [3217-delay-model-gc]
#4030 Create backup collections for Group's default SharePoint site (meain) [group-files]
 #4043 Group CLI (meain) [group-cli]
#4050 add handlers for channels (neha-Gupta1) [channelHandlers]
 #4068 channels and messages API (neha-Gupta1) [HandlerImplemenation]
```

### Open a PR chain

```
$ chainlink open --repo alcionai/corso group-cli

Opening https://github.com/alcionai/corso/pull/4030
Opening https://github.com/alcionai/corso/pull/4043
```

### Rebase a PR chain

```
$ chainlink rebase 3217-model-mod-time --push

#!/bin/sh

set -ex

git checkout 3217-incomplete-backup-cleanup
git rebase --update-refs main
git push --force-with-lease

git checkout 3217-model-mod-time
git rebase --update-refs 3217-incomplete-backup-cleanup
git push --force-with-lease

git checkout 3217-delay-model-gc
git rebase --update-refs 3217-model-mod-time
git push --force-with-lease
```

## Filter Options Summary

These options are available for both `log` and `open` commands:

- `--author <author>`
- `--review-status <status>` (`approved`, `pending`, `all`)
- `--labels <label1>,<label2>`
- `--reviewer <reviewer>`
- `--draft-status <status>` (`draft`, `ready`, `all`)
- `--age <duration>` (e.g., `24h`, `7d`)
- `--size <size>` (`small`, `medium`, `large`, `all`)

## Alternatives

- [git-spice](https://abhinav.github.io/git-spice/)
