package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/alecthomas/kong"
	"golang.org/x/net/context"
)

var CLI struct {
	Log struct {
		// TODO: do we need filter?
		// TODO: more filter options (eg: by author)
		Output string `help:"How to format the output"`
		All    bool   `help:"Print all PRs and not just chains"`
	} `cmd:"" help:"Log PR chains"`

	Open struct {
		Filter string `arg:"" help:"Number or branch to select chain"`
		Print  bool   `help:"Print URLs instead of opening"`
	} `cmd:"" help:"Open specific PR chain"`

	Rebase struct {
		Filter string `arg:"" help:"Number or branch to select chain"`
		Push   bool   `help:"Push changes to upstream"`
		Args   string `help:"Extra args for pushing" default:"--force-with-lease"`
		Run    bool   `help:"Run the commands instead of printing"`
		Shell  string `help:"Shell for running commands" default:"$SHELL"`
	} `cmd:"" help:"Rebase specific PR chain"`

	Repo    string `help:"Repository to operate on"`
	NoCache bool   `help:"Ignore cache (cached for 1m)"` // TODO: not sure if cache will be a bad idea
}

// TODO: Automatically fetch repo from .git/config
func getOrgRepo(arg string) (string, string, error) {
	splits := strings.Split(arg, "/")
	if len(splits) != 2 {
		return "", "", fmt.Errorf("unknown repo format: %s", arg)
	}

	return splits[0], splits[1], nil
}

func main() {
	ctx := kong.Parse(&CLI)

	org, repo, err := getOrgRepo(CLI.Repo)
	if err != nil {
		log.Fatal(err)
	}

	data, err := getData(context.TODO(), org, repo, !CLI.NoCache)
	if err != nil {
		log.Fatal(err)
	}

	switch ctx.Command() {
	case "log":
		logChains(data, CLI.Log.All)
	case "open <filter>":
		openChain(data, CLI.Open.Filter, CLI.Open.Print)
	case "rebase <filter>":
		err := rebaseChain(
			data,
			CLI.Rebase.Filter,
			CLI.Rebase.Push,
			CLI.Rebase.Run,
			CLI.Rebase.Args,
			CLI.Rebase.Shell)
		if err != nil {
			log.Fatal(err)
		}
	default:
		panic(ctx.Command())
	}
}
