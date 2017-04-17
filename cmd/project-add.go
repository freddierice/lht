package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/freddierice/lht/project"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var projectAddCmd = &cobra.Command{
	Use:   "add <project name>",
	Short: "Add a new project to lht",
	Long: `Each project has its own compilations of linux, for different
architectures, configurations, etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		}
		projectName := args[0]

		err := func() error {
			fmt.Printf("creating project %v\n", projectName)
			proj, err := project.Create(projectName)
			if err != nil {
				return fmt.Errorf("could not create project: %v", err)
			}

			proj.Meta.Arch, _ = cmd.Flags().GetString("arch")
			proj.Target, _ = cmd.Flags().GetString("target")
			proj.Host, _ = cmd.Flags().GetString("host")
			proj.GlibcVersion, _ = cmd.Flags().GetString("glibc-version")
			proj.BusyBoxVersion, _ = cmd.Flags().GetString("busybox-version")
			proj.FsSize, err = cmd.Flags().GetUint64("fs-size")
			if err != nil {
				proj.Delete()
				return fmt.Errorf("invalid fsSize: %v", err)
			}

			defconfigFile, _ := cmd.Flags().GetString("defconfig")
			if defconfigFile != "" {
				buf, err := ioutil.ReadFile(defconfigFile)
				if err != nil {
					proj.Delete()
					return fmt.Errorf("could not read defconfig: %v", err)
				}
				proj.Defconfig = string(buf)
			}

			if err := proj.Commit(); err != nil {
				proj.Delete()
				return fmt.Errorf("could not write out project configuration: %v", err)
			}

			proj.Close()

			return nil
		}()

		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	projectAddCmd.Flags().StringP("arch", "a", "", "Architecture for which to build linux. Only use this flag if cross compiling")
	projectAddCmd.Flags().StringP("target", "t", "x86_64-pc-linux-gnu", "A prefix for compiling to non-host archtectures")
	projectAddCmd.Flags().StringP("host", "H", "x86_64-pc-linux-gnu", "A prefix to define host architecture")
	projectAddCmd.Flags().StringP("defconfig", "d", "", "Defconfig file for configuring the kernel")
	projectAddCmd.Flags().String("glibc-version", "2.25", "Glibc version to use")
	projectAddCmd.Flags().String("busybox-version", "1.26.2", "Busybox version to use")
	projectAddCmd.Flags().Uint64P("fs-size", "s", 536870912, "Size of the root filesystem (defaults to 512 megabytes)")
	projectCmd.AddCommand(projectAddCmd)
}
