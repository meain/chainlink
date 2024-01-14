package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
)

//go:embed request.graphql
var request string

//go:embed response
var response string

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

func getData(ctx context.Context, org, repo string) (data, error) {
	d := data{
		prs:      map[int]pr{},
		branch:   map[string]int{},
		mappings: map[int]mapping{},
	}

	// TODO: Switch to doing a graphql query
	resp := Response{}
	err := json.Unmarshal([]byte(response), &resp)
	if err != nil {
		return d, fmt.Errorf("unable to marshal response: %e", err)
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
			log.Print(fmt.Printf("base missing for %d, using %s", id, d.defaultBranch))
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
