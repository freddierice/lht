package project

import (
	"encoding/json"
	"fmt"
	"os"
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
	FsSize         uint64 `json:"fsSize"`
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
