package cmd

import (
	"fmt"
	"os"

	"gopkg.in/freddierice/lht.v1/project"
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

		err := func() error {
			proj, err := project.Open(projectName)
			if err != nil {
				return fmt.Errorf("could not open project: %v", err)

			}
			defer proj.Close()

			builder, err := proj.GetBuilder(buildName)
			if err != nil {
				return fmt.Errorf("could not open build: %v", err)
			}

			err = builder.BuildAll()
			proj.Builds[buildName] = builder.LinuxBuild
			if err != nil {
				return fmt.Errorf("could not build all: %v", err)
			}

			return nil
		}()

		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	linuxCmd.AddCommand(buildLinuxCmd)
}
