package project

// LinuxBuild contains information about a specific build within a project.
type LinuxBuild struct {
	Name         string          `json:"name"`
	LinuxVersion string          `json:"linuxVersion"`
	Tag          string          `json:"tag"`
	Status       map[string]bool `json:"status"`
}
