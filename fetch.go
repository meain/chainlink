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
	number     int
	base       string
	head       string
	title      string
	author     string
	approvedBy string
	// TODO: graphql query contains labels, keep them?
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
const CACHE_INTERVAL = 1 * time.Minute

func getToken() (string, error) {
	token := os.Getenv("CHAINLINK_TOKEN")
	if len(token) > 0 {
		return token, nil
	}

	return "", fmt.Errorf("missing GitHub token in CHAINLINK_TOKEN")
}

func fetchData(org, repo string, cache bool) ([]byte, error) {
	cacheFile := fmt.Sprintf("%s/%s/%s", CACHE_DIR_BASE, org, repo)

	if cache {
		if st, err := os.Stat(cacheFile); !os.IsNotExist(err) {
			if time.Now().Sub(st.ModTime()) < CACHE_INTERVAL {
				bts, err := os.ReadFile(cacheFile)
				if err == nil {
					return bts, nil
				}
			}
		}
	}

	resp, err := makeRequest(org, repo)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// Read the response body
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	bts := buf.Bytes()

	err = os.MkdirAll(filepath.Dir(cacheFile), os.ModePerm)
	if err != nil {
		log.Print("Unable to create cache data dir", err)
		return bts, nil
	}

	err = os.WriteFile(cacheFile, bts, 0644)
	if err != nil {
		log.Print("Unable to cache data", err)
		return bts, nil
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

func getData(ctx context.Context, org, repo string, cache bool) (data, error) {
	d := data{
		prs:      map[int]pr{},
		branch:   map[string]int{},
		mappings: map[int]mapping{},
	}

	response, err := fetchData(org, repo, cache)
	if err != nil {
		return d, err
	}

	resp := Response{}
	err = json.Unmarshal(response, &resp)
	if err != nil {
		return d, fmt.Errorf("unable to marshal response: %v", err)
	}

	if len(resp.Errors) > 0 {
		for _, e := range resp.Errors {
			fmt.Println(e.Message)
		}
		return d, fmt.Errorf("unable to fetch PRs")
	}

	d.defaultBranch = resp.Data.Repository.DefaultBranchRef.Name
	d.url = resp.Data.Repository.URL
	d.branch[resp.Data.Repository.DefaultBranchRef.Name] = 0

	for _, p := range resp.Data.Repository.PullRequests.Edges {
		n := p.Node

		aby := "" // We filter by approved reviews in graphql query
		if len(n.Reviews.Edges) > 0 {
			aby = n.Reviews.Edges[0].Node.Author.Login
		}

		mpr := pr{
			number:     n.Number,
			head:       n.HeadRefName,
			base:       n.BaseRefName,
			author:     n.Author.Login,
			title:      n.Title,
			approvedBy: aby,
		}

		d.prs[n.Number] = mpr
		d.branch[n.HeadRefName] = n.Number
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

		fl, ok := d.branch[base]
		if !ok {
			// This can happen because we only fetch so many PRs
			// Limited to 100 due to GH limitation for a single page
			fmt.Fprintf(os.Stderr, "base missing for %d, using %s\n", id, d.defaultBranch)
		}

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
