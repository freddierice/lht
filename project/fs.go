package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"gopkg.in/freddierice/go-losetup.v1"

	"golang.org/x/sys/unix"
)

// Metafile holds data usefule for updating a filesystem.
type Metafile struct {
	HostPath   string `json:"hostPath"`
	TargetPath string `json:"targetPath"`
	UID        int    `json:"uid"`
	GID        int    `json:"gid"`
	Mode       string `json:"mode"`
}

func doUnderMount(task func(string) error, fsImagePath string) error {

	mountDir, err := ioutil.TempDir("", "lht")
	if err != nil {
		return err
	}
	defer os.RemoveAll(mountDir)

	dev, err := losetup.Attach(fsImagePath, 0, false)
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

// UpdateFS will update the filesystem. If the metafile describes HostPath,
// then UpdateFS will copy that file to the target system at targetFile. If
// HostPath is not provided, then UpdateFS will create a directory. Paths can
// also use golang template formats to write files to different locations based
// on the current build. The templating engine will get run with a builder.
func (builder *Builder) UpdateFS(r io.Reader) error {

	fsImagePath := builder.GetBuildDir("rootfs.img")
	if !exists(fsImagePath) {
		return fmt.Errorf("filesystem does not exist")
	}

	metafiles := []Metafile{}
	dec := json.NewDecoder(r)
	if err := dec.Decode(&metafiles); err != nil {
		return err
	}

	copyFiles := func(mountDir string) error {
		for _, m := range metafiles {

			perms, err := strconv.ParseInt(m.Mode, 8, 32)
			if err != nil {
				return fmt.Errorf("invalid perms: %v", err)
			}

			buf := bytes.NewBuffer(make([]byte, 0, 4096))
			hostTemplate, err := template.New("hostPath").Parse(m.HostPath)
			if err != nil {
				return fmt.Errorf("invalid template: %v", err)
			}
			if err := hostTemplate.Execute(buf, builder); err != nil {
				return fmt.Errorf("could not execute template: %v", err)
			}
			m.HostPath = buf.String()

			buf.Reset()
			targetTemplate, err := template.New("targetPath").Parse(m.TargetPath)
			if err != nil {
				return fmt.Errorf("invalid template: %v", err)
			}
			if err := targetTemplate.Execute(buf, builder); err != nil {
				return fmt.Errorf("could not execute template: %v", err)
			}
			m.TargetPath = buf.String()
			m.TargetPath = filepath.Join(mountDir, m.TargetPath)

			// HostPath non-empty implies file, directory otherwise
			if m.HostPath != "" {
				hostFile, err := os.Open(m.HostPath)
				if err != nil {
					return fmt.Errorf("could not open %v", m.HostPath)
				}
				targetFile, err := os.Create(m.TargetPath)
				if err != nil {
					return fmt.Errorf("could not open target %v", m.TargetPath)
				}
				if _, err := io.Copy(targetFile, hostFile); err != nil {
					return fmt.Errorf("could not copy file: %v", err)
				}
				hostFile.Close()
				if err := targetFile.Chown(m.UID, m.GID); err != nil {
					return fmt.Errorf("could not chown file: %v", err)
				}
				if err := targetFile.Chmod(os.FileMode(perms)); err != nil {
					return fmt.Errorf("could not chmod file: %v", err)
				}
				targetFile.Close()
			} else {
				if err := os.MkdirAll(m.TargetPath, os.FileMode(perms)); err != nil {
					return fmt.Errorf("could not create directory: %v", err)
				}
				if err := os.Chown(m.TargetPath, m.UID, m.GID); err != nil {
					return fmt.Errorf("could not chown directory: %v", err)
				}
				if err := os.Chmod(m.TargetPath, os.FileMode(perms)); err != nil {
					return fmt.Errorf("could not chmod directory")
				}
			}
		}

		return nil
	}

	return doUnderMount(copyFiles, fsImagePath)
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
