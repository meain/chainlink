package main

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

func parseDuration(s string) (time.Duration, error) {
	// First try standard duration parsing
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	// Parse custom formats
	re := regexp.MustCompile(`^(\d+)(d|w)$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("invalid duration format")
	}

	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}

	switch matches[2] {
	case "d":
		return time.Duration(num) * 24 * time.Hour, nil
	case "w":
		return time.Duration(num) * 7 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown duration unit")
	}
}

// FilterOptions contains all possible filter parameters
type FilterOptions struct {
	Author       string
	ReviewStatus string
	Labels       []string
	Reviewer     string
	DraftStatus  string // "draft", "ready", "all"
	Age          string // "24h", "7d", etc.
	Size         string // "small", "medium", "large", "all"
}

// ApplyPRFilters filters a PR based on the given options
func ApplyPRFilters(pr pr, opts FilterOptions) bool {
	// Apply author filter
	if len(opts.Author) != 0 && pr.author != opts.Author {
		return false
	}

	// Apply review status filter
	switch opts.ReviewStatus {
	case "approved":
		if len(pr.approvedBy) == 0 {
			return false
		}
	case "pending":
		if len(pr.approvedBy) > 0 {
			return false
		}
	case "unapproved":
		if len(pr.approvedBy) > 0 {
			return false
		}
	case "changes-requested":
		if !pr.hasChangesRequested {
			return false
		}
	}

	// Apply labels filter
	if len(opts.Labels) > 0 {
		hasLabel := false
		for _, wantLabel := range opts.Labels {
			for _, label := range pr.labels {
				if label == wantLabel {
					hasLabel = true
					break
				}
			}
			if hasLabel {
				break
			}
		}
		if !hasLabel {
			return false
		}
	}

	// Apply reviewer filter
	if len(opts.Reviewer) > 0 {
		hasReviewer := false
		for _, reviewer := range pr.reviewers {
			if reviewer == opts.Reviewer {
				hasReviewer = true
				break
			}
		}
		if !hasReviewer {
			return false
		}
	}

	// Apply draft status filter
	switch opts.DraftStatus {
	case "draft":
		if !pr.isDraft {
			return false
		}
	case "ready":
		if pr.isDraft {
			return false
		}
	}

	// Apply age filter
	if len(opts.Age) > 0 {
		duration, err := parseDuration(opts.Age)
		if err == nil {
			if time.Since(pr.createdAt) > duration {
				return false
			}
		}
	}

	// Apply size filter
	switch opts.Size {
	case "small":
		if pr.additions+pr.deletions > 100 {
			return false
		}
	case "medium":
		if pr.additions+pr.deletions <= 100 || pr.additions+pr.deletions > 500 {
			return false
		}
	case "large":
		if pr.additions+pr.deletions <= 500 {
			return false
		}
	}

	return true
}

// FilterPRNumbers filters a slice of PR numbers based on the given options
func FilterPRNumbers(d data, prNumbers []int, opts FilterOptions) []int {
	filtered := []int{}
	for _, num := range prNumbers {
		if ApplyPRFilters(d.prs[num], opts) {
			filtered = append(filtered, num)
		}
	}
	return filtered
}
