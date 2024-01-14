package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func rebaseChain(
	d data,
	filter string,
	push, run bool,
	args string,
	shell string,
) error {
	prns := filterChain(d, filter)
	if len(prns) == 0 {
		fmt.Println("No PR chain found with filter")
		return nil
	}

	lines := []string{"#!/bin/sh", "", "set -e"}

	for _, p := range prns {
		lines = append(
			lines,
			"",
			fmt.Sprintf("git checkout %s", d.prs[p].head),
			fmt.Sprintf("git rebase --update-refs %s", d.prs[p].base))

		if push {
			lines = append(lines, fmt.Sprintf("git push %s", args))
		}
	}

	script := strings.Join(lines, "\n")

	if run {
		if shell == "$SHELL" {
			shell = os.Getenv("SHELL")
			if len(shell) == 0 {
				shell = "/bin/sh"
			}
		}

		err := execScript(script, shell)
		if err != nil {
			return fmt.Errorf("unable to exec script: %v", err)
		}

		return nil
	}

	fmt.Println(script)
	return nil
}

func execScript(script, shell string) error {
	cmd := exec.Command(shell)
	cmd.Stdin = strings.NewReader(script)
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
