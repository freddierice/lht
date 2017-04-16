package cmd

import (
	"fmt"
	"os"

	"github.com/freddierice/lht/project"
	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var buildLinuxCmd = &cobra.Command{
	Use:   "build <project name> <build name>",
	Short: "builds a linux image with a project and version number",
	Long:  ``,
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

		builder, err := proj.GetBuilder(buildName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not open build: %v\n", err)
			proj.Close()
			os.Exit(1)
		}

		err = builder.BuildAll()
		proj.Builds[buildName] = builder.LinuxBuild
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not build all: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	linuxCmd.AddCommand(buildLinuxCmd)
}
