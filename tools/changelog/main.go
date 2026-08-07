package main

import (
	"os"

	changelogcmd "changelog/internal/cmd"
)

func main() {
	err := changelogcmd.NewRootCmd().Execute()
	if err != nil {
		_, _ = os.Stderr.WriteString("error: " + err.Error() + "\n")
		os.Exit(1)
	}
}
