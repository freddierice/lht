package project

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func dirCreate(path string) string {
	os.MkdirAll(path, 0755)
	return path
}

func getRootDir() string {
	return viper.GetString("RootDirectory")
}

func getDownloadDir() string {
	return dirCreate(filepath.Join(getRootDir(), ".downloads"))
}

func getConfDir() string {
	return dirCreate(filepath.Join(getRootDir(), "conf"))
}
