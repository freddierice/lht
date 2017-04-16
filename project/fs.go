package project

import "path/filepath"

// Compile builds a root file system
func (proj Project) CreateRootFS(version string) error {
	versionDir := filepath.Join(proj.Path(), version)
	return doMountCopy(versionDir)
}
