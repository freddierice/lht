// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/freddierice/lht/project"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete <project name>",
	Short: "Delete an existing project",
	Long:  `Deletes an existing project and all of the linux builds associated with it`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(os.Args) != 1 {
			cmd.Usage()
			os.Exit(1)
		}
		projectName := os.Args[0]

		// make sure it exists
		proj, err := project.Open(projectName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "project does not exist\n")
			os.Exit(1)
		}

		if err := os.RemoveAll(proj.Path()); err != nil {
			fmt.Fprintf(os.Stderr, "could not remove project: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	projectCmd.AddCommand(deleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
