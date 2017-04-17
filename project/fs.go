package project

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"gopkg.in/freddierice/go-losetup.v1"

	"golang.org/x/sys/unix"
)

func doUnderMount(task func(string) error, imagePath string) error {

	mountDir, err := ioutil.TempDir("", "lht")
	if err != nil {
		return err
	}
	defer os.RemoveAll(mountDir)

	dev, err := losetup.Attach(imagePath, 0, false)
	if err != nil {
		return err
	}
	defer dev.Detach()

	// TODO: add unix.MS_NOEXEC | unix.MS_NOSUID
	if err := unix.Mount(dev.Path(), mountDir, "ext2", 0, ""); err != nil {
		return err
	}
	defer unix.Unmount(mountDir, 0)

	return task(mountDir)
}

// CreateRootFS creates a root filesystem.
func (builder *Builder) CreateRootFS() error {

	fsImagePath := builder.GetBuildDir("rootfs.img")
	os.Remove(fsImagePath)

	stats, err := os.Stat(filepath.Dir(fsImagePath))
	if err != nil {
		return err
	}
	f, err := os.Create(fsImagePath)
	if err != nil {
		return err
	}

	unixStats, ok := stats.Sys().(*syscall.Stat_t)
	if !ok {
		panic("only runs on unix")
	}

	if err := f.Chown(int(unixStats.Uid), int(unixStats.Gid)); err != nil {
		return err
	}

	if err := f.Truncate(int64(builder.Meta.FsSize)); err != nil {
		f.Close()
		os.Remove(fsImagePath)
		return err
	}

	cmd := exec.Command("mkfs", "-t", "ext2", fsImagePath)
	if err := cmd.Run(); err != nil {
		f.Close()
		os.Remove(fsImagePath)
		return err
	}

	doCopyAll := func(mountDir string) error {
		if err := copyAll(builder.GetBuildDir("sysroot"), mountDir); err != nil {
			return err
		}

		if err := copyAll(filepath.Join(builder.GetBuildDir("busybox"), "_install"), mountDir); err != nil {
			return err
		}

		// setup filesystem
		for _, dir := range []string{"dev", "etc", "proc", "sys"} {
			os.MkdirAll(filepath.Join(mountDir, dir), 0755)
		}
		os.MkdirAll(filepath.Join(mountDir, "etc", "init.d"), 0755)

		simpleStart := `
#!/bin/bash
mknod /dev/tty0 c 4 0
mknod /dev/tty1 c 4 1
mknod /dev/tty2 c 4 2
mknod /dev/tty3 c 4 3
mknod /dev/tty4 c 4 4
mount -t proc proc /proc -o rw,nosuid,nodev,noexec,relatime
mount -t sysfs sys /sys -o rw,nosuid,nodev,noexec,relatime
`
		f, err = os.Create(filepath.Join(mountDir, "etc", "init.d", "rcS"))
		if err != nil {
			return err
		}
		f.Chmod(0755)
		f.WriteString(simpleStart)
		f.Close()
		return nil
	}

	return doUnderMount(doCopyAll, fsImagePath)
}
