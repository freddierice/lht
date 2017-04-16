package project

import (
	"path/filepath"

	"github.com/spf13/viper"
)

func getRootDir() string {
	return viper.GetString("RootDirectory")
}

func getDownloadDir() string {
	return filepath.Join(getRootDir(), ".downloads")
}

func getConfDir() string {
	return filepath.Join(getRootDir(), "conf")
}
