package main

import (
	"fmt"
	"os"

	"github.com/freddierice/lht/cmd"
)

func main() {
	// hack -- do initConfig first
	cmd.InitConfig()

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
