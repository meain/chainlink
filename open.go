package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"slices"
	"strconv"
)

func filterByNumber(d data, num int) []int {
	if num == 0 {
		return []int{}
	}

	prns := []int{}

	// items before
	iter := num
	for {
		base := d.mappings[iter].base
		prns = append([]int{base}, prns...)

		if base == 0 {
			break
		}

		iter = base
	}

	// items after
	stack := []int{num}
	for {
		if len(stack) == 0 {
			break
		}

		last := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		prns = append(prns, last)
		following := d.mappings[last].following

		if len(following) > 0 {
			slices.Reverse(following)
			stack = append(stack, following...)
		}
	}

	return prns
}

func filterChain(d data, filter string) []int {
	prns := []int{}

	num, err := strconv.Atoi(filter)
	if err != nil {
		num = d.branch[filter]
		if num == 0 {
			fmt.Printf("No branch found for filter %s\n", filter)
		}
	}

	prns = filterByNumber(d, num)

	if len(prns) == 0 {
		return prns
	}

	// first item will be base(0)
	return prns[1:]
}

func openChain(d data, filter string, print bool) {
	prns := filterChain(d, filter)
	if len(prns) == 0 {
		fmt.Println("No PR chain found with filter")
		return
	}

	for _, p := range prns {
		if print {
			fmt.Println(fmt.Sprintf("%s/pull/%d", d.url, p))
		} else {
			openBrowser(fmt.Sprintf("%s/pull/%d", d.url, p))
		}
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		fmt.Println(url)
	}

	if cmd != nil {
		fmt.Println("Opening", url)
		return cmd.Run()
	}

	return nil
}
