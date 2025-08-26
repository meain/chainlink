package main

import "time"

type Response struct {
	Data struct {
		Repository struct {
			URL              string `json:"url"`
			DefaultBranchRef struct {
				Name string `json:"name"`
			} `json:"defaultBranchRef"`
			PullRequests struct {
				Edges []struct {
					Node struct {
						Title  string `json:"title"`
						Number int    `json:"number"`
						Author struct {
							Login string `json:"login"`
						} `json:"author"`
						HeadRefName string `json:"headRefName"`
						BaseRefName string `json:"baseRefName"`
						Reviews     struct {
							Edges []struct {
								Node struct {
									State  string `json:"state"`
									Author struct {
										Login string `json:"login"`
									} `json:"author"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"reviews"`
						Labels struct {
							Nodes []struct {
								Name string `json:"name"`
							} `json:"nodes"`
						} `json:"labels"`
						IsDraft        bool   `json:"isDraft"`
						CreatedAt      string `json:"createdAt"`
						ReviewRequests struct {
							Nodes []struct {
								RequestedReviewer struct {
									Login string `json:"login"`
								} `json:"requestedReviewer"`
							} `json:"nodes"`
						} `json:"reviewRequests"`
						Additions int `json:"additions"`
						Deletions int `json:"deletions"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"pullRequests"`
		} `json:"repository"`
	} `json:"data"`
	Errors []struct {
		Type      string   `json:"type"`
		Path      []string `json:"path"`
		Locations []struct {
			Line   int `json:"line"`
			Column int `json:"column"`
		} `json:"locations"`
		Message string `json:"message"`
	} `json:"errors"`
}

type JSONPullRequest struct {
	Number              int       `json:"number"`
	Base                string    `json:"base"`
	Head                string    `json:"head"`
	Title               string    `json:"title"`
	Author              string    `json:"author"`
	ApprovedBy          string    `json:"approvedBy"`
	HasChangesRequested bool      `json:"hasChangesRequested"`
	HasComments         bool      `json:"hasComments"`
	Labels              []string  `json:"labels"`
	IsDraft             bool      `json:"isDraft"`
	CreatedAt           time.Time `json:"createdAt"`
	Reviewers           []string  `json:"reviewers"`
	Additions           int       `json:"additions"`
	Deletions           int       `json:"deletions"`
	URL                 string    `json:"url"`
}

type JSONChain struct {
	PullRequest JSONPullRequest `json:"pullRequest"`
	Children    []JSONChain     `json:"children"`
}

type JSONOutput struct {
	Chains []JSONChain `json:"chains"`
}

type JSONRebaseOutput struct {
	Script   string   `json:"script"`
	Commands []string `json:"commands"`
}
