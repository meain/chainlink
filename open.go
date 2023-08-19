package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch {
	case isWindows():
		cmd = exec.Command("cmd", "/c", "start", url)
	case isLinux():
		cmd = exec.Command("xdg-open", url)
	case isMac():
		cmd = exec.Command("open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Run()
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}

func isLinux() bool {
	return runtime.GOOS == "linux"
}

func isMac() bool {
	return runtime.GOOS == "darwin"
}
