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

var majorRegex *regexp.Regexp = regexp.MustCompile("^[0-9]*")

func (proj Project) DownloadVulnKo() (string, error) {
	downloadDirectory, err := getDownloadDirectory()
	if err != nil {
		return "", err
	}
	vulnKoSrc := filepath.Join(downloadDirectory, "vuln-ko")
	if exists(vulnKoSrc) {
		return vulnKoSrc, nil
	}

	// TODO: do this in code
	cmd := exec.Command("git", "clone", "https://github.com/freddierice/vuln-ko.git", vulnKoSrc)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return vulnKoSrc, nil
}

// DownloadGlibc downloads a version of linux, and returns the filepath
func (proj Project) DownloadGlibc() (string, error) {
	glibcArchiveFilename := fmt.Sprintf("glibc-%v.tar.gz", proj.GlibcVersion)
	glibcArchiveUrl := fmt.Sprintf("https://ftp.gnu.org/gnu/glibc/glibc-%v.tar.gz", proj.GlibcVersion)
	return download(glibcArchiveFilename, glibcArchiveUrl)
}

// DownloadLinux downloads a version of linux, and returns the filepath.
func (proj Project) DownloadLinux(version string) (string, error) {
	versionMajor := majorRegex.FindString(version)

	linuxFilename := fmt.Sprintf("linux-%v.tar.xz", version)
	linuxUrl := fmt.Sprintf("https://cdn.kernel.org/pub/linux/kernel/v%v.x/linux-%v.tar.xz", versionMajor, version)
	return download(linuxFilename, linuxUrl)
}

// DownloadBuysBox downloads BusyBox with version and returns its filepath.
func DownloadBusyBox(version string) (string, error) {

	busyBoxFilename := fmt.Sprintf("busybox-%v.tar.bz2", version)
	busyBoxUrl := fmt.Sprintf("https://busybox.net/downloads/busybox-%v.tar.bz2", version)
	return download(busyBoxFilename, busyBoxUrl)
}

// getDownloadDirectory returns the folder used for downloading archives to be
// built by projects. If the directory does not exist, it will be made. Returns
// an error if the directory could not be created.
func getDownloadDirectory() (string, error) {
	rootDirectory := viper.GetString("RootDirectory")
	downloadDirectory := filepath.Join(rootDirectory, ".downloads")
	return downloadDirectory, os.MkdirAll(downloadDirectory, 0755)
}

// download attempts to save the file at fileUrl to filename in the download
// directory. download will return the full path to the file after it has
// downloaded completely, or return an error.
func download(filename, fileUrl string) (string, error) {
	downloadDirectory, err := getDownloadDirectory()
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(downloadDirectory, filename)
	if exists(filePath) {
		return filePath, nil
	}

	downloadFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer downloadFile.Close()

	resp, err := http.Get(fileUrl)
	if err != nil {
		downloadFile.Close()
		os.Remove(filePath)
		return "", err
	}

	if _, err := io.Copy(downloadFile, resp.Body); err != nil {
		downloadFile.Close()
		os.Remove(filePath)
		return "", err
	}

	return filePath, nil
}
