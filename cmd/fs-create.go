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

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <project> <build>",
	Short: "Creates a root filesystem image",
	Long:  `Creates a root filesystem image`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(os.Args) != 2 {
			cmd.Usage()
			os.Exit(1)
		}

		projectName := args[0]
		buildName := args[1]

		proj, err := project.Open(projectName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not open project: %v\n", err)
			os.Exit(1)
		}

		if err := proj.CreateRootFS(buildName); err != nil {
			fmt.Fprintf(os.Stderr, "could not create root filesystem: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	fsCmd.AddCommand(createCmd)
}
