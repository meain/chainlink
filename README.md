# chainlink

> Get a handle on PRs chains on GitHub

```
Usage: chainlink <command>

Flags:
  -h, --help           Show context-sensitive help.
      --repo=STRING    Repository to operate on
      --no-cache       Ignore cache (cached for 1m)

Commands:
  log
    Log PR chains

  open <filter>
    Open specific PR chain

  rebase <filter>
    Rebase specific PR chain

Run "chainlink <command> --help" for more information on a command.
```

# Configuration

You'll have to setup `CHAINLINK_TOKEN` env variable to a GitHub personal access token.

### Getting the token for `CHAINLINK_TOKEN`

- Go to [tokens](https://github.com/settings/tokens)
- Click on "Generate new token" > "Generate new token (classic)"
- Set note, and expiration your preferred values
- Select "repo" in "Select scopes" section
- Click on "Generate token"
- Set the token you get as env variable (export CHAINLINK_TOKEN="gh_...")

> If you only need to use on public repos, you can generate a [fine
> grained](https://github.com/settings/tokens?type=beta) token with
> just read only access to public repositories.

# Examples

## Log PR chains

_Approved PRs will be in green._

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

## Open a PR chain in the correct order

```
$ chainlink open --repo alcionai/corso group-cli

Opening https://github.com/alcionai/corso/pull/4030
Opening https://github.com/alcionai/corso/pull/4043
```

## Rebase a PR chain with main

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

# Alternatives

- [git-spice](https://abhinav.github.io/git-spice/)
