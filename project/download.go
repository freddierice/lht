package project

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// DownloadVulnKo downloads the vuln-ko project to the download directory and
// returns the filepath to the git repository.
func (builder *Builder) DownloadVulnKo() (string, error) {
	vulnKoSrc := filepath.Join(builder.DownloadDir, "vuln-ko")
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
func (builder *Builder) DownloadGlibc() (string, error) {
	glibcArchiveFilename := fmt.Sprintf("glibc-%v.tar.gz", builder.Meta.GlibcVersion)
	glibcArchiveURL := fmt.Sprintf("https://ftp.gnu.org/gnu/glibc/glibc-%v.tar.gz", builder.Meta.GlibcVersion)
	return download(glibcArchiveFilename, builder.DownloadDir, glibcArchiveURL)
}

// DownloadLinux downloads a version of linux, and returns the filepath.
func (builder *Builder) DownloadLinux() (string, error) {

	linuxDir := filepath.Join(builder.DownloadDir, "linux-stable")
	if !exists(linuxDir) {
		cmd := exec.Command("git", "clone", "git://git.kernel.org/pub/scm/linux/kernel/git/stable/linux-stable.git", linuxDir)
		if err := cmd.Run(); err != nil {
			return "", err
		}
	}
	linuxTag := fmt.Sprintf("%v", builder.LinuxBuild.Tag)

	wd, err := os.Getwd()
	if err != nil {
		return "", nil
	}

	os.Chdir(linuxDir)
	cmd := exec.Command("git", "checkout", linuxTag)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	os.Chdir(wd)

	return linuxDir, nil
}

// DownloadBusyBox downloads BusyBox with version and returns its filepath.
func (builder *Builder) DownloadBusyBox() (string, error) {

	busyBoxFilename := fmt.Sprintf("busybox-%v.tar.bz2", builder.Meta.BusyBoxVersion)
	busyBoxURL := fmt.Sprintf("https://busybox.net/downloads/busybox-%v.tar.bz2", builder.Meta.BusyBoxVersion)
	return download(busyBoxFilename, builder.DownloadDir, busyBoxURL)
}

// download attempts to save the file at fileUrl to filename in the download
// directory. download will return the full path to the file after it has
// downloaded completely, or return an error.
func download(filename, downloadDir, fileURL string) (string, error) {
	filePath := filepath.Join(downloadDir, filename)
	if exists(filePath) {
		return filePath, nil
	}

	downloadFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer downloadFile.Close()

	resp, err := http.Get(fileURL)
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
