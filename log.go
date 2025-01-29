package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

// filterChains filters our PR which aren't chains
func filterChains(m map[int]mapping) map[int]mapping {
	nm := make(map[int]mapping, len(m))
	for k, v := range m {
		nm[k] = v
	}

	items := []int{}
	for _, v := range m[0].following {
		if len(m[v].following) > 0 {
			items = append(items, v)
		}
	}

	nm[0] = mapping{following: items}

	return nm
}

func logChains(d data, all bool, opts FilterOptions) {
	mappings := d.mappings
	if !all {
		mappings = filterChains(d.mappings)
	}

	if len(mappings) == 0 {
		fmt.Println("No PR chains")
		return
	}
	printChildren(d, mappings, 0, 0, all, CLI.Log.Output, opts)
}

func printChildren(
	d data,
	mappings map[int]mapping,
	base, level int,
	all bool,
	output string,
	opts FilterOptions,
) {
	for _, p := range mappings[base].following {
		if !ApplyPRFilters(d.prs[p], opts) {
			continue
		}

		indent := strings.Repeat("  ", level) // TODO: print a tree like structure
		var line string
		switch output {
		case "small":
			line = formatPRSmall(d.prs[p], d.url)
		case "markdown":
			line = formatPRMarkdown(d.prs[p], d.url)
		default:
			line = formatPR(d.prs[p], d.url)
		}
		fmt.Println(indent + line)
		printChildren(d, mappings, p, level+1, all, output, opts)
	}
}

// generateColor generates a random color for any string
// using hsl and converting to rgb as it is easier to make it look
// nicer for random colors
func generateColor(str string) *color.Color {
	authorSum := 0
	for _, c := range str {
		authorSum += int(c)
	}

	authorHash := authorSum % 355
	hsl := HSL{float64(authorHash), 0.5, .5}
	rgb := HSLToRGB(hsl)

	return color.New(38, 2, color.Attribute(rgb.R*255), color.Attribute(rgb.G*255), color.Attribute(rgb.B*255))
}

func formatPRSmall(p pr, url string) string {
	green := color.New(color.FgGreen).SprintFunc()
	number := fmt.Sprintf("#%d", p.number)

	if len(p.approvedBy) > 0 {
		number = green(number)
	}

	line := fmt.Sprintf("%s %s", number, p.title)

	return line
}

func formatPRMarkdown(p pr, url string) string {
	line := fmt.Sprintf(
		"- [#%d](%s/pull/%d) %s",
		p.number,
		url,
		p.number,
		p.title)

	return line
}

func formatPR(p pr, url string) string {
	authorColor := generateColor(p.author).SprintFunc()
	author := authorColor(p.author)

	green := color.New(color.FgGreen).SprintFunc()
	number := fmt.Sprintf("#%d", p.number)

	if len(p.approvedBy) > 0 {
		number = green(number)
	}

	age := time.Since(p.createdAt)
	ageStr := ""
	switch {
	case age < 24*time.Hour:
		ageStr = fmt.Sprintf("%.0fh", age.Hours())
	case age < 30*24*time.Hour:
		ageStr = fmt.Sprintf("%.0fd", age.Hours()/24)
	default:
		ageStr = fmt.Sprintf("%.0fmo", age.Hours()/24/30)
	}

	line := fmt.Sprintf(
		"\x1b]8;;%s/pull/%d\x07%s\x1b]8;;\x07 %s (%s) [%s] %s ago",
		url,
		p.number,
		number,
		p.title,
		author,
		p.head,
		ageStr)

	return line
}
