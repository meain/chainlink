package main

import (
	"testing"
	"time"
)

// helper to build test data with PRs and mappings
func makeTestData(prs map[int]pr, mappings map[int]mapping) data {
	return data{
		url:           "https://github.com/test/repo",
		defaultBranch: "main",
		prs:           prs,
		branch:        map[string]int{},
		mappings:      mappings,
	}
}

func TestCollectJSONChains_FilteredParentKeepsChildren(t *testing.T) {
	// Chain: 0 -> 1 -> 2
	// PR 1 doesn't match filter, PR 2 does.
	// PR 2 should still appear in output.
	d := makeTestData(
		map[int]pr{
			1: {number: 1, author: "alice", reviewers: []string{}},
			2: {number: 2, author: "bob", reviewers: []string{"charlie"}},
		},
		map[int]mapping{
			0: {following: []int{1}},
			1: {base: 0, following: []int{2}},
			2: {base: 1, following: []int{}},
		},
	)

	opts := FilterOptions{Reviewer: "charlie"}
	output := buildJSONOutput(d, d.mappings, 0, opts)

	if len(output.Chains) != 1 {
		t.Fatalf("expected 1 chain, got %d", len(output.Chains))
	}
	if output.Chains[0].PullRequest.Number != 2 {
		t.Errorf("expected PR #2, got #%d", output.Chains[0].PullRequest.Number)
	}
}

func TestCollectJSONChains_AllMatch(t *testing.T) {
	d := makeTestData(
		map[int]pr{
			1: {number: 1, author: "alice"},
			2: {number: 2, author: "alice"},
		},
		map[int]mapping{
			0: {following: []int{1}},
			1: {base: 0, following: []int{2}},
			2: {base: 1, following: []int{}},
		},
	)

	opts := FilterOptions{Author: "alice"}
	output := buildJSONOutput(d, d.mappings, 0, opts)

	if len(output.Chains) != 1 {
		t.Fatalf("expected 1 chain, got %d", len(output.Chains))
	}
	if output.Chains[0].PullRequest.Number != 1 {
		t.Errorf("expected PR #1 at root, got #%d", output.Chains[0].PullRequest.Number)
	}
	if len(output.Chains[0].Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(output.Chains[0].Children))
	}
	if output.Chains[0].Children[0].PullRequest.Number != 2 {
		t.Errorf("expected PR #2 as child, got #%d", output.Chains[0].Children[0].PullRequest.Number)
	}
}

func TestCollectJSONChains_NoneMatch(t *testing.T) {
	d := makeTestData(
		map[int]pr{
			1: {number: 1, author: "alice"},
			2: {number: 2, author: "alice"},
		},
		map[int]mapping{
			0: {following: []int{1}},
			1: {base: 0, following: []int{2}},
			2: {base: 1, following: []int{}},
		},
	)

	opts := FilterOptions{Author: "nobody"}
	output := buildJSONOutput(d, d.mappings, 0, opts)

	if len(output.Chains) != 0 {
		t.Errorf("expected 0 chains, got %d", len(output.Chains))
	}
}

func TestCollectJSONChains_MultipleChains(t *testing.T) {
	// Two independent chains from root:
	// 0 -> 1 -> 2  (alice)
	// 0 -> 3 -> 4  (bob)
	// Filter by author=alice: only first chain
	d := makeTestData(
		map[int]pr{
			1: {number: 1, author: "alice"},
			2: {number: 2, author: "alice"},
			3: {number: 3, author: "bob"},
			4: {number: 4, author: "bob"},
		},
		map[int]mapping{
			0: {following: []int{1, 3}},
			1: {base: 0, following: []int{2}},
			2: {base: 1, following: []int{}},
			3: {base: 0, following: []int{4}},
			4: {base: 3, following: []int{}},
		},
	)

	opts := FilterOptions{Author: "alice"}
	output := buildJSONOutput(d, d.mappings, 0, opts)

	if len(output.Chains) != 1 {
		t.Fatalf("expected 1 chain, got %d", len(output.Chains))
	}
	if output.Chains[0].PullRequest.Number != 1 {
		t.Errorf("expected PR #1, got #%d", output.Chains[0].PullRequest.Number)
	}
}

func TestCollectJSONChains_DeepChainMiddleFiltered(t *testing.T) {
	// Chain: 0 -> 1 -> 2 -> 3
	// PR 2 doesn't match, PRs 1 and 3 match.
	// Expected: PR 1 at root, PR 3 promoted as child of PR 1.
	d := makeTestData(
		map[int]pr{
			1: {number: 1, author: "alice", labels: []string{"bug"}},
			2: {number: 2, author: "alice", labels: []string{}},
			3: {number: 3, author: "alice", labels: []string{"bug"}},
		},
		map[int]mapping{
			0: {following: []int{1}},
			1: {base: 0, following: []int{2}},
			2: {base: 1, following: []int{3}},
			3: {base: 2, following: []int{}},
		},
	)

	opts := FilterOptions{Labels: []string{"bug"}}
	output := buildJSONOutput(d, d.mappings, 0, opts)

	if len(output.Chains) != 1 {
		t.Fatalf("expected 1 chain, got %d", len(output.Chains))
	}
	if output.Chains[0].PullRequest.Number != 1 {
		t.Errorf("expected PR #1 at root, got #%d", output.Chains[0].PullRequest.Number)
	}
	if len(output.Chains[0].Children) != 1 {
		t.Fatalf("expected 1 child of PR #1, got %d", len(output.Chains[0].Children))
	}
	if output.Chains[0].Children[0].PullRequest.Number != 3 {
		t.Errorf("expected PR #3 as child, got #%d", output.Chains[0].Children[0].PullRequest.Number)
	}
}

func TestFilterChains(t *testing.T) {
	// 0 -> 1 (standalone, no children)
	// 0 -> 2 -> 3 (chain)
	mappings := map[int]mapping{
		0: {following: []int{1, 2}},
		1: {base: 0, following: []int{}},
		2: {base: 0, following: []int{3}},
		3: {base: 2, following: []int{}},
	}

	filtered := filterChains(mappings)

	// Only PR 2 should remain as a top-level entry (it has children)
	if len(filtered[0].following) != 1 {
		t.Fatalf("expected 1 top-level chain, got %d", len(filtered[0].following))
	}
	if filtered[0].following[0] != 2 {
		t.Errorf("expected PR #2, got #%d", filtered[0].following[0])
	}
}

func TestCollectJSONChains_DraftFilter(t *testing.T) {
	// Chain: 0 -> 1(draft) -> 2(ready)
	// Filter draft=ready: PR 1 filtered, PR 2 promoted
	d := makeTestData(
		map[int]pr{
			1: {number: 1, isDraft: true},
			2: {number: 2, isDraft: false},
		},
		map[int]mapping{
			0: {following: []int{1}},
			1: {base: 0, following: []int{2}},
			2: {base: 1, following: []int{}},
		},
	)

	opts := FilterOptions{DraftStatus: "ready"}
	output := buildJSONOutput(d, d.mappings, 0, opts)

	if len(output.Chains) != 1 {
		t.Fatalf("expected 1 chain, got %d", len(output.Chains))
	}
	if output.Chains[0].PullRequest.Number != 2 {
		t.Errorf("expected PR #2, got #%d", output.Chains[0].PullRequest.Number)
	}
}

func TestCollectJSONChains_SizeFilter(t *testing.T) {
	// Chain: 0 -> 1(large) -> 2(small)
	// Filter size=small: PR 1 filtered, PR 2 promoted
	d := makeTestData(
		map[int]pr{
			1: {number: 1, additions: 400, deletions: 200},
			2: {number: 2, additions: 10, deletions: 5},
		},
		map[int]mapping{
			0: {following: []int{1}},
			1: {base: 0, following: []int{2}},
			2: {base: 1, following: []int{}},
		},
	)

	opts := FilterOptions{Size: "small"}
	output := buildJSONOutput(d, d.mappings, 0, opts)

	if len(output.Chains) != 1 {
		t.Fatalf("expected 1 chain, got %d", len(output.Chains))
	}
	if output.Chains[0].PullRequest.Number != 2 {
		t.Errorf("expected PR #2, got #%d", output.Chains[0].PullRequest.Number)
	}
}

func TestCollectJSONChains_AgeFilter(t *testing.T) {
	// Chain: 0 -> 1(old) -> 2(recent)
	// Filter age=24h: PR 1 filtered, PR 2 promoted
	d := makeTestData(
		map[int]pr{
			1: {number: 1, createdAt: time.Now().Add(-72 * time.Hour)},
			2: {number: 2, createdAt: time.Now().Add(-1 * time.Hour)},
		},
		map[int]mapping{
			0: {following: []int{1}},
			1: {base: 0, following: []int{2}},
			2: {base: 1, following: []int{}},
		},
	)

	opts := FilterOptions{Age: "24h"}
	output := buildJSONOutput(d, d.mappings, 0, opts)

	if len(output.Chains) != 1 {
		t.Fatalf("expected 1 chain, got %d", len(output.Chains))
	}
	if output.Chains[0].PullRequest.Number != 2 {
		t.Errorf("expected PR #2, got #%d", output.Chains[0].PullRequest.Number)
	}
}
