package cmd

import (
	"fmt"
	"os"

	"github.com/freddierice/lht/project"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <project> <build>",
	Short: "Creates a root filesystem image",
	Long:  `Creates a root filesystem image`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.Usage()
			os.Exit(1)
		}

		projectName := args[0]
		buildName := args[1]

		proj, err := project.Open(projectName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not open project: %v\n", err)
			os.Exit(1)
		}

		if err := proj.CreateRootFS(buildName); err != nil {
			fmt.Fprintf(os.Stderr, "could not create root filesystem: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	fsCmd.AddCommand(createCmd)
}
