package cmd

import (
	"fmt"
	"os"

	"github.com/freddierice/lht/project"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <project> <build>",
	Short: "Creates a root filesystem image",
	Long:  `Creates a root filesystem image`,
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
				return fmt.Errorf("could not open builder: %v", err)
			}

			if err := builder.CreateRootFS(); err != nil {
				return fmt.Errorf("could not create root filesystem: %v", err)
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
	fsCmd.AddCommand(createCmd)
}
