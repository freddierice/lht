package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

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
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	// get user and home directory for defaults
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get current user: %v\n", err)
		os.Exit(1)
	}
	if usr.HomeDir == "" {
		fmt.Fprintf(os.Stderr, "user has no home directory\n")
		os.Exit(1)
	}
	rootDirectory := filepath.Join(usr.HomeDir, ".lht")
	if err := os.MkdirAll(rootDirectory, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "could not create root directory: %v\n", err)
		os.Exit(1)
	}

	viper.SetDefault("Threads", "4")
	viper.Set("RootDirectory", rootDirectory)
	viper.SetConfigName("lht")  // name of config file (without extension)
	viper.AddConfigPath("/etc") // adding /etc directory as first search path

	// If a config file is not found, create one
	if err := viper.ReadInConfig(); err != nil {
		// create a new one
		confFile, err := os.Create(filepath.Join(usr.HomeDir, ".lht.yaml"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not create root configuration file (%v): %v\n", confFile, err)
			os.Exit(1)
		}
		confFile.Close()
	}
}
