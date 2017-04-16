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

// CreateRootFS creates a root filesystem.
func (builder *Builder) CreateRootFS() error {

	mountDir, err := ioutil.TempDir("", "lht")
	if err != nil {
		return err
	}
	defer os.RemoveAll(mountDir)

	fsImagePath := builder.GetBuildDir("rootfs.img")

	if !exists(fsImagePath) {
		stats, err := os.Stat(filepath.Dir(fsImagePath))
		if err != nil {
			return err
		}
		f, err := os.Create(fsImagePath)
		if err != nil {
			return err
		}
		if err := f.Chown(int(stats.Sys().(*syscall.Stat_t).Uid), int(stats.Sys().(*syscall.Stat_t).Gid)); err != nil {
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
	}

	dev, err := losetup.Attach(fsImagePath, 0, false)
	if err != nil {
		return err
	}

	// TODO: add unix.MS_NOEXEC | unix.MS_NOSUID
	if err := unix.Mount(dev.Path(), mountDir, "ext2", 0, ""); err != nil {
		return err
	}

	if err := copyAll(builder.GetBuildDir("sysroot"), mountDir); err != nil {
		return err
	}

	if err := copyAll(filepath.Join(builder.GetBuildDir("busybox"), "_install"), mountDir); err != nil {
		return err
	}

	if err := unix.Unmount(mountDir, 0); err != nil {
		return err
	}

	return nil
}
