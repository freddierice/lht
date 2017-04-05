package project

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/spf13/viper"
)

var linuxRegex *regexp.Regexp = regexp.MustCompile("^[0-9]*\\.[0-9]*\\.[0-9]*$")
var majorRegex *regexp.Regexp = regexp.MustCompile("^[0-9]*")

func DownloadVulnKo() (string, error) {
	// TODO: do this in code
	downloadDirectory, err := getDownloadDirectory()
	if err != nil {
		return "", err
	}
	vulnKoSrc := filepath.Join(downloadDirectory, "vuln-ko")
	if exists(vulnKoSrc) {
		return vulnKoSrc, nil
	}

	cmd := exec.Command("git", "clone", "https://github.com/freddierice/vuln-ko.git", vulnKoSrc)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return vulnKoSrc, nil
}

// Download a version of linux, and returns the filepath
func DownloadLinux(version string) (string, error) {
	if !linuxRegex.MatchString(version) {
		return "", fmt.Errorf("invalid linux version")
	}
	versionMajor := majorRegex.FindString(version)

	downloadDirectory, err := getDownloadDirectory()
	if err != nil {
		return "", err
	}

	linuxFilename := filepath.Join(downloadDirectory, fmt.Sprintf("linux-%v.tar.xz", version))

	// check if already downloaded
	if f, err := os.Open(linuxFilename); err == nil {
		f.Close()
		return linuxFilename, nil
	}

	linuxFile, err := os.Create(linuxFilename)
	if err != nil {
		return linuxFilename, err
	}

	linuxUrl := fmt.Sprintf("https://cdn.kernel.org/pub/linux/kernel/v%v.x/linux-%v.tar.xz", versionMajor, version)
	resp, err := http.Get(linuxUrl)
	if err != nil {
		linuxFile.Close()
		os.Remove(linuxFilename)
		return linuxFilename, err
	}

	if _, err := io.Copy(linuxFile, resp.Body); err != nil {
		linuxFile.Close()
		os.Remove(linuxFilename)
		return linuxFilename, err
	}

	return linuxFilename, nil
}

func getDownloadDirectory() (string, error) {
	rootDirectory := viper.GetString("RootDirectory")
	downloadDirectory := filepath.Join(rootDirectory, ".downloads")
	return downloadDirectory, os.MkdirAll(downloadDirectory, 0755)
}
