package cmd

import "github.com/spf13/cobra"

// linuxCmd represents the linux command
var linuxCmd = &cobra.Command{
	Use:   "linux",
	Short: "A set of subcommands that build linux",
	Long:  `A set of subcommands that build linux`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func init() {
	RootCmd.AddCommand(linuxCmd)
}
