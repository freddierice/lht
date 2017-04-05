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
	Name               string
	Arch               string `json:"arch"`
	CrossCompilePrefix string `json:"crossCompile"`
	Defconfig          string `json:"defconfig"`
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
	}
	if proj.CrossCompilePrefix != "" {
		os.Setenv("CROSS_COMPILE", proj.CrossCompilePrefix)
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

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(vulnDirNew); err != nil {
		return err
	}

	fmt.Println("building vuln-ko")
	if err := exec.Command("make").Run(); err != nil {
		os.Chdir(wd)
		return err
	}

	return os.Chdir(wd)
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

	fmt.Println("extracting linux")
	linuxSrc := filepath.Join(versionPath, "linux")
	linuxSrcOld := filepath.Join(versionPath, "linux-"+version)
	if !exists(linuxSrc) {
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
	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(linuxSrc); err != nil {
		return err
	}

	if proj.Defconfig != "" {
		fmt.Println("making defconfig")
		cmd := exec.Command("make", "lht_defconfig")
		if err := cmd.Run(); err != nil {
			out, err2 := cmd.Output()
			if err2 != nil {
				fmt.Println("could not print output, but there was an error during the make")
				return err
			}
			fmt.Println(string(out))
			return err
		}
	}

	fmt.Println("making linux")
	cmd := exec.Command("make", "-j", viper.GetString("Threads"))
	if err := cmd.Run(); err != nil {
		out, err2 := cmd.Output()
		if err2 != nil {
			fmt.Println("could not print output, but there was an error during the make")
			return err
		}
		fmt.Println(string(out))
		return err
	}

	os.Chdir(oldWorkingDirectory)

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
