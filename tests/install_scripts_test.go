package tests

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstallPowerShellFromReleaseMetadata(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell installer smoke test only runs on Windows")
	}

	binaryPath := filepath.Join("..", "promptc.exe")
	binaryData, err := os.ReadFile(binaryPath)
	if err != nil {
		t.Fatalf("read test binary %s: %v", binaryPath, err)
	}

	archiveName := "promptc_1.2.3_windows_amd64.zip"
	archiveBytes := createZipArchive(t, "promptc.exe", binaryData)
	checksums := checksumFile(archiveName, archiveBytes)
	server := newReleaseServer(archiveName, checksums, archiveBytes)
	defer server.Close()

	installDir := t.TempDir()
	cmd := exec.Command(
		"powershell.exe",
		"-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-File", filepath.Join("..", "install.ps1"),
	)
	cmd.Env = append(os.Environ(),
		"PROMPTC_REPO=test/repo",
		"PROMPTC_RELEASES_API_BASE="+server.URL+"/repos",
		"PROMPTC_RELEASE_TAG=v1.2.3",
		"PROMPTC_INSTALL_DIR="+installDir,
		"PROMPTC_SKIP_PATH_UPDATE=1",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install.ps1 failed: %v\n%s", err, output)
	}

	installedPath := filepath.Join(installDir, "promptc.exe")
	if _, err := os.Stat(installedPath); err != nil {
		t.Fatalf("installed binary missing: %v", err)
	}
	if !strings.Contains(string(output), "Installed promptc") {
		t.Fatalf("installer output missing success message:\n%s", output)
	}
}

func TestInstallShellFromReleaseMetadata(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell installer smoke test runs on unix-like environments")
	}

	binaryData := []byte("#!/bin/sh\necho 'promptc test build'\n")
	archiveName := "promptc_1.2.3_linux_amd64.tar.gz"
	archiveBytes := createTarGzArchive(t, "promptc", binaryData, 0o755)
	checksums := checksumFile(archiveName, archiveBytes)
	server := newReleaseServer(archiveName, checksums, archiveBytes)
	defer server.Close()

	installDir := t.TempDir()
	cmd := exec.Command("sh", filepath.Join("..", "install.sh"))
	cmd.Env = append(os.Environ(),
		"PROMPTC_REPO=test/repo",
		"PROMPTC_RELEASES_API_BASE="+server.URL+"/repos",
		"PROMPTC_RELEASE_TAG=v1.2.3",
		"PROMPTC_INSTALL_DIR="+installDir,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install.sh failed: %v\n%s", err, output)
	}

	installedPath := filepath.Join(installDir, "promptc")
	data, err := os.ReadFile(installedPath)
	if err != nil {
		t.Fatalf("installed binary missing: %v", err)
	}
	if !bytes.Equal(data, binaryData) {
		t.Fatalf("installed file contents do not match test archive")
	}
	if !strings.Contains(string(output), "Installed promptc") {
		t.Fatalf("installer output missing success message:\n%s", output)
	}
}

func createZipArchive(t *testing.T, name string, data []byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(name)
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	if _, err := w.Write(data); err != nil {
		t.Fatalf("write zip entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip archive: %v", err)
	}
	return buf.Bytes()
}

func createTarGzArchive(t *testing.T, name string, data []byte, mode int64) []byte {
	t.Helper()

	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)
	if err := tw.WriteHeader(&tar.Header{
		Name: name,
		Mode: mode,
		Size: int64(len(data)),
	}); err != nil {
		t.Fatalf("write tar header: %v", err)
	}
	if _, err := tw.Write(data); err != nil {
		t.Fatalf("write tar data: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
	return buf.Bytes()
}

func checksumFile(name string, archive []byte) []byte {
	sum := sha256.Sum256(archive)
	return []byte(hex.EncodeToString(sum[:]) + "  " + name + "\n")
}

func newReleaseServer(archiveName string, checksums []byte, archive []byte) *httptest.Server {
	type asset struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	}
	type release struct {
		TagName string  `json:"tag_name"`
		Assets  []asset `json:"assets"`
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	baseURL := "http://" + listener.Addr().String()

	mux := http.NewServeMux()
	mux.HandleFunc("/downloads/"+archiveName, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archive)
	})
	mux.HandleFunc("/downloads/checksums.txt", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(checksums)
	})
	mux.HandleFunc("/repos/test/repo/releases/tags/v1.2.3", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(release{
			TagName: "v1.2.3",
			Assets: []asset{
				{
					Name:               archiveName,
					BrowserDownloadURL: baseURL + "/downloads/" + archiveName,
				},
				{
					Name:               "checksums.txt",
					BrowserDownloadURL: baseURL + "/downloads/checksums.txt",
				},
			},
		})
	})

	server := httptest.NewUnstartedServer(mux)
	server.Listener = listener
	server.Start()
	return server
}
