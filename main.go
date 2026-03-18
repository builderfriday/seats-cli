// main.go
package main

import (
	"github.com/derek/seats-cli/cmd"
)

func main() {
	// cmd.Execute() handles os.Exit() with proper exit codes internally.
	// Cobra prints usage errors to stderr and exits with code 1.
	cmd.Execute()
}
