package project

import "path/filepath"

// CreateRootFS builds a root file system.
func (proj Project) CreateRootFS(version string) error {
	versionDir := filepath.Join(proj.Path(), version)
	return doMountCopy(versionDir)
}
