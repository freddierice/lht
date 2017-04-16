package project

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func doMountCopy(versionDir string) error {

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

	/*
		dev, err := losetup.Attach("rootfs.img", 0, false)
		if err != nil {
			return err
		}
	*/

	// TODO: add unix.MS_NOEXEC | unix.MS_NOSUID
	if err := unix.Mount("rootfs.img", mountDir, "ext2", 0, ""); err != nil {
		return err
	}

	if err := copyAll(filepath.Join(versionDir, "sysroot"), mountDir); err != nil {
		return err
	}

	if err := copyAll(filepath.Join(versionDir, "busybox", "_install"), mountDir); err != nil {
		return err
	}

	if err := unix.Unmount(mountDir, 0); err != nil {
		return err
	}

	return nil
}
