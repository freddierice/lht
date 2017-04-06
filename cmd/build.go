package cmd

import (
	"fmt"
	"os"

	"github.com/freddierice/lht/project"
	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build <project name> <linux version>",
	Short: "builds a linux image with a project and version number",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.Usage()
			os.Exit(1)
		}
		projectName := args[0]
		linuxVersion := args[1]

		proj, err := project.Open(projectName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not open project: %v\n", err)
			os.Exit(1)
		}

		if err := proj.BuildLinux(linuxVersion); err != nil {
			fmt.Fprintf(os.Stderr, "could not build linux: %v\n", err)
			os.Exit(1)
		}

		if err := proj.BuildVulnKo(linuxVersion); err != nil {
			fmt.Fprintf(os.Stderr, "could not build vuln-ko: %v\n", err)
			os.Exit(1)
		}

		if err := proj.BuildGlibc(linuxVersion); err != nil {
			fmt.Fprintf(os.Stderr, "could not build glibc: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(buildCmd)

}
