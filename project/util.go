package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// execAt executes a program at a given working directory. If there are any
// errors while executing the program, print the string out.
func execAt(dir, cmdStr string, args ...string) error {

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := os.Chdir(dir); err != nil {
		return err
	}

	cmd := exec.Command(cmdStr, args...)
	if err := cmd.Run(); err != nil {
		out, err2 := cmd.CombinedOutput()
		fmt.Printf("error executing %v at %v.\n", cmdStr, dir)
		if err2 != nil {
			fmt.Printf("no output could be printed: %v", err2)
			return err
		}
		fmt.Println(string(out))
		return err
	}

	return os.Chdir(wd)
}

func copyAll(src, dest string) error {
	var err error

	if src, err = filepath.Abs(src); err != nil {
		return err
	}

	copyAllWalkFunc := func(path string, info os.FileInfo, err error) error {
		return nil
	}

	return filepath.Walk(src, copyAllWalkFunc)
}
