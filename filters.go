package main

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
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
	Size         string // "small", "medium", "large", "all"
	Mergeable    string // "mergeable", "conflicting", "all"
	Checks       string // "pass", "fail", "pending", "all"
	UpdatedSince string // "24h", "7d", etc.
	CreatedSince string // "24h", "7d", etc.
}

// ApplyPRFilters filters a PR based on the given options
func ApplyPRFilters(pr pr, opts FilterOptions) bool {
	// Apply author filter (prefix with - to exclude)
	if len(opts.Author) != 0 {
		if strings.HasPrefix(opts.Author, "-") {
			if pr.author == opts.Author[1:] {
				return false
			}
		} else if pr.author != opts.Author {
			return false
		}
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

	// Apply labels filter (prefix with - to exclude)
	if len(opts.Labels) > 0 {
		includeLabels := []string{}
		excludeLabels := []string{}
		for _, l := range opts.Labels {
			if strings.HasPrefix(l, "-") {
				excludeLabels = append(excludeLabels, l[1:])
			} else {
				includeLabels = append(includeLabels, l)
			}
		}

		// Check exclude labels: PR must not have any of these
		for _, el := range excludeLabels {
			if slices.Contains(pr.labels, el) {
				return false
			}
		}

		// Check include labels: PR must have at least one of these
		if len(includeLabels) > 0 {
			hasLabel := false
			for _, wantLabel := range includeLabels {
				if slices.Contains(pr.labels, wantLabel) {
					hasLabel = true
					break
				}
			}
			if !hasLabel {
				return false
			}
		}
	}

	// Apply reviewer filter
	if len(opts.Reviewer) > 0 {
		hasReviewer := slices.Contains(pr.reviewers, opts.Reviewer)
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

	// Apply created-since filter
	if len(opts.CreatedSince) > 0 {
		duration, err := parseDuration(opts.CreatedSince)
		if err == nil {
			if time.Since(pr.createdAt) > duration {
				return false
			}
		}
	}

	// Apply updated-since filter
	if len(opts.UpdatedSince) > 0 {
		duration, err := parseDuration(opts.UpdatedSince)
		if err == nil {
			if time.Since(pr.updatedAt) > duration {
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

	// Apply mergeable filter
	switch opts.Mergeable {
	case "mergeable":
		if pr.mergeable != "mergeable" {
			return false
		}
	case "conflicting":
		if pr.mergeable != "conflicting" {
			return false
		}
	}

	// Apply checks filter
	switch opts.Checks {
	case "pass":
		if pr.checksState != "success" {
			return false
		}
	case "fail":
		if pr.checksState != "failure" && pr.checksState != "error" {
			return false
		}
	case "pending":
		if pr.checksState != "pending" && pr.checksState != "expected" {
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
