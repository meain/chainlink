package main

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
									Author struct {
										Login string `json:"login"`
									} `json:"author"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"reviews"`
						Labels struct {
							Edges []struct {
								Node struct {
									Name  string `json:"name"`
									Color string `json:"color"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"labels"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"pullRequests"`
		} `json:"repository"`
	} `json:"data"`
}
