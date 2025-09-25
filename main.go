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
		Output       string   `help:"How to format the output (default,small,markdown,json)" enum:"default,small,markdown,json" default:"default"`
		All          bool     `help:"Print all PRs and not just chains"`
		Author       string   `help:"Filter by author"`
		ReviewStatus string   `help:"Filter by review status (approved,pending,unapproved,changes-requested,all)" enum:"approved,pending,unapproved,changes-requested,all" default:"all"`
		Labels       []string `help:"Filter by labels"`
		Reviewer     string   `help:"Filter by assigned reviewer"`
		DraftStatus  string   `help:"Filter by draft status (draft,ready,all)" enum:"draft,ready,all" default:"all"`
		Age          string   `help:"Filter by age (e.g., 24h, 7d)"`
		Size         string   `help:"Filter by PR size (small,medium,large,all)" enum:"small,medium,large,all" default:"all"`
	} `cmd:"" help:"Log PR chains"`

	Open struct {
		Output       string   `help:"How to format the output (default,json)" enum:"default,json" default:"default"`
		Filter       string   `arg:"" help:"Number or branch to select chain"`
		Print        bool     `help:"Print URLs instead of opening"`
		Author       string   `help:"Filter by author"`
		ReviewStatus string   `help:"Filter by review status (approved,pending,unapproved,changes-requested,all)" enum:"approved,pending,unapproved,changes-requested,all" default:"all"`
		Labels       []string `help:"Filter by labels"`
		Reviewer     string   `help:"Filter by assigned reviewer"`
		DraftStatus  string   `help:"Filter by draft status (draft,ready,all)" enum:"draft,ready,all" default:"all"`
		Age          string   `help:"Filter by age (e.g., 24h, 7d)"`
		Size         string   `help:"Filter by PR size (small,medium,large,all)" enum:"small,medium,large,all" default:"all"`
	} `cmd:"" help:"Open specific PR chain"`

	Rebase struct {
		Output string `help:"How to format the output (default,json)" enum:"default,json" default:"default"`
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
		fmt.Fprintln(os.Stderr, "No PRs and therefore no chains")
		os.Exit(1)
	}

	switch cmd {
	case "log":
		opts := FilterOptions{
			Author:       CLI.Log.Author,
			ReviewStatus: CLI.Log.ReviewStatus,
			Labels:       CLI.Log.Labels,
			Reviewer:     CLI.Log.Reviewer,
			DraftStatus:  CLI.Log.DraftStatus,
			Age:          CLI.Log.Age,
			Size:         CLI.Log.Size,
		}
		logChains(data, CLI.Log.All, opts)
	case "open <filter>":
		opts := FilterOptions{
			Author:       CLI.Open.Author,
			ReviewStatus: CLI.Open.ReviewStatus,
			Labels:       CLI.Open.Labels,
			Reviewer:     CLI.Open.Reviewer,
			DraftStatus:  CLI.Open.DraftStatus,
			Age:          CLI.Open.Age,
			Size:         CLI.Open.Size,
		}
		openChain(data, CLI.Open.Filter, CLI.Open.Print, CLI.Open.Output, opts)
	case "rebase <filter>":
		err := rebaseChain(
			data,
			CLI.Rebase.Filter,
			CLI.Rebase.Push,
			CLI.Rebase.Run,
			CLI.Rebase.Args,
			CLI.Rebase.Shell,
			CLI.Rebase.Output)
		if err != nil {
			log.Fatal(err)
		}
	default:
		panic(cmd)
	}
}
