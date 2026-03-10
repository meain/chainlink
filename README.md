# chainlink

> Manage GitHub PR chains with ease

`chainlink` discovers PR dependency chains in a GitHub repository and lets you visualize, open, and rebase them. It works by inspecting each PR's base branch -- if PR B targets PR A's branch instead of `main`, they form a chain.

## Install

```bash
go install github.com/meain/chainlink@latest
```

Requires a GitHub token with `repo` scope:

```bash
export CHAINLINK_TOKEN="ghp_..."
```

See [Token Setup](#token-setup) for details.

## Usage

### `log` -- Visualize PR chains

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

Approved PRs are highlighted in green. Use `--all` to include standalone PRs (not just chains).

Output formats: `--output default|small|markdown|json`

### `open` -- Open a PR chain in the browser

Select a chain by branch name or PR number:

```
$ chainlink open --repo alcionai/corso group-cli

Opening https://github.com/alcionai/corso/pull/4030
Opening https://github.com/alcionai/corso/pull/4043
```

Use `--print` to print URLs without opening them.

### `rebase` -- Rebase a PR chain

Generates a shell script that rebases the chain in the correct dependency order:

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

Use `--run` to execute directly instead of printing. Use `--push` to push after each rebase.

## Filters

Available on `log` and `open` commands:

| Flag | Values |
|---|---|
| `--author` | GitHub username (prefix with `-` to exclude) |
| `--reviewer` | Assigned reviewer |
| `--review-status` | `approved`, `pending`, `unapproved`, `changes-requested`, `all` |
| `--draft-status` | `draft`, `ready`, `all` |
| `--labels` | Comma-separated (prefix with `-` to exclude) |
| `--updated-since` | Duration, e.g. `24h`, `7d` |
| `--created-since` | Duration, e.g. `24h`, `7d` |
| `--size` | `small`, `medium`, `large`, `all` |
| `--mergeable` | `mergeable`, `conflicting`, `all` |
| `--checks` | `pass`, `fail`, `pending`, `all` |

Examples:

```bash
# Show only my approved PRs
chainlink log --author meain --review-status approved

# Show PRs awaiting my review (excluding drafts, excluding my own)
chainlink log --reviewer meain --review-status unapproved --draft-status ready --author=-meain

# Show PRs updated in the last week
chainlink log --updated-since 7d

# Show conflicting PRs
chainlink log --mergeable conflicting
```

## Global Options

| Flag | Default | Description |
|---|---|---|
| `--repo <org/repo>` | current directory's origin | Repository to operate on |
| `--no-cache` | | Ignore cached data |
| `--cache-time` | `1m` | Cache duration (e.g. `1m`, `5m`, `1h`) |

## Example Workflows

### PR dashboard across multiple repos

Use `chainlink` with a list of repos to build a personal PR dashboard:

```bash
#!/bin/sh

repos="
org/repo-one
org/repo-two
org/repo-three
"

for repo in $repos; do
    output=$(chainlink log --cache-time 5m --all --output markdown --repo "$repo" "$@" 2>/dev/null)
    if [ -n "$output" ]; then
        printf "# %s\n%s\n" "$repo" "$output"
    fi
done
```

### Quick review queue

Find PRs that need your review across repos:

```bash
chainlink log --repo org/repo --reviewer yourname --review-status unapproved --draft-status ready --author=-yourname
```

### Merge-ready PRs

Find your PRs that are approved and ready to merge:

```bash
chainlink log --repo org/repo --author yourname --review-status approved
```

### Daily standup helper

Show all your pending PRs, formatted as markdown for pasting into chat:

```bash
chainlink log --repo org/repo --author yourname --review-status unapproved --output markdown
```

### Rebase and push in one shot

```bash
chainlink rebase my-feature-branch --push --run
```

## Token Setup

1. Go to [GitHub Personal Access Tokens](https://github.com/settings/tokens).
2. Click "Generate new token" > "Generate new token (classic)".
3. Check the **repo** scope.
4. Generate and export:
   ```bash
   export CHAINLINK_TOKEN="ghp_..."
   ```

For public repositories, a [fine-grained token](https://github.com/settings/tokens?type=beta) with read-only access works too.

## Alternatives

- [git-spice](https://abhinav.github.io/git-spice/)
