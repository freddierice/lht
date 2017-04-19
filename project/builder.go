package project

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/viper"
)

// Builder contains metadata to build projects. It is generally created by
// a project combined with a LinuxBuild.
type Builder struct {
	RootDir     string
	DownloadDir string
	ProjectDir  string
	BuildDir    string

	Meta
	LinuxBuild
}

// BuildSetenv sets the environment variables needed to build the projects.
func (builder *Builder) BuildSetenv() {
	if builder.Meta.Arch != "" {
		os.Setenv("ARCH", builder.Meta.Arch)
		os.Setenv("CROSS_COMPILE", builder.Meta.Target+"-")
	}
}

// GetStatus determines if a builder has completed a task.
func (builder *Builder) GetStatus(task string) bool {
	status, ok := builder.LinuxBuild.Status[task]
	return status && ok
}

// SetStatus saves the status of a given task.
func (builder *Builder) SetStatus(task string, status bool) {
	builder.LinuxBuild.Status[task] = status
}

// GetBuildDir appends build to the root directory of this build
func (builder *Builder) GetBuildDir(build string) string {
	return filepath.Join(builder.BuildDir, build)
}

// BuildAll builds all of the projects for a specific version of linux.
func (builder *Builder) BuildAll() error {
	if err := builder.BuildLinux(); err != nil {
		return fmt.Errorf("could not build linux: %v", err)
	}

	if err := builder.BuildVulnKo(); err != nil {
		return fmt.Errorf("could not build vuln-ko: %v", err)
	}

	if err := builder.BuildGlibc(); err != nil {
		return fmt.Errorf("could not build glibc: %v", err)
	}

	if err := builder.BuildBusyBox(); err != nil {
		return fmt.Errorf("could not build busybox: %v", err)
	}

	return nil
}

// BuildVulnKo builds the vuln-ko kernel module.
func (builder *Builder) BuildVulnKo() error {
	if builder.GetStatus("BuildVulnKo") {
		return nil
	}

	builder.BuildSetenv()

	fmt.Println("downloading vuln-ko")
	vulnDir, err := builder.DownloadVulnKo()
	if err != nil {
		return err
	}

	vulnDirNew := builder.GetBuildDir("vuln-ko")
	if !exists(vulnDirNew) {
		if err := os.MkdirAll(vulnDirNew, 0755); err != nil {
			return err
		}
		if err := copyAll(vulnDir, vulnDirNew); err != nil {
			os.RemoveAll(vulnDirNew)
			return err
		}
	}

	os.Setenv("BUILD_DIR", builder.GetBuildDir("linux"))
	fmt.Println("building vuln-ko")

	err = execAt(vulnDirNew, "make")
	if err != nil {
		return err
	}
	builder.SetStatus("BuildVulnKo", true)
	return nil
}

// BuildBusyBox builds the busybox.
func (builder *Builder) BuildBusyBox() error {
	if builder.GetStatus("BuildBusyBox") {
		return nil
	}
	builder.BuildSetenv()
	fmt.Println("downloading busybox")
	filename, err := builder.DownloadBusyBox()
	if err != nil {
		return err
	}

	busyBoxPath := builder.GetBuildDir("busybox")
	busyBoxPathOld := builder.GetBuildDir("busybox-" + builder.Meta.BusyBoxVersion)
	busyBoxInstallPath := filepath.Join(busyBoxPath, "_install")
	if exists(busyBoxInstallPath) {
		return nil
	}
	if !exists(busyBoxPath) {
		fmt.Println("extracting busybox")
		//TODO: replace with in-code solution
		cmd := exec.Command("tar", "-C", builder.BuildDir, "-xf", filename)
		if err := cmd.Run(); err != nil {
			return err
		}
		if err := os.Rename(busyBoxPathOld, busyBoxPath); err != nil {
			return err
		}
	}

	fmt.Println("making busybox defconfig")
	if err := execAt(busyBoxPath, "make", "defconfig"); err != nil {
		return err
	}

	fmt.Println("making busybox")
	if err := execAt(busyBoxPath, "make", "-j", viper.GetString("Threads")); err != nil {
		return err
	}

	fmt.Println("installing busybox")
	if err := execAt(busyBoxPath, "make", "install"); err != nil {
		return err
	}

	builder.SetStatus("BuildBusyBox", true)

	return nil
}

// BuildGlibc builds the Glibc project for a given linux version.
func (builder *Builder) BuildGlibc() error {
	if builder.GetStatus("BuildGlibc") {
		return nil
	}
	builder.BuildSetenv()
	fmt.Println("downloading glibc")
	filename, err := builder.DownloadGlibc()
	if err != nil {
		return err
	}

	fmt.Println("extracting glibc")
	glibcPath := builder.GetBuildDir("glibc")
	glibcPathOld := builder.GetBuildDir("glibc-" + builder.Meta.GlibcVersion)
	glibcBuildPath := builder.GetBuildDir("glibc-build")
	sysroot := builder.GetBuildDir("sysroot")
	if exists(sysroot) {
		return nil
	}
	if !exists(glibcPath) {
		//TODO: replace with in-code solution
		cmd := exec.Command("tar", "-C", builder.BuildDir, "-xf", filename)
		if err := cmd.Run(); err != nil {
			return err
		}
		if err := os.Rename(glibcPathOld, glibcPath); err != nil {
			return err
		}
	}

	if exists(glibcBuildPath) {
		if err := os.RemoveAll(glibcBuildPath); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(glibcBuildPath, 0755); err != nil {
		return err
	}

	fmt.Println("configuring glibc")
	os.Setenv("CFLAGS", "-O2")
	os.Setenv("CPPFLAGS", "-O2")
	os.Setenv("BUILD_CC", "gcc")
	os.Setenv("CC", builder.Meta.Target+"-gcc")
	os.Setenv("CXX", builder.Meta.Target+"-g++")
	os.Setenv("AR", builder.Meta.Target+"-ar")
	os.Setenv("RANLIB", builder.Meta.Target+"-ranlib")

	configparams := fmt.Sprintf("slibdir=/lib\nrtlddir=/lib\nsbindir=/bin\nrootsbindir=/bin\nbuild-programs=no\n")
	configparamsFilename := filepath.Join(glibcPath, "configparams")
	if err := ioutil.WriteFile(configparamsFilename, []byte(configparams), 0644); err != nil {
		return err
	}

	headersPath := filepath.Join(builder.GetBuildDir("headers"), "include")
	err = execAt(glibcBuildPath, "../glibc/configure", "--prefix=/", "--libdir=/lib", "--libexecdir=/lib",
		"--enable-add-ons", "--enable-kernel=2.6.32", "--enable-lock-elision",
		"--enable-stackguard-randomization", "--enable-bind-now", "--disable-profile",
		"--disable-multi-arch", "--disable-werror", "--target="+builder.Meta.Target,
		"--host="+builder.Meta.Target, "--build="+builder.Meta.Host, "--with-headers="+headersPath)
	if err != nil {
		return err
	}

	fmt.Println("building glibc")
	if err := execAt(glibcBuildPath, "make", "-j", viper.GetString("Threads")); err != nil {
		return err
	}

	if err := os.MkdirAll(sysroot, 0755); err != nil {
		return err
	}

	fmt.Println("installing glibc")
	if err := execAt(glibcBuildPath, "make", "install_root="+sysroot, "install"); err != nil {
		return err
	}

	builder.SetStatus("BuildGlibc", true)

	return nil
}

// BuildLinux compiles the linux source.
func (builder *Builder) BuildLinux() error {
	if builder.GetStatus("BuildLinux") {
		return nil
	}

	builder.BuildSetenv()
	fmt.Println("downloading linux")
	filename, err := builder.DownloadLinux()
	if err != nil {
		return err
	}

	linuxSrc := builder.GetBuildDir("linux")
	linuxSrcOld := builder.GetBuildDir("linux-" + builder.LinuxBuild.LinuxVersion)
	if !exists(linuxSrc) {
		fmt.Println("extracting linux")
		if err := copyAllGit(filename, linuxSrcOld); err != nil {
			return err
		}
		if err := os.Rename(linuxSrcOld, linuxSrc); err != nil {
			return err
		}
	}

	defconfigFilename := filepath.Join(linuxSrc, "arch", builder.Meta.Arch, "configs", "lht_defconfig")
	if builder.Meta.Defconfig != "" {
		fmt.Println("writing defconfig")
		if err := ioutil.WriteFile(defconfigFilename, []byte(builder.Meta.Defconfig), 0644); err != nil {
			return err
		}
	}

	fmt.Println("making defconfig")
	if builder.Meta.Defconfig == "" {
		if err := execAt(linuxSrc, "make", "defconfig"); err != nil {
			return err
		}
	} else {
		if err := execAt(linuxSrc, "make", "lht_defconfig"); err != nil {
			return err
		}
	}

	fmt.Println("making linux")
	if err := execAt(linuxSrc, "make", "-j", viper.GetString("Threads")); err != nil {
		return err
	}

	fmt.Println("installing headers")
	headersDir := builder.GetBuildDir("headers")
	if !exists(headersDir) {
		if err := execAt(linuxSrc, "make", "headers_install",
			"INSTALL_HDR_PATH=../headers"); err != nil {
			return err
		}
	}

	builder.SetStatus("BuildLinux", true)

	return nil
}
