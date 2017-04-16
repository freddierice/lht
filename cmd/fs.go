package cmd

import "github.com/spf13/cobra"

// fsCmd represents the fs command
var fsCmd = &cobra.Command{
	Use:   "fs",
	Short: "A set of commands to manipulate the filesystem",
	Long:  `A set of commands to manipulate the filesystem.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func init() {
	RootCmd.AddCommand(fsCmd)
}
