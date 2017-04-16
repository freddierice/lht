package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nightlyone/lockfile"
	"github.com/spf13/viper"
)

// Project holds variables needed to compile a project.
type Project struct {
	Name           string
	Arch           string                `json:"arch"`
	Target         string                `json:"target"`
	Host           string                `json:"host"`
	Defconfig      string                `json:"defconfig"`
	GlibcVersion   string                `json:"glibcVersion"`
	BusyBoxVersion string                `json:"busyBoxVersion"`
	FsSize         uint64                `json:"fsSize"`
	Builds         map[string]LinuxBuild `json:"build"`

	lock lockfile.Lockfile `json:"-"`
}

// Create creates a new project with a name, and returns a non-nil error on
// failure.
func Create(name string) (*Project, error) {
	projectRoot := filepath.Join(viper.GetString("RootDirectory"), name)
	if f, err := os.Open(projectRoot); err == nil {
		f.Close()
		return nil, fmt.Errorf("project exists")
	}
	if err := os.MkdirAll(projectRoot, 0750); err != nil {
		return nil, err
	}

	proj := &Project{Name: name}
	return proj, proj.Commit()
}

// Open reads a project configuration, and returns it as a Project. Open only
// returns an error if the configuration cannot be parsed, or if the
// configuration could not be found.
func Open(name string) (*Project, error) {

	readConf := &Project{}

	projectRoot := filepath.Join(viper.GetString("RootDirectory"), name)
	projectConf := filepath.Join(projectRoot, "conf.json")
	projectLock := filepath.Join(projectRoot, ".lock")

	// take the lock
	retries := 0
	lock, err := lockfile.New(projectLock)
	for err != nil && retries < 5 {
		retries++
		switch err {
		case lockfile.ErrNotExist:
			time.Sleep(time.Millisecond * 200)
			lock, err = lockfile.New(projectLock)
		case lockfile.ErrBusy:
			return nil, fmt.Errorf("project in use by another process")
		case lockfile.ErrNeedAbsPath:
			panic("project root is not absoute")
		case lockfile.ErrDeadOwner:
			fallthrough
		case lockfile.ErrInvalidPid:
			os.Remove(projectLock)
			lock, err = lockfile.New(projectLock)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("could not get lock for project")
	}

	readConf.lock = lock

	projectConfFile, err := os.Open(projectConf)
	if err != nil {
		return nil, err
	}
	defer projectConfFile.Close()

	if err := json.NewDecoder(projectConfFile).Decode(&readConf); err != nil {
		return nil, err
	}

	return readConf, nil
}

// Close cleans up after a project when it is no longer in use. For the time
// being this is only the project's file lock.
func (proj *Project) Close() error {
	proj.Commit()
	if err := proj.lock.Unlock(); err != nil {
		return err
	}

	return nil
}

// Path gets the project's root directory.
func (proj *Project) Path() string {
	return filepath.Join(viper.GetString("RootDirectory"), proj.Name)
}

// Commit takes the in memory version of the project and writes it to the
// configuration file on disk.
func (proj *Project) Commit() error {
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
