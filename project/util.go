package project

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/docker/docker/pkg/reexec"
)

func init() {
	reexec.Register("doMount", doMount)
	if reexec.Init() {
		os.Exit(0)
	}
}

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

// mount mounts dev onto dir. Note: this implementation is gross right now. In
// the future, this will be implemented without mounting.
func mount(dev, dir string) error {
	cmd := reexec.Command("doMount", "dev1", "dir1")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS |
			syscall.CLONE_NEWPID,
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not mount: %v", err)
	}

	return nil
}

func doMount() {
	fmt.Printf("in doMount(%v, %v)\n", os.Args[1], os.Args[2])

	os.Exit(0)
}
