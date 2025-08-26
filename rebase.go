package main

import (
	"encoding/json"
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
	output string,
) error {
	prns := filterChain(d, filter)
	if len(prns) == 0 {
		if output == "json" {
			jsonOutput := JSONRebaseOutput{
				Script:   "",
				Commands: []string{},
			}
			outputBytes, _ := json.MarshalIndent(jsonOutput, "", "  ")
			fmt.Println(string(outputBytes))
		} else {
			fmt.Println("No PR chain found with filter")
		}
		return nil
	}

	lines := []string{"#!/bin/sh", "", "set -e"}
	commands := []string{}

	for _, p := range prns {
		checkoutCmd := fmt.Sprintf("git checkout %s", d.prs[p].head)
		rebaseCmd := fmt.Sprintf("git rebase --update-refs %s", d.prs[p].base)

		lines = append(
			lines,
			"",
			checkoutCmd,
			rebaseCmd)

		commands = append(commands, checkoutCmd, rebaseCmd)

		if push {
			pushCmd := fmt.Sprintf("git push %s", args)
			lines = append(lines, pushCmd)
			commands = append(commands, pushCmd)
		}
	}

	script := strings.Join(lines, "\n")

	if output == "json" {
		jsonOutput := JSONRebaseOutput{
			Script:   script,
			Commands: commands,
		}
		outputBytes, _ := json.MarshalIndent(jsonOutput, "", "  ")
		fmt.Println(string(outputBytes))
		return nil
	}

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
