package main

import (
	"testing"
	"time"
)

func TestApplyPRFilters_Author(t *testing.T) {
	p := pr{number: 1, author: "alice"}

	if !ApplyPRFilters(p, FilterOptions{}) {
		t.Error("empty filter should match")
	}
	if !ApplyPRFilters(p, FilterOptions{Author: "alice"}) {
		t.Error("matching author should pass")
	}
	if ApplyPRFilters(p, FilterOptions{Author: "bob"}) {
		t.Error("non-matching author should fail")
	}
}

func TestApplyPRFilters_ReviewStatus(t *testing.T) {
	approved := pr{number: 1, approvedBy: "bob"}
	pending := pr{number: 2}
	changesReq := pr{number: 3, hasChangesRequested: true}

	tests := []struct {
		name   string
		pr     pr
		status string
		want   bool
	}{
		{"approved matches approved", approved, "approved", true},
		{"pending fails approved", pending, "approved", false},
		{"pending matches pending", pending, "pending", true},
		{"approved fails pending", approved, "pending", false},
		{"pending matches unapproved", pending, "unapproved", true},
		{"approved fails unapproved", approved, "unapproved", false},
		{"changes-requested matches", changesReq, "changes-requested", true},
		{"no changes-requested fails", pending, "changes-requested", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyPRFilters(tt.pr, FilterOptions{ReviewStatus: tt.status})
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyPRFilters_Labels(t *testing.T) {
	p := pr{number: 1, labels: []string{"bug", "urgent"}}
	noLabels := pr{number: 2, labels: []string{}}

	tests := []struct {
		name   string
		pr     pr
		labels []string
		want   bool
	}{
		{"matching label", p, []string{"bug"}, true},
		{"other matching label", p, []string{"urgent"}, true},
		{"any match suffices", p, []string{"feature", "bug"}, true},
		{"no match", p, []string{"feature"}, false},
		{"no labels on PR", noLabels, []string{"bug"}, false},
		{"empty filter", p, []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyPRFilters(tt.pr, FilterOptions{Labels: tt.labels})
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyPRFilters_Reviewer(t *testing.T) {
	p := pr{number: 1, reviewers: []string{"alice", "bob"}}
	noReviewers := pr{number: 2, reviewers: []string{}}

	tests := []struct {
		name     string
		pr       pr
		reviewer string
		want     bool
	}{
		{"matching reviewer", p, "alice", true},
		{"other matching reviewer", p, "bob", true},
		{"non-matching reviewer", p, "charlie", false},
		{"no reviewers on PR", noReviewers, "alice", false},
		{"empty filter", p, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyPRFilters(tt.pr, FilterOptions{Reviewer: tt.reviewer})
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyPRFilters_DraftStatus(t *testing.T) {
	draft := pr{number: 1, isDraft: true}
	ready := pr{number: 2, isDraft: false}

	tests := []struct {
		name   string
		pr     pr
		status string
		want   bool
	}{
		{"draft matches draft", draft, "draft", true},
		{"ready fails draft", ready, "draft", false},
		{"ready matches ready", ready, "ready", true},
		{"draft fails ready", draft, "ready", false},
		{"all matches draft", draft, "all", true},
		{"all matches ready", ready, "all", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyPRFilters(tt.pr, FilterOptions{DraftStatus: tt.status})
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyPRFilters_CreatedSince(t *testing.T) {
	recent := pr{number: 1, createdAt: time.Now().Add(-1 * time.Hour)}
	old := pr{number: 2, createdAt: time.Now().Add(-48 * time.Hour)}

	tests := []struct {
		name string
		pr   pr
		dur  string
		want bool
	}{
		{"recent within 24h", recent, "24h", true},
		{"old outside 24h", old, "24h", false},
		{"old within 7d", old, "7d", true},
		{"empty filter", old, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyPRFilters(tt.pr, FilterOptions{CreatedSince: tt.dur})
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyPRFilters_Size(t *testing.T) {
	small := pr{number: 1, additions: 30, deletions: 20}   // 50
	medium := pr{number: 2, additions: 200, deletions: 100} // 300
	large := pr{number: 3, additions: 400, deletions: 200}  // 600

	tests := []struct {
		name string
		pr   pr
		size string
		want bool
	}{
		{"small matches small", small, "small", true},
		{"medium fails small", medium, "small", false},
		{"medium matches medium", medium, "medium", true},
		{"small fails medium", small, "medium", false},
		{"large fails medium", large, "medium", false},
		{"large matches large", large, "large", true},
		{"medium fails large", medium, "large", false},
		{"all matches any", large, "all", true},
		{"empty matches any", small, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyPRFilters(tt.pr, FilterOptions{Size: tt.size})
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyPRFilters_Combined(t *testing.T) {
	p := pr{
		number:    1,
		author:    "alice",
		reviewers: []string{"bob"},
		labels:    []string{"bug"},
		isDraft:   false,
		createdAt: time.Now().Add(-1 * time.Hour),
		additions: 30,
		deletions: 20,
	}

	// All filters match
	if !ApplyPRFilters(p, FilterOptions{
		Author:   "alice",
		Reviewer: "bob",
		Labels:   []string{"bug"},
		Size:         "small",
		CreatedSince: "24h",
	}) {
		t.Error("all matching filters should pass")
	}

	// One filter fails
	if ApplyPRFilters(p, FilterOptions{
		Author:   "alice",
		Reviewer: "charlie",
	}) {
		t.Error("should fail when reviewer doesn't match")
	}
}

func TestFilterPRNumbers(t *testing.T) {
	d := data{
		prs: map[int]pr{
			1: {number: 1, author: "alice", reviewers: []string{"bob"}},
			2: {number: 2, author: "bob", reviewers: []string{"alice"}},
			3: {number: 3, author: "alice", reviewers: []string{"charlie"}},
		},
	}

	tests := []struct {
		name string
		nums []int
		opts FilterOptions
		want []int
	}{
		{
			"filter by author",
			[]int{1, 2, 3},
			FilterOptions{Author: "alice"},
			[]int{1, 3},
		},
		{
			"filter by reviewer",
			[]int{1, 2, 3},
			FilterOptions{Reviewer: "bob"},
			[]int{1},
		},
		{
			"no match",
			[]int{1, 2, 3},
			FilterOptions{Reviewer: "nobody"},
			[]int{},
		},
		{
			"empty filter returns all",
			[]int{1, 2, 3},
			FilterOptions{},
			[]int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterPRNumbers(d, tt.nums, tt.opts)
			if len(got) != len(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
				return
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("got %v, want %v", got, tt.want)
					return
				}
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
		err   bool
	}{
		{"24h", 24 * time.Hour, false},
		{"1d", 24 * time.Hour, false},
		{"7d", 7 * 24 * time.Hour, false},
		{"2w", 14 * 24 * time.Hour, false},
		{"30m", 30 * time.Minute, false},
		{"bad", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseDuration(tt.input)
			if tt.err && err == nil {
				t.Error("expected error")
			}
			if !tt.err && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.err && got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
