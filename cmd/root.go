package cmd

import (
	"fmt"
	"os"

	"github.com/freddierice/lht/project"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "lht",
	Short: "Linux Hacking Toolkit",
	Long: `Linux Hacking Toolkit is a utility that will download, cross-compile, 
and launch linux under different environments to facilitate kernel exploitation`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(InitConfig)
}

// InitConfig reads in config file and ENV variables if set.
func InitConfig() {

	viper.SetDefault("Threads", "4")
	viper.SetDefault("RootDirectory", "/opt/lht")
	viper.SetConfigName("lht")  // name of config file (without extension)
	viper.AddConfigPath("/etc") // adding /etc directory as first search path

	if !project.CheckInstalled() {
		fmt.Fprintf(os.Stderr, "lht is not configured.. running installation.\n")
		if err := project.Install(); err != nil {
			fmt.Fprintf(os.Stderr, "could not install: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "lht has been installed.\n")
		os.Exit(0)
	}

	// config file should be found since we have checked the installation
	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "could not read configuration file: %v\n", err)
		os.Exit(1)
	}
}
