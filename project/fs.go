package project

import "os"

func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
