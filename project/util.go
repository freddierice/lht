package project

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"golang.org/x/sys/unix"

	"github.com/spf13/viper"
)

// ErrNotRoot is an error produced when the program expects to be running as
// root, but instead is running as an unprivileged user.
var ErrNotRoot = fmt.Errorf("not root")

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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
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
func copyAllGit(src, dest string) error {
	var err error

	if src, err = filepath.Abs(src); err != nil {
		return err
	}
	if dest, err = filepath.Abs(dest); err != nil {
		return err
	}
	srcLen := len(src)

	copyAllWalkFunc := func(path string, info os.FileInfo, err error) error {
		if filepath.Base(path) == ".git" {
			return nil
		}
		if info == nil {
			return nil
		}

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

// CheckInstalled returns true if the system is configured correctly, and false
// if lht needs to run the install again.
func CheckInstalled() bool {
	// look for a configuration file
	if !exists("/etc/lht.yaml") {
		return false
	}

	// check for root directory
	if !exists(viper.GetString("RootDirectory")) {
		return false
	}

	return true
}

// Install sets lht up correctly.
func Install() error {
	if unix.Getuid() != 0 || unix.Getgid() != 0 {
		return ErrNotRoot
	}
	if !exists("/etc/lht.yaml") {
		f, err := os.Create("/etc/lht.yaml")
		if err != nil {
			return err
		}
		if err := f.Chown(0, 0); err != nil {
			return err
		}
		if err := f.Chmod(0644); err != nil {
			return err
		}
		f.Close()
	}

	uid := 0
	gid := 0
	if uidStr := os.Getenv("SUDO_UID"); uidStr != "" {
		var err error
		if uid, err = strconv.Atoi(uidStr); err != nil {
			return err
		}
	}
	if uidStr := os.Getenv("SUDO_GID"); uidStr != "" {
		var err error
		if uid, err = strconv.Atoi(uidStr); err != nil {
			return err
		}
	}

	rootDirectory := viper.GetString("RootDirectory")
	if !exists(rootDirectory) {
		if err := os.MkdirAll(rootDirectory, 0755); err != nil {
			return err
		}
		if err := os.Chown(rootDirectory, uid, gid); err != nil {
			return err
		}
	}

	return nil
}
