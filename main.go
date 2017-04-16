package main

import (
	"fmt"
	"os"

	"gopkg.in/freddierice/lht.v1/cmd"
)

func main() {
	// hack -- do initConfig first
	cmd.InitConfig()

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
