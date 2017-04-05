package cmd

import (
	"fmt"
	"os"

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
		defconfigName, _ := cmd.Flags().GetString("defconfig")

		if cmd.Flags().Changed("defconfig") {
			fmt.Printf("creating project %v with defconfig %v\n", projectName, defconfigName)
		} else {
			fmt.Printf("creating project %v\n", projectName)
		}
	},
}

func init() {
	initCmd.Flags().StringP("arch", "a", "x86_64", "Architecture for which to build linux")
	initCmd.Flags().StringP("cross-compile-prefix", "c", "", "A prefix for compiling to non-host archtectures")
	initCmd.Flags().StringP("defconfig", "d", "", "Defconfig file for configuring the kernel")
	RootCmd.AddCommand(initCmd)
}
