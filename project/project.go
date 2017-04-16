package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nightlyone/lockfile"
)

// Project holds variables needed to compile a project.
type Project struct {
	Name   string
	Builds map[string]LinuxBuild `json:"build"`
	lock   lockfile.Lockfile     `json:"-"`

	Meta `json:"projectMeta"`
}

// Meta holds metadata that is consistent throughout a project.
type Meta struct {
	Arch           string `json:"arch"`
	Target         string `json:"target"`
	Host           string `json:"host"`
	Defconfig      string `json:"defconfig"`
	GlibcVersion   string `json:"glibcVersion"`
	BusyBoxVersion string `json:"busyBoxVersion"`
	FsSize         uint64 `json:"fsSize"`
}

// Create creates a new project with a name, and returns a non-nil error on
// failure.
func Create(name string) (*Project, error) {
	projectRoot := filepath.Join(getRootDir(), name)
	if f, err := os.Open(projectRoot); err == nil {
		f.Close()
		return nil, fmt.Errorf("project exists")
	}
	if err := os.MkdirAll(projectRoot, 0750); err != nil {
		return nil, err
	}

	proj := &Project{
		Name:   name,
		Builds: map[string]LinuxBuild{},
	}
	return proj, proj.Commit()
}

// Open reads a project configuration, and returns it as a Project. Open only
// returns an error if the configuration cannot be parsed, or if the
// configuration could not be found.
func Open(name string) (*Project, error) {

	readConf := &Project{}

	projectRoot := filepath.Join(getRootDir(), name)
	projectConf := filepath.Join(getConfDir(), name+".json")
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

	if readConf.Builds == nil {
		readConf.Builds = map[string]LinuxBuild{}
	}

	return readConf, nil
}

// Delete removes a project.
func (proj *Project) Delete() error {
	proj.Close()
	os.RemoveAll(proj.Path())
	os.Remove(proj.ConfigPath())
	return nil
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

// GetBuilder takes a project and a buildName to produce a Builder. If the
// buildName exists, then we create it, otherwise produce an error.
func (proj *Project) GetBuilder(buildName string) (*Builder, error) {
	build, ok := proj.Builds[buildName]
	if !ok {
		return nil, fmt.Errorf("build doesn't exist")
	}

	if build.Status == nil {
		build.Status = map[string]bool{}
	}

	builder := &Builder{
		RootDir:     getRootDir(),
		ProjectDir:  proj.Path(),
		BuildDir:    filepath.Join(proj.Path(), buildName),
		DownloadDir: getDownloadDir(),
		Meta:        proj.Meta,
		LinuxBuild:  build,
	}

	return builder, nil
}

// Path gets the project's root directory.
func (proj *Project) Path() string {
	return filepath.Join(getRootDir(), proj.Name)
}

// ConfigPath gets the project's configuration file.
func (proj *Project) ConfigPath() string {
	return filepath.Join(getConfDir(), proj.Name+".json")
}

// Commit takes the in memory version of the project and writes it to the
// configuration file on disk.
func (proj *Project) Commit() error {

	projectConf := proj.ConfigPath()
	projectConfTmp := filepath.Join(filepath.Dir(projectConf), "."+filepath.Base(projectConf))

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
