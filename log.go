package main

import (
	"fmt"
	"hash/fnv"
	"strings"

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

func logChains(d data, all bool) {
	mappings := d.mappings
	if !all {
		mappings = filterChains(d.mappings)
	}

	if len(mappings) == 0 {
		fmt.Println("No PR chains")
		return
	}

	printChildren(d, mappings, 0, 0, all)
}

func printChildren(d data, mappings map[int]mapping, base, level int, all bool) {
	for _, p := range mappings[base].following {
		indent := strings.Repeat("  ", level) // TODO: print a tree like structure
		fmt.Println(indent + formatPR(d.prs[p], d.url))
		printChildren(d, mappings, p, level+1, all)
	}
}
func hashString(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func generateColor(hash uint32) *color.Color {
	// Extract individual components of the hash value
	r := int((hash >> 16) & 0xFF)
	g := int((hash >> 8) & 0xFF)
	b := int(hash & 0xFF)

	return color.New(38, 2, color.Attribute(r), color.Attribute(g), color.Attribute(b))
}

func formatPR(p pr, url string) string {
	authorColor := generateColor(hashString(p.author)).SprintFunc()
	author := authorColor(p.author)

	green := color.New(color.FgGreen).SprintFunc()
	number := fmt.Sprintf("#%d", p.number)

	if len(p.approvedBy) > 0 {
		number = green(number)
	}

	line := fmt.Sprintf(
		"\x1b]8;;%s/pull/%d\x07%s\x1b]8;;\x07 %s (%s) [%s]",
		url,
		p.number,
		number,
		p.title,
		author,
		p.head)

	return line

}
