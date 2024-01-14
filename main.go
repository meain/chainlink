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
		Output string `help:"How to format the output"`
		All    bool   `help:"Print all PRs and not just chains"`
	} `cmd:"" help:"Log PR chains"`

	Open struct {
		Push bool `help:"Push changes to upstream"`
	} `cmd:"" help:"Log PR chains"`

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
	default:
		panic(ctx.Command())
	}
}
