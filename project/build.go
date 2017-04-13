package project

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/viper"
)

// BuildSetenv sets the environment variables needed to build the projects.
func (proj Project) BuildSetenv() {
	if proj.Arch != "" {
		os.Setenv("ARCH", proj.Arch)
		os.Setenv("CROSS_COMPILE", proj.Target+"-")
	}
}

// BuildAll builds all of the projects for a specific version of linux.
func (proj Project) BuildAll(version string) error {
	if err := proj.BuildLinux(version); err != nil {
		return fmt.Errorf("could not build linux: %v", err)
	}

	if err := proj.BuildVulnKo(version); err != nil {
		return fmt.Errorf("could not build vuln-ko: %v", err)
	}

	if err := proj.BuildGlibc(version); err != nil {
		return fmt.Errorf("could not build glibc: %v", err)
	}

	if err := proj.BuildBusyBox(version); err != nil {
		return fmt.Errorf("could not build busybox: %v", err)
	}

	return nil
}

// BuildVulnKo builds the vuln-ko kernel module.
func (proj Project) BuildVulnKo(version string) error {
	proj.BuildSetenv()

	fmt.Println("downloading vuln-ko")
	vulnDir, err := proj.DownloadVulnKo()
	if err != nil {
		return err
	}

	vulnDirNew := filepath.Join(proj.Path(), version, "vuln-ko")
	if !exists(vulnDirNew) {
		// TODO: put this in code
		cmd := exec.Command("cp", "-rf", vulnDir, vulnDirNew)
		if err := cmd.Run(); err != nil {
			os.RemoveAll(vulnDirNew)
			return err
		}
	}

	os.Setenv("BUILD_DIR", filepath.Join(proj.Path(), version, "linux"))
	fmt.Println("building vuln-ko")
	return execAt(vulnDirNew, "make")
}

// BuildBusyBox builds the busybox.
func (proj Project) BuildBusyBox(version string) error {
	proj.BuildSetenv()
	fmt.Println("downloading busybox")
	filename, err := DownloadBusyBox(proj.BusyBoxVersion)
	if err != nil {
		return err
	}

	versionPath := filepath.Join(proj.Path(), version)
	if err := os.MkdirAll(versionPath, 0755); err != nil {
		return err
	}

	busyBoxPath := filepath.Join(versionPath, "busybox")
	busyBoxPathOld := filepath.Join(versionPath, "busybox-"+proj.BusyBoxVersion)
	busyBoxInstallPath := filepath.Join(busyBoxPath, "_install")
	if exists(busyBoxInstallPath) {
		return nil
	}
	if !exists(busyBoxPath) {
		fmt.Println("extracting busybox")
		//TODO: replace with in-code solution
		cmd := exec.Command("tar", "-C", versionPath, "-xf", filename)
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

	return nil
}

// BuildGlibc builds the Glibc project for a given linux version.
func (proj Project) BuildGlibc(version string) error {
	proj.BuildSetenv()
	fmt.Println("downloading glibc")
	filename, err := proj.DownloadGlibc()
	if err != nil {
		return err
	}

	versionPath := filepath.Join(proj.Path(), version)
	if err := os.MkdirAll(versionPath, 0755); err != nil {
		return err
	}

	fmt.Println("extracting glibc")
	glibcPath := filepath.Join(versionPath, "glibc")
	glibcPathOld := filepath.Join(versionPath, "glibc-"+proj.GlibcVersion)
	glibcBuildPath := filepath.Join(versionPath, "glibc-build")
	sysroot := filepath.Join(versionPath, "sysroot")
	if exists(sysroot) {
		return nil
	}
	if !exists(glibcPath) {
		//TODO: replace with in-code solution
		cmd := exec.Command("tar", "-C", versionPath, "-xf", filename)
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
	os.Setenv("CC", proj.Target+"-gcc")
	os.Setenv("CXX", proj.Target+"-g++")
	os.Setenv("AR", proj.Target+"-ar")
	os.Setenv("RANLIB", proj.Target+"-ranlib")

	configparams := fmt.Sprintf("slibdir=/lib\nrtlddir=/lib\nsbindir=/bin\nrootsbindir=/bin\nbuild-programs=no\n")
	configparamsFilename := filepath.Join(glibcPath, "configparams")
	if err := ioutil.WriteFile(configparamsFilename, []byte(configparams), 0644); err != nil {
		return err
	}

	headersPath := filepath.Join(versionPath, "headers", "include")
	err = execAt(glibcBuildPath, "../glibc/configure", "--prefix=/", "--libdir=/lib", "--libexecdir=/lib",
		"--enable-add-ons", "--enable-kernel=2.6.32", "--enable-lock-elision",
		"--enable-stackguard-randomization", "--enable-bind-now", "--disable-profile",
		"--disable-multi-arch", "--disable-werror", "--target="+proj.Target,
		"--host="+proj.Target, "--build="+proj.Host, "--with-headers="+headersPath)
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

	return nil
}

// BuildLinux compiles the linux source.
func (proj Project) BuildLinux(version string) error {

	proj.BuildSetenv()
	fmt.Println("downloading linux")
	filename, err := proj.DownloadLinux(version)
	if err != nil {
		return err
	}

	versionPath := filepath.Join(proj.Path(), version)
	if err := os.MkdirAll(versionPath, 0755); err != nil {
		return err
	}

	linuxSrc := filepath.Join(versionPath, "linux")
	linuxSrcOld := filepath.Join(versionPath, "linux-"+version)
	if !exists(linuxSrc) {
		fmt.Println("extracting linux")
		// TODO: replace with in-code solution
		cmd := exec.Command("tar", "-C", versionPath, "-xf", filename)
		if err := cmd.Run(); err != nil {
			return err
		}
		if err := os.Rename(linuxSrcOld, linuxSrc); err != nil {
			return err
		}
	}

	defconfigFilename := filepath.Join(linuxSrc, "arch", proj.Arch, "configs", "lht_defconfig")
	if proj.Defconfig != "" {
		fmt.Println("writing defconfig")
		if err := ioutil.WriteFile(defconfigFilename, []byte(proj.Defconfig), 0644); err != nil {
			return err
		}
	}

	fmt.Println("making defconfig")
	if proj.Defconfig == "" {
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
	headersDir := filepath.Join(versionPath, "headers")
	if !exists(headersDir) {
		if err := execAt(linuxSrc, "make", "headers_install",
			"INSTALL_HDR_PATH=../headers"); err != nil {
			return err
		}
	}

	return nil
}
