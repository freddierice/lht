package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/freddierice/lht.v1/project"
)

// addCmd represents the add command
var linuxAddCmd = &cobra.Command{
	Use:   "add <project name> <linux version> <linux tag> [flags]",
	Short: "Add a version of linux to a project",
	Long:  `Adds a version of linux to compile`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 3 {
			cmd.Usage()
			os.Exit(1)
		}

		projectName := args[0]
		buildVersion := args[1]
		tag := args[2]

		buildName, _ := cmd.Flags().GetString("name")
		if buildName == "" {
			buildName = buildVersion
		}

		proj, err := project.Open(projectName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not open project: %v\n", err)
			os.Exit(1)
		}

		if _, ok := proj.Builds[buildName]; ok {
			fmt.Fprintf(os.Stderr, "build already exists\n")
			proj.Close()
			os.Exit(1)
		}

		proj.Builds[buildName] = project.LinuxBuild{
			Name:         buildName,
			LinuxVersion: buildVersion,
			Tag:          tag,
		}

		if err := os.MkdirAll(filepath.Join(proj.Path(), buildName), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "could not create the build directory: %v\n", err)
			os.Exit(1)
		}

		proj.Close()
	},
}

func init() {
	linuxAddCmd.Flags().StringP("name", "n", "", "name for the linux build (defaults to version number)")
	linuxCmd.AddCommand(linuxAddCmd)
}
