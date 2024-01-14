package main

import (
	"fmt"
	"strings"
)

// filterChains filters our PR which aren't chains
func filterChains(m map[int]mapping) map[int]mapping {
	nm := map[int]mapping{}

	for _, p := range m[0].following {
		if len(m[p].following) == 0 {
			continue
		}

		nm[p] = m[p]
	}

	mp := []int{}
	for p := range nm {
		mp = append(mp, p)
	}

	nm[0] = mapping{following: mp}

	return nm
}

func logChains(d data, all bool) error {
	mappings := d.mappings
	if !all {
		mappings = filterChains(d.mappings)
	}
	return printChildren(d, mappings, 0, 0, all)
}

func printChildren(d data, mappings map[int]mapping, base, level int, all bool) error {
	for _, p := range mappings[base].following {
		indent := strings.Repeat(" ", level)
		fmt.Println(indent, formatPR(d.prs[p]))
		printChildren(d, mappings, p, level+1, all)
	}
	return nil
}

func formatPR(p pr) string {
	return fmt.Sprintf("#%d", p.number)
}
