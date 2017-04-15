package cmd

import (
	"fmt"
	"io/ioutil"
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
		proj.Target, _ = cmd.Flags().GetString("target")
		proj.Host, _ = cmd.Flags().GetString("host")
		proj.GlibcVersion, _ = cmd.Flags().GetString("glibc-version")
		proj.BusyBoxVersion, _ = cmd.Flags().GetString("busybox-version")
		proj.FsSize, err = cmd.Flags().GetUint64("fs-size")
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid fsSize: %v\n", err)
			os.Exit(1)
		}

		defconfigFile, _ := cmd.Flags().GetString("defconfig")
		if defconfigFile != "" {
			buf, err := ioutil.ReadFile(defconfigFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "could not read defconfig file %v\n", err)
				os.Exit(1)
			}
			proj.Defconfig = string(buf)
		}

		if err := proj.Write(); err != nil {
			fmt.Fprintf(os.Stderr, "could not write out project configuration: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	initCmd.Flags().StringP("arch", "a", "", "Architecture for which to build linux. Only use this flag if cross compiling")
	initCmd.Flags().StringP("target", "t", "x86_64-pc-linux-gnu", "A prefix for compiling to non-host archtectures")
	initCmd.Flags().StringP("host", "H", "x86_64-pc-linux-gnu", "A prefix to define host architecture")
	initCmd.Flags().StringP("defconfig", "d", "", "Defconfig file for configuring the kernel")
	initCmd.Flags().String("glibc-version", "2.25", "Glibc version to use")
	initCmd.Flags().String("busybox-version", "1.26.2", "Busybox version to use")
	initCmd.Flags().Uint64P("fs-size", "s", 549755813888, "Size of the root filesystem (defaults to 512 megabytes)")
	RootCmd.AddCommand(initCmd)
}
