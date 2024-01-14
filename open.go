package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"slices"
	"strconv"
)

func filterByBranch(d data, branch string) []int {
	num := d.branch[branch]
	prns := []int{num}

	// items before
	for {
		base := d.mappings[num].base
		prns = append([]int{base}, prns...)

		if base == 0 {
			break
		}

		num = base
	}

	// items after
	stack := []int{num}
	for {
		if len(stack) == 0 {
			break
		}

		last := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		following := d.mappings[last].following

		if len(following) > 0 {
			prns = append(prns, following[0])
		}

		if len(following) > 1 {
			slices.Reverse(following)
			stack = append(stack[:len(stack)-1], following[:len(following)-1]...)
		}
	}

	return prns
}

func filterByNumber(d data, num int) []int {
	prns := []int{}

	return prns
}

func openChain(d data, filter string, print bool) {
	prns := []int{}

	num, err := strconv.Atoi(filter)
	if err != nil {
		prns = filterByBranch(d, filter)
	} else {
		prns = filterByNumber(d, num)
	}

	// first item will be base(0)
	for _, p := range prns[1:] {
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
		return cmd.Run()
	}

	return nil
}
