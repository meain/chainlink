package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"github.com/tcnksm/go-gitconfig"
	"golang.org/x/oauth2"
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
	// check for api rate limit error
	if strings.Contains(err.Error(), "403 API rate limit exceeded") {
		fmt.Println("Error: API rate limit exceeded")
		fmt.Println("You can increase the limit by setting the GITHUB_TOKEN env variable")
		fmt.Println("See https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %s\n\n", err)
	}
	sps := strings.Split(os.Args[0], "/")
	fmt.Printf(`Usage: %s [flags] <action> [filter:branch|number]

Actions:
log  - log all chains in repo
open - open all links in a chain
rebase - give command to rebase all prs in chain

You can provider branch name or pr number to filter the chains.

Flags:
`,
		sps[len(sps)-1])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	or := flag.String("repo", "", "repo to work on (eg: meain/chainlink)")
	om := flag.String("output", "plain", "format for output (options: plain, fancy)")
	noPush := flag.Bool("no-push", false, "push rebased branches")
	flag.Parse()

	args := flag.Args()

	push := !*noPush

	action := "log"
	if len(args) > 0 {
		action = args[0]
	}

	if action == "help" {
		// doing it early to avoid pulling prs
		printHelp(nil)
	}

	if strings.ToLower(*om) != "plain" && strings.ToLower(*om) != "fancy" {
		printHelp(fmt.Errorf("invalid output mode %s", *om))
	}

	org, repo, err := parseOrgAndRepo(*or)
	if err != nil {
		printHelp(err)
	}

	var tc *http.Client

	if os.Getenv("GITHUB_TOKEN") != "" {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
		)
		tc = oauth2.NewClient(ctx, ts)
	}

	client := github.NewClient(tc)

	// get the name of the base branch from gh api
	r, _, err := client.Repositories.Get(context.Background(), org, repo)
	if err != nil {
		printHelp(fmt.Errorf("unable to get repo: %s", err))
	}

	base := *r.DefaultBranch
	ctx := context.Background()

	prs, err := getPRs(ctx, org, repo, client)
	if err != nil {
		printHelp(err)
	}

	if action == "rebase" && len(args) < 2 {
		printHelp(fmt.Errorf("rebase action requires a filter"))
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
	switch action {
	case "log":
		printChains(ctx, client, base, basePRMap, *om)
	case "rebase":
		rebaseChains(ctx, base, basePRMap, push)
	case "open":
		printChains(ctx, client, base, basePRMap, *om)
		openChainsLinks(base, basePRMap)
	default:
		printHelp(fmt.Errorf("unknown action %s", action))
	}
}

// TODO: handle pagination
func getPRs(ctx context.Context, org, repo string, client *github.Client) ([]*github.PullRequest, error) {
	prs, _, err := client.PullRequests.List(ctx, org, repo, nil)
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

func rebaseChains(
	ctx context.Context,
	base string,
	basePRMap map[string][]*github.PullRequest,
	push bool,
) {
	if len(basePRMap[base]) != 0 {
		fmt.Println("set -ex")
	}
	rebaseChainsInner(ctx, base, basePRMap, push)
}

func rebaseChainsInner(
	ctx context.Context,
	base string,
	basePRMap map[string][]*github.PullRequest,
	push bool,
) {
	for _, pr := range basePRMap[base] {
		fmt.Printf("git checkout %s\n", *pr.Head.Ref)
		fmt.Printf("git rebase --update-refs %s\n", *pr.Base.Ref)

		if push {
			fmt.Printf("git push --force-with-lease\n")
		}
		rebaseChainsInner(ctx, *pr.Head.Ref, basePRMap, push)
	}
}

func printChains(
	ctx context.Context,
	client *github.Client,
	base string,
	basePRMap map[string][]*github.PullRequest,
	outputMode string,
) {
	printChainLevel(ctx, client, base, 0, basePRMap, outputMode)
}

func printChainLevel(
	ctx context.Context,
	client *github.Client,
	base string,
	level int,
	basePRMap map[string][]*github.PullRequest,
	outputMode string,
) {
	for _, pr := range basePRMap[base] {
		approvedBy, err := reviewers(ctx, client, pr)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(formatPR(outputMode, level, pr, approvedBy))
		printChainLevel(ctx, client, *pr.Head.Ref, level+1, basePRMap, outputMode)

		if level == 0 {
			fmt.Println()
		}
	}
}

func reviewers(ctx context.Context, client *github.Client, pr *github.PullRequest) ([]string, error) {
	reviews, _, err := client.PullRequests.ListReviews(
		ctx,
		pr.Base.Repo.Owner.GetLogin(),
		pr.Base.Repo.GetName(),
		*pr.Number,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("unable to get reviews: %s", err)
	}

	// TODO: add review requested from names if not approved
	approvedBy := []string{}
	for _, review := range reviews {
		if review.GetState() == "APPROVED" {
			approvedBy = append(approvedBy, review.GetUser().GetLogin())
		}
	}

	return approvedBy, nil
}

func formatPR(mode string, level int, pr *github.PullRequest, approvedBy []string) string {
	switch mode {
	case "fancy":
		approval := "🔃"
		if len(approvedBy) > 0 {
			approval = "✅ " + strings.Join(approvedBy, ", ")
		}

		return fmt.Sprintf(
			"%s[#%s] %s (%s)\n%s 🙇 %s 🔗 %s %s\n",
			strings.Repeat("\t", level),
			strconv.Itoa(*pr.Number),
			*pr.Title,
			*pr.Head.Ref,
			strings.Repeat("\t", level),
			*pr.User.Login,
			*pr.State,
			approval,
		)
	default:
		approval := "not approved"
		if len(approvedBy) > 0 {
			approval = "approved"
		}

		return fmt.Sprintf(
			"%s[#%s] %s (%s) <%s> [%s:%s]",
			strings.Repeat("\t", level),
			strconv.Itoa(*pr.Number),
			*pr.Title,
			*pr.Head.Ref,
			*pr.User.Login,
			*pr.State,
			approval,
		)
	}
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
