package project

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/viper"
)

type Project struct {
	Name           string
	Arch           string `json:"arch"`
	Target         string `json:"target"`
	Host           string `json:"host"`
	Defconfig      string `json:"defconfig"`
	GlibcVersion   string `json:"glibcVersion"`
	BusyBoxVersion string `json:"busyBoxVersion"`
}

func Create(name string) (Project, error) {
	projectRoot := filepath.Join(viper.GetString("RootDirectory"), name)
	if f, err := os.Open(projectRoot); err == nil {
		f.Close()
		return Project{}, fmt.Errorf("project exists")
	}
	if err := os.MkdirAll(projectRoot, 0750); err != nil {
		return Project{}, err
	}

	proj := Project{Name: name}

	return proj, proj.Write()
}

func Open(name string) (Project, error) {
	readConf := Project{}

	projectRoot := filepath.Join(viper.GetString("RootDirectory"), name)
	projectConf := filepath.Join(projectRoot, "conf.json")

	projectConfFile, err := os.Open(projectConf)
	if err != nil {
		return Project{}, err
	}
	defer projectConfFile.Close()

	if err := json.NewDecoder(projectConfFile).Decode(&readConf); err != nil {
		return Project{}, err
	}

	return readConf, nil
}

func (proj Project) BuildSetenv() {
	if proj.Arch != "" {
		os.Setenv("ARCH", proj.Arch)
		os.Setenv("CROSS_COMPILE", proj.Target+"-")
	}
}

func (proj Project) BuildVulnKo(version string) error {
	proj.BuildSetenv()

	fmt.Println("downloading vuln-ko")
	vulnDir, err := DownloadVulnKo()
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

func (proj Project) BuildGlibc(version string) error {
	proj.BuildSetenv()
	fmt.Println("downloading glibc")
	filename, err := DownloadGlibc(proj.GlibcVersion)
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

func (proj Project) BuildLinux(version string) error {

	proj.BuildSetenv()
	fmt.Println("downloading linux")
	filename, err := DownloadLinux(version)
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

	if proj.Defconfig != "" {
		fmt.Println("making defconfig")
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

func (proj Project) Path() string {
	return filepath.Join(viper.GetString("RootDirectory"), proj.Name)
}

func (proj Project) Write() error {
	projectConf := filepath.Join(proj.Path(), "conf.json")
	projectConfTmp := filepath.Join(proj.Path(), ".conf.json")

	projectConfFile, err := os.Create(projectConfTmp)
	if err != nil {
		return err
	}

	if err := json.NewEncoder(projectConfFile).Encode(proj); err != nil {
		return err
	}
	if err := projectConfFile.Close(); err != nil {
		return err
	}

	return os.Rename(projectConfTmp, projectConf)
}
