package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"github.com/tcnksm/go-gitconfig"
)

func parseOrgAndRepo(repo string) (string, string, error) {
	if repo == "" {
		url, err := gitconfig.OriginURL()
		if err != nil {
			return "", "", fmt.Errorf("unable to read gitconfig: %s", err)
		}

		if strings.HasPrefix(url, "git@") {
			url = strings.TrimSuffix(url, ".git")
			splits := strings.Split(url, ":")
			if len(splits) != 2 {
				return "", "", errors.New("invalid repo url format")
			}

			splits = strings.Split(splits[1], "/")

			return splits[0], splits[1], nil
		}

		if strings.HasPrefix(url, "https://") {
			url = strings.TrimSuffix(url, ".git")
			splits := strings.Split(url, "/")
			if len(splits) != 5 {
				return "", "", errors.New("invalid repo url format")
			}

			return splits[3], splits[4], nil
		}
	}

	splits := strings.Split(repo, "/")
	if len(splits) != 2 {
		return "", "", errors.New("invalid repo format")
	}

	return splits[0], splits[1], nil
}

func printHelp(err error) {
	if err != nil {
		fmt.Printf("Error: %s\n\n", err)
	}
	sps := strings.Split(os.Args[0], "/")
	fmt.Printf(`Usage: %s [flags] <action> [filter:branch|number]

Actions:
log  - log all chains in repo
open - open all links in a chain

You can provider branch name or pr number to filter the chains.

Flags:
`,
		sps[len(sps)-1])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	or := flag.String("repo", "", "repo to work on (eg: meain/chainlink)")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 && len(args) != 2 {
		printHelp(nil)
	}

	org, repo, err := parseOrgAndRepo(*or)
	if err != nil {
		printHelp(err)
	}

	client := github.NewClient(nil)

	// get the name of the base branch from gh api
	r, _, err := client.Repositories.Get(context.Background(), org, repo)
	if err != nil {
		printHelp(fmt.Errorf("unable to get repo: %s", err))
	}

	base := *r.DefaultBranch

	prs, err := getPRs(org, repo, client)
	if err != nil {
		printHelp(err)
	}

	filter := ""
	if len(args) == 2 {
		filter = args[1]
	}

	basePRMap := getBasePRMap(base, prs)

	if filter != "" {
		basePRMap = filterPRs(base, filter, basePRMap)
	}

	// TODO: add auto rebasing for all items in chain
	action := args[0]
	switch action {
	case "log":
		printChains(base, basePRMap)
	case "open":
		printChains(base, basePRMap)
		openChainsLinks(base, basePRMap)
	default:
		printHelp(fmt.Errorf("unknown action %s", action))
	}
}

// TODO: handle pagination
func getPRs(org, repo string, client *github.Client) ([]*github.PullRequest, error) {
	prs, _, err := client.PullRequests.List(context.Background(), org, repo, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to get pull requests: %s", err)
	}

	return prs, nil

}

func filterPRs(
	base, filter string,
	basePRMap map[string][]*github.PullRequest,
) map[string][]*github.PullRequest {
	if base == filter {
		return basePRMap
	}

	filteredBasePrs := []*github.PullRequest{}
	for _, pr := range basePRMap[base] {
		if *pr.Head.Ref == filter || strconv.Itoa(*pr.Number) == filter {
			filteredBasePrs = append(filteredBasePrs, pr)
			continue
		}

		if hasBranch(*pr.Head.Ref, filter, basePRMap) {
			filteredBasePrs = append(filteredBasePrs, pr)
		}
	}

	// TODO: remove inaccessible prs
	basePRMap[base] = filteredBasePrs

	return basePRMap
}

func hasBranch(start, branch string, basePRMap map[string][]*github.PullRequest) bool {
	prs := basePRMap[start]

	for _, pr := range prs {
		if *pr.Head.Ref == branch {
			return true
		}

		if hasBranch(*pr.Head.Ref, branch, basePRMap) {
			return true
		}
	}

	return false
}

// getBasePRMap returns a map from the base branch of prs to the
// prs. This can be used to find chains.
func getBasePRMap(base string, prs []*github.PullRequest) map[string][]*github.PullRequest {
	basePRMap := map[string][]*github.PullRequest{}
	for _, pr := range prs {
		if pr.Base.Ref == nil || pr.Head.Ref == nil {
			continue
		}

		basePRMap[*pr.Base.Ref] = append(basePRMap[*pr.Base.Ref], pr)
	}

	// filer out branches with only one pr
	filteredBasePRs := []*github.PullRequest{}
	for _, pr := range basePRMap[base] {
		_, ok := basePRMap[*pr.Head.Ref]
		if ok {
			filteredBasePRs = append(filteredBasePRs, pr)
		}
	}

	basePRMap[base] = filteredBasePRs

	return basePRMap
}

func printChains(base string, basePRMap map[string][]*github.PullRequest) {
	printChainLevel(base, 0, basePRMap)
}

func printChainLevel(base string, level int, basePRMap map[string][]*github.PullRequest) {
	for _, pr := range basePRMap[base] {
		fmt.Println(formatPR(level, pr))
		printChainLevel(*pr.Head.Ref, level+1, basePRMap)

		if level == 0 {
			fmt.Println()
		}
	}
}

func formatPR(level int, pr *github.PullRequest) string {
	return fmt.Sprintf(
		"%s[#%s] %s (%s) <%s>",
		strings.Repeat("\t", level),
		strconv.Itoa(*pr.Number),
		*pr.Title,
		*pr.Head.Ref,
		*pr.User.Login,
	)
}

func openChainsLinks(base string, basePRMap map[string][]*github.PullRequest) {
	prs, ok := basePRMap[base]
	if !ok {
		return
	}

	for _, pr := range prs {
		err := openBrowser(*pr.HTMLURL)
		if err != nil {
			printHelp(err)
		}

		openChainsLinks(*pr.Head.Ref, basePRMap)
	}
}
