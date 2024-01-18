package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/tcnksm/go-gitconfig"
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

	Repo    string `help:"Repository to operate on (default: current)"`
	NoCache bool   `help:"Ignore cache (cached for 1m)"` // TODO: not sure if cache will be a bad idea
}

func getOrgRepo(arg string) (string, string, error) {
	if len(arg) > 0 {
		splits := strings.Split(arg, "/")
		if len(splits) != 2 {
			return "", "", fmt.Errorf("unknown repo format: %s", arg)
		}

		return splits[0], splits[1], nil
	}

	url, err := gitconfig.OriginURL()
	if err != nil {
		return "", "", fmt.Errorf("unable to read url in gitconfig: %v", err)
	}

	if strings.HasPrefix(url, "git@") {
		url = strings.TrimSuffix(url, ".git")

		splits := strings.Split(url, ":")
		if len(splits) != 2 {
			return "", "", fmt.Errorf("invalid repo url format %s", url)
		}

		splits = strings.Split(splits[1], "/")
		return splits[0], splits[1], nil
	}

	if strings.HasPrefix(url, "https://") {
		url = strings.TrimSuffix(url, ".git")

		splits := strings.Split(url, "/")
		if len(splits) != 5 {
			return "", "", fmt.Errorf("invalid repo url format %s", url)
		}

		return splits[3], splits[4], nil
	}

	return "", "", fmt.Errorf("could not parse repository")

}

func main() {
	cmd := "log"
	if len(os.Args) > 1 {
		ctx := kong.Parse(&CLI)
		cmd = ctx.Command()
	}

	org, repo, err := getOrgRepo(CLI.Repo)
	if err != nil {
		log.Fatal(err)
	}

	data, err := getData(context.TODO(), org, repo, !CLI.NoCache)
	if err != nil {
		log.Fatal(err)
	}

	if len(data.prs) == 0 {
		fmt.Println("No PRs and therefore no chains")
		return
	}

	switch cmd {
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
		panic(cmd)
	}
}
