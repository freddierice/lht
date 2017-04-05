package cmd

import (
	"fmt"
	"os"

	"github.com/freddierice/lht/project"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init <project name>",
	Short: "Initializes a project.",
	Long: `Each project has its own compilations of linux, for different
architectures, configurations, etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		}
		projectName := args[0]

		fmt.Printf("creating project %v\n", projectName)
		proj, err := project.Create(projectName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not create project: %v\n", err)
			os.Exit(1)
		}

		proj.Arch, _ = cmd.Flags().GetString("arch")
		proj.CrossCompilePrefix, _ = cmd.Flags().GetString("cross-compile-prefix")
		proj.Defconfig, _ = cmd.Flags().GetString("defconfig")

		if err := proj.Write(); err != nil {
			fmt.Fprintf(os.Stderr, "could not write out project configuration: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	initCmd.Flags().StringP("arch", "a", "x86_64", "Architecture for which to build linux")
	initCmd.Flags().StringP("cross-compile-prefix", "c", "", "A prefix for compiling to non-host archtectures")
	initCmd.Flags().StringP("defconfig", "d", "", "Defconfig file for configuring the kernel")
	RootCmd.AddCommand(initCmd)
}
