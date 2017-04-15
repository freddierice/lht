package project

import (
	"fmt"
	"io"
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

// copyAll copies all files from src to dest.
func copyAll(src, dest string) error {
	var err error

	if src, err = filepath.Abs(src); err != nil {
		return err
	}
	if dest, err = filepath.Abs(dest); err != nil {
		return err
	}
	srcLen := len(src)

	copyAllWalkFunc := func(path string, info os.FileInfo, err error) error {
		newDest := filepath.Join(dest, string(path[srcLen:]))
		if info.IsDir() {
			return os.MkdirAll(newDest, info.Mode())
		}

		destFile, err := os.OpenFile(newDest, os.O_RDWR|os.O_CREATE, info.Mode())
		if err != nil {
			return err
		}
		defer destFile.Close()

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		_, err = io.Copy(destFile, srcFile)
		return err
	}

	return filepath.Walk(src, copyAllWalkFunc)
}
