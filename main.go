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
)

func parseOrgAndRepo(repo string) (string, string, error) {
	// TODO: if repo is empty, read current repo from git config file
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
	fmt.Printf(`Usage: %s [flags] <action> [branch]

Actions:
log  - log all chains in repo (provide branch to limit)
open - open all links in a chain (provide any branch in chain)

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

	prs, err := getPRs(org, repo)
	if err != nil {
		printHelp(err)
	}

	branch := ""
	if len(args) == 2 {
		branch = args[1]
	}

	chains := findChains("main", prs) // TODO: master

	if branch != "" {
		chain := findChainWithBranch(chains, branch)
		chains = map[string][]*github.PullRequest{"": chain}
	}

	// TODO add auto rebasing for all items in chain
	action := args[0]
	switch action {
	case "log":
		printChains(chains)
	case "open":
		openChainsLinks(chains)
	default:
		printHelp(fmt.Errorf("unknown action %s", action))
	}
}

// TODO handle pagination
func getPRs(org, repo string) ([]*github.PullRequest, error) {
	client := github.NewClient(nil)

	prs, _, err := client.PullRequests.List(context.Background(), org, repo, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to get pull requests: %s", err)
	}

	return prs, nil

}

// TODO add option to filter by url or pr number
func findChainWithBranch(chains map[string][]*github.PullRequest, branch string) []*github.PullRequest {
	for _, chain := range chains {
		for _, pr := range chain {
			if *pr.Head.Ref == branch {
				return chain
			}
		}
	}

	return nil
}

func findChains(base string, prs []*github.PullRequest) map[string][]*github.PullRequest {
	basePRMap := map[string]*github.PullRequest{}
	headPRMap := map[string]*github.PullRequest{}
	starts := []string{}
	for _, pr := range prs {
		if pr.Base.Ref == nil || pr.Head.Ref == nil {
			continue
		}

		headPRMap[*pr.Head.Ref] = pr

		if *pr.Base.Ref == base {
			starts = append(starts, *pr.Head.Ref)
		} else {
			basePRMap[*pr.Base.Ref] = pr
		}
	}

	chains := map[string][]*github.PullRequest{}
	for _, start := range starts {
		sp := headPRMap[start]
		chains[start] = []*github.PullRequest{sp}

		var ok bool
		for {
			sp, ok = basePRMap[*sp.Head.Ref]
			if !ok {
				break
			}

			chains[start] = append(chains[start], sp)
		}
	}

	for base, chain := range chains {
		if len(chain) == 1 {
			delete(chains, base)
		}
	}

	return chains
}

func printChains(chains map[string][]*github.PullRequest) {
	for _, chain := range chains {
		for i, pr := range chain {
			fmt.Printf(
				"%s[#%s] %s (%s) <%s>\n",
				strings.Repeat("\t", i),
				strconv.Itoa(*pr.Number),
				*pr.Title,
				*pr.Head.Ref,
				*pr.User.Login,
			)
		}

		fmt.Println()
	}
}

func openChainsLinks(chains map[string][]*github.PullRequest) {
	for _, chain := range chains {
		for _, pr := range chain {
			fmt.Println("Opening", *pr.HTMLURL)
			err := openBrowser(*pr.HTMLURL)
			if err != nil {
				printHelp(err)
			}
		}
	}
}
