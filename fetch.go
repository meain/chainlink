package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const githubURL = "https://api.github.com/graphql"

//go:embed request.graphql
var request string

type pr struct {
	number              int
	base                string
	head                string
	title               string
	author              string
	approvedBy          string
	hasChangesRequested bool
	hasComments         bool
	labels              []string
	isDraft             bool
	createdAt           time.Time
	updatedAt           time.Time
	mergeable           string
	checksState         string
	reviewers           []string
	additions           int
	deletions           int
}

type mapping struct {
	base      int
	following []int
}

type data struct {
	url           string
	defaultBranch string
	prs           map[int]pr
	branch        map[string]int
	mappings      map[int]mapping
}

const CACHE_DIR_BASE = "/tmp/chainlink" // TODO: make cross platform

func getToken() (string, error) {
	token := os.Getenv("CHAINLINK_TOKEN")
	if len(token) > 0 {
		return token, nil
	}

	return "", fmt.Errorf("missing GitHub token in CHAINLINK_TOKEN")
}

func cacheFilePath(org, repo string) string {
	return fmt.Sprintf("%s/%s/%s", CACHE_DIR_BASE, org, repo)
}

func readCache(org, repo string, cacheTime time.Duration) ([]byte, bool) {
	cacheFile := cacheFilePath(org, repo)
	st, err := os.Stat(cacheFile)
	if os.IsNotExist(err) {
		return nil, false
	}
	if time.Since(st.ModTime()) >= cacheTime {
		return nil, false
	}
	bts, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, false
	}
	return bts, true
}

func writeCache(org, repo string, bts []byte) {
	cacheFile := cacheFilePath(org, repo)
	err := os.MkdirAll(filepath.Dir(cacheFile), os.ModePerm)
	if err != nil {
		log.Print("Unable to create cache data dir", err)
		return
	}
	err = os.WriteFile(cacheFile, bts, 0644)
	if err != nil {
		log.Print("Unable to cache data", err)
	}
}

func fetchData(org, repo string) ([]byte, error) {
	resp, err := makeRequest(org, repo)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	bts := buf.Bytes()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(bts))
	}

	return bts, nil
}

func makeRequest(org string, repo string) (*http.Response, error) {
	fmt.Fprintf(os.Stderr, "Fetching data for %s/%s...\r", org, repo)
	defer func() { fmt.Fprint(os.Stderr, "\x1b[2K") }()

	gql := fmt.Sprintf(request, org, repo)
	body := fmt.Sprintf(`{"query": "%s"}`, strings.ReplaceAll(strings.ReplaceAll(gql, `"`, `\"`), "\n", "\\n"))
	bodyReader := bytes.NewReader([]byte(body))

	req, err := http.NewRequest(http.MethodPost, githubURL, bodyReader)
	if err != nil {
		return nil, err
	}

	token, err := getToken()
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "bearer "+token)

	client := &http.Client{}
	return client.Do(req)
}

func getData(ctx context.Context, org, repo string, cache bool, cacheTime time.Duration) (data, error) {
	d := data{
		prs:      map[int]pr{},
		branch:   map[string]int{},
		mappings: map[int]mapping{},
	}

	var response []byte
	var fromCache bool
	if cache {
		response, fromCache = readCache(org, repo, cacheTime)
	}
	if !fromCache {
		var err error
		response, err = fetchData(org, repo)
		if err != nil {
			return d, err
		}
	}

	resp := Response{}
	err := json.Unmarshal(response, &resp)
	if err != nil {
		return d, fmt.Errorf("unable to marshal response: %v", err)
	}

	if len(resp.Errors) > 0 {
		for _, e := range resp.Errors {
			fmt.Fprintln(os.Stderr, e.Message)
		}
		return d, fmt.Errorf("unable to fetch PRs")
	}

	// Only cache after confirming the response is valid
	if !fromCache {
		writeCache(org, repo, response)
	}

	d.defaultBranch = resp.Data.Repository.DefaultBranchRef.Name
	d.url = resp.Data.Repository.URL
	d.branch[resp.Data.Repository.DefaultBranchRef.Name] = 0

	for _, p := range resp.Data.Repository.PullRequests.Edges {
		n := p.Node

		aby := ""
		hasChangesRequested := false
		hasComments := false

		// Track latest review state per user (last review wins)
		latestReview := make(map[string]string)
		for _, review := range n.Reviews.Edges {
			login := review.Node.Author.Login
			if login != "" {
				latestReview[login] = review.Node.State
			}
		}
		for login, state := range latestReview {
			switch state {
			case "APPROVED":
				if aby == "" {
					aby = login
				}
			case "CHANGES_REQUESTED":
				hasChangesRequested = true
			case "COMMENTED":
				hasComments = true
			}
		}

		labels := make([]string, 0)
		for _, label := range n.Labels.Nodes {
			labels = append(labels, label.Name)
		}

		reviewers := make([]string, 0)
		reviewerSeen := make(map[string]bool)
		// Include users who have already submitted reviews
		for _, review := range n.Reviews.Edges {
			login := review.Node.Author.Login
			if login != "" && !reviewerSeen[login] {
				reviewers = append(reviewers, login)
				reviewerSeen[login] = true
			}
		}
		// Include pending review requests
		for _, req := range n.ReviewRequests.Nodes {
			login := req.RequestedReviewer.Login
			if login != "" && !reviewerSeen[login] {
				reviewers = append(reviewers, login)
				reviewerSeen[login] = true
			}
		}

		createdAt, _ := time.Parse(time.RFC3339, n.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, n.UpdatedAt)

		checksState := ""
		if len(n.Commits.Nodes) > 0 && n.Commits.Nodes[0].Commit.StatusCheckRollup != nil {
			checksState = strings.ToLower(n.Commits.Nodes[0].Commit.StatusCheckRollup.State)
		}

		mpr := pr{
			number:              n.Number,
			head:                n.HeadRefName,
			base:                n.BaseRefName,
			author:              n.Author.Login,
			title:               n.Title,
			approvedBy:          aby,
			hasChangesRequested: hasChangesRequested,
			hasComments:         hasComments,
			labels:              labels,
			isDraft:             n.IsDraft,
			createdAt:           createdAt,
			updatedAt:           updatedAt,
			mergeable:           strings.ToLower(n.Mergeable),
			checksState:         checksState,
			reviewers:           reviewers,
			additions:           n.Additions,
			deletions:           n.Deletions,
		}

		d.prs[n.Number] = mpr
		d.branch[n.HeadRefName] = n.Number
	}

	// Register base branches that aren't already tracked (e.g. from
	// merged PRs or long-lived branches like develop/release/*).
	// These are treated as roots, same as the default branch.
	// When we hit the 100 PR pagination limit, a missing base may
	// indicate a truncated chain, so warn in that case.
	prCount := len(resp.Data.Repository.PullRequests.Edges)
	for _, p := range resp.Data.Repository.PullRequests.Edges {
		base := p.Node.BaseRefName
		if _, ok := d.branch[base]; !ok {
			if prCount >= 100 {
				fmt.Fprintf(os.Stderr, "base branch %q missing for #%d, possibly due to pagination limit\n", base, p.Node.Number)
			}
			d.branch[base] = 0
		}
	}

	for _, p := range resp.Data.Repository.PullRequests.Edges {
		n := p.Node
		id := n.Number
		base := n.BaseRefName

		following := []int{}
		if len(d.mappings[id].following) > 0 {
			following = d.mappings[id].following
		}

		d.mappings[id] = mapping{
			base:      d.branch[base],
			following: following,
		}

		fl := d.branch[base]

		following = []int{id}
		if len(d.mappings[fl].following) > 0 {
			following = append(d.mappings[fl].following, following...)
		}
		d.mappings[fl] = mapping{
			base:      d.mappings[fl].base,
			following: following,
		}
	}

	return d, nil
}
