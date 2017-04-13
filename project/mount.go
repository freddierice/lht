package project

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"

	"gopkg.in/freddierice/go-losetup.v1"

	"github.com/docker/docker/pkg/reexec"
)

func init() {
	reexec.Register("doMountCopyCmd", doMountCopyCmd)
	if reexec.Init() {
		os.Exit(0)
	}
}

// mount mounts dev onto dir. Note: this implementation is gross right now. In
// the future, this will be implemented without mounting.
func mount(versionDir string) error {
	cmd := reexec.Command("doMountCopyCmd", versionDir)
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

// called with doMount <version dir>
func doMountCopyCmd() {
	if len(os.Args) != 2 {
		os.Exit(1)
	}
	versionDir := os.Args[1]

	if err := doMountCopy(versionDir); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func doMountCopy(versionDir string) error {

	if err := os.Chdir(versionDir); err != nil {
		return err
	}

	mountDir, err := ioutil.TempDir("", "lht")
	if err != nil {
		return err
	}
	defer os.RemoveAll(mountDir)

	if !exists("rootfs.img") {
		f, err := os.Create("rootfs.img")
		if err != nil {
			return err
		}
		// 512 megs
		if err := f.Truncate(549755813888); err != nil {
			f.Close()
			os.Remove("rootfs.img")
			return err
		}

		cmd := exec.Command("mkfs", "-t", "ext2", "rootfs.img")
		if err := cmd.Run(); err != nil {
			f.Close()
			os.Remove("rootfs.img")
			return err
		}
	}

	dev, err := losetup.Attach("rootfs.img", 0, false)
	if err != nil {
		return err
	}

	// TODO: add unix.MS_NOEXEC | unix.MS_NOSUID
	if err := unix.Mount(dev.Path(), mountDir, "ext2", 0, ""); err != nil {
		return err
	}

	for _, dir := range []string{"bin", "etc", "include", "lib", "sbin", "share", "var"} {
		cmd := exec.Command("cp", "-r", filepath.Join(versionDir, "sysroot", dir), mountDir)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	for _, dir := range []string{"bin", "sbin", "usr", "linuxrc"} {
		cmd := exec.Command("cp", "-r", filepath.Join(versionDir, "busybox", "_install", dir), mountDir)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	if err := unix.Unmount(mountDir, 0); err != nil {
		return err
	}

	return nil
}
