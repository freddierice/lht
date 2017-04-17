package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/freddierice/lht/project"
)

// updateCmd represents the update command
var fsUpdateCmd = &cobra.Command{
	Use:   "update <project> <build> <metafile>",
	Short: "update the filesystem using a metafile",
	Long:  `the meta file has a list of files to add/delete from the filesystem`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 3 {
			cmd.Usage()
			os.Exit(1)
		}
		projectName := args[0]
		buildName := args[1]
		metafile := args[2]

		err := func() error {
			proj, err := project.Open(projectName)
			if err != nil {
				return fmt.Errorf("could not open project: %v")
			}
			defer proj.Close()
			builder, err := proj.GetBuilder(buildName)
			if err != nil {
				return fmt.Errorf("could not open build: %v", err)
			}
			f, err := os.Open(metafile)
			if err != nil {
				return fmt.Errorf("could not open file: %v", err)
			}
			defer f.Close()

			if err := builder.UpdateFS(f); err != nil {
				return fmt.Errorf("could not update filesystem: %v", err)
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
	fsCmd.AddCommand(fsUpdateCmd)
}
