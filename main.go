package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/tcnksm/go-gitconfig"
	"golang.org/x/net/context"
)

var CLI struct {
	Log struct {
		Output       string   `help:"How to format the output (default,small,markdown,json)" enum:"default,small,markdown,json" default:"default"`
		All          bool     `help:"Print all PRs and not just chains"`
		Author       string   `help:"Filter by author (prefix with - to exclude)"`
		ReviewStatus string   `help:"Filter by review status (approved,pending,unapproved,changes-requested,all)" enum:"approved,pending,unapproved,changes-requested,all" default:"all"`
		Labels       []string `help:"Filter by labels (prefix with - to exclude)"`
		Reviewer     string   `help:"Filter by assigned reviewer"`
		DraftStatus  string   `help:"Filter by draft status (draft,ready,all)" enum:"draft,ready,all" default:"all"`
		Size         string   `help:"Filter by PR size (small,medium,large,all)" enum:"small,medium,large,all" default:"all"`
		Mergeable    string   `help:"Filter by merge status (mergeable,conflicting,all)" enum:"mergeable,conflicting,all" default:"all"`
		Checks       string   `help:"Filter by CI checks (pass,fail,pending,all)" enum:"pass,fail,pending,all" default:"all"`
		UpdatedSince string   `help:"Filter by last update time (e.g., 24h, 7d)"`
		CreatedSince string   `help:"Filter by creation time (e.g., 24h, 7d)"`
	} `cmd:"" help:"Log PR chains" default:"1"`

	Open struct {
		Output       string   `help:"How to format the output (default,json)" enum:"default,json" default:"default"`
		Filter       string   `arg:"" help:"Number or branch to select chain"`
		Print        bool     `help:"Print URLs instead of opening"`
		Author       string   `help:"Filter by author (prefix with - to exclude)"`
		ReviewStatus string   `help:"Filter by review status (approved,pending,unapproved,changes-requested,all)" enum:"approved,pending,unapproved,changes-requested,all" default:"all"`
		Labels       []string `help:"Filter by labels (prefix with - to exclude)"`
		Reviewer     string   `help:"Filter by assigned reviewer"`
		DraftStatus  string   `help:"Filter by draft status (draft,ready,all)" enum:"draft,ready,all" default:"all"`
		Size         string   `help:"Filter by PR size (small,medium,large,all)" enum:"small,medium,large,all" default:"all"`
		Mergeable    string   `help:"Filter by merge status (mergeable,conflicting,all)" enum:"mergeable,conflicting,all" default:"all"`
		Checks       string   `help:"Filter by CI checks (pass,fail,pending,all)" enum:"pass,fail,pending,all" default:"all"`
		UpdatedSince string   `help:"Filter by last update time (e.g., 24h, 7d)"`
		CreatedSince string   `help:"Filter by creation time (e.g., 24h, 7d)"`
	} `cmd:"" help:"Open specific PR chain"`

	Rebase struct {
		Output string `help:"How to format the output (default,json)" enum:"default,json" default:"default"`
		Filter string `arg:"" help:"Number or branch to select chain"`
		Push   bool   `help:"Push changes to upstream"`
		Args   string `help:"Extra args for pushing" default:"--force-with-lease"`
		Run    bool   `help:"Run the commands instead of printing"`
		Shell  string `help:"Shell for running commands" default:"$SHELL"`
	} `cmd:"" help:"Rebase specific PR chain"`

	Repo      string `help:"Repository to operate on (default: current)"`
	NoCache   bool   `help:"Ignore cache"`
	CacheTime string `help:"Cache duration (e.g., 1m, 5m, 1h)" default:"1m"`
}

func parseRepoURL(url string) (string, string, error) {
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

	return "", "", fmt.Errorf("could not parse repository URL: %s", url)
}

func getOriginURL() (string, error) {
	// Try git first
	url, err := gitconfig.OriginURL()
	if err == nil {
		return url, nil
	}

	// Fall back to jj for non-colocated repos
	out, jjErr := exec.Command("jj", "git", "remote", "list").Output()
	if jjErr != nil {
		return "", fmt.Errorf("unable to read origin url from git (%v) or jj (%v)", err, jjErr)
	}

	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "origin" {
			return fields[1], nil
		}
	}

	return "", fmt.Errorf("no origin remote found in git (%v) or jj", err)
}

func getOrgRepo(arg string) (string, string, error) {
	if len(arg) > 0 {
		splits := strings.Split(arg, "/")
		if len(splits) != 2 {
			return "", "", fmt.Errorf("unknown repo format: %s", arg)
		}

		return splits[0], splits[1], nil
	}

	url, err := getOriginURL()
	if err != nil {
		return "", "", err
	}

	return parseRepoURL(url)
}

func buildFilterOptions(
	author string,
	reviewStatus string,
	labels []string,
	reviewer string,
	draftStatus string,
	size string,
	mergeable string,
	checks string,
	updatedSince string,
	createdSince string,
) FilterOptions {
	return FilterOptions{
		Author:       author,
		ReviewStatus: reviewStatus,
		Labels:       labels,
		Reviewer:     reviewer,
		DraftStatus:  draftStatus,
		Size:         size,
		Mergeable:    mergeable,
		Checks:       checks,
		UpdatedSince: updatedSince,
		CreatedSince: createdSince,
	}
}

func main() {
	ctx := kong.Parse(&CLI)
	cmd := ctx.Command()

	org, repo, err := getOrgRepo(CLI.Repo)
	if err != nil {
		log.Fatal(err)
	}

	var cacheTime time.Duration
	if !CLI.NoCache {
		var err error
		cacheTime, err = time.ParseDuration(CLI.CacheTime)
		if err != nil {
			log.Fatalf("Invalid cache time format '%s': %v", CLI.CacheTime, err)
		}
	}

	data, err := getData(context.Background(), org, repo, !CLI.NoCache, cacheTime)
	if err != nil {
		log.Fatal(err)
	}

	if len(data.prs) == 0 {
		fmt.Fprintln(os.Stderr, "No PRs and therefore no chains")
		os.Exit(1)
	}

	switch cmd {
	case "log":
		opts := buildFilterOptions(
			CLI.Log.Author,
			CLI.Log.ReviewStatus,
			CLI.Log.Labels,
			CLI.Log.Reviewer,
			CLI.Log.DraftStatus,

			CLI.Log.Size,
			CLI.Log.Mergeable,
			CLI.Log.Checks,
			CLI.Log.UpdatedSince,
			CLI.Log.CreatedSince,
		)
		logChains(data, CLI.Log.All, opts)
	case "open <filter>":
		opts := buildFilterOptions(
			CLI.Open.Author,
			CLI.Open.ReviewStatus,
			CLI.Open.Labels,
			CLI.Open.Reviewer,
			CLI.Open.DraftStatus,

			CLI.Open.Size,
			CLI.Open.Mergeable,
			CLI.Open.Checks,
			CLI.Open.UpdatedSince,
			CLI.Open.CreatedSince,
		)
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
