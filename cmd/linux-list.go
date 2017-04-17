package cmd

import (
	"fmt"
	"os"

	"github.com/freddierice/lht/project"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list <project name>",
	Short: "list all linux builds for a project",
	Long:  `list all linux builds for a project`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		}
		projectName := args[0]

		proj, err := project.Open(projectName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not open project: %v\n", err)
		}

		for buildName, linuxBuild := range proj.Builds {
			fmt.Printf("%v:\n    Version: %v\n", buildName, linuxBuild.LinuxVersion)
		}

		proj.Close()
	},
}

func init() {
	linuxCmd.AddCommand(listCmd)
}
