package cmd

import "github.com/spf13/cobra"

// projectCmd represents the project command
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "A set of subcommands that deal with manipulating projects",
	Long: `Project is a set of subcommands that deal with manipulating projects. The
'add' function creates a new project and 'delete' deletes an existing project.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func init() {
	RootCmd.AddCommand(projectCmd)
}
