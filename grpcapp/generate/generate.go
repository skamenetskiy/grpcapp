package generate

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/skamenetskiy/grpcapp/grpcapp/h"
)

func Run(_ []string) {
	if err := generateProto(); err != nil {
		h.Die("failed to generate proto: %s", err)
	}
}

const (
	vendorDir     = ".proto"
	protocVersion = "21.6"
)

func generate(_ []string) {
	if err := generateProto(); err != nil {
		h.Die("failed to generate proto: %s", err)
	}
}

func generateProto() error {
	detectedOs := getOS()
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if exists, empty := h.DirInfo(vendorDir); !exists || empty {
		h.Mkdir(vendorDir)
	}
	if exists, empty := h.DirInfo(filepath.Join(vendorDir, "protoc")); !exists || empty {
		h.Mkdir(filepath.Join(vendorDir, "protoc"))
		protocUri := fmt.Sprintf(
			"https://github.com/protocolbuffers/protobuf/releases/download/v%s/protoc-%s-%s.zip",
			protocVersion, protocVersion, detectedOs)
		if err = downloadAndUnzip(protocUri,
			filepath.Join(vendorDir, "protoc.zip"),
			filepath.Join(vendorDir, "protoc")); err != nil {
			return err
		}
	}
	if exists, empty := h.DirInfo(filepath.Join(vendorDir, "bin")); !exists || empty {
		h.Mkdir(filepath.Join(vendorDir, "bin"))
		install := func(pkg string) error {
			cmd := exec.Command("go", "install", pkg)
			cmd.Env = append(os.Environ(), fmt.Sprintf("GOBIN=%s", filepath.Join(wd, vendorDir, "bin")))
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
			return nil
		}
		for _, pkg := range []string{
			"google.golang.org/protobuf/cmd/protoc-gen-go@v1.26",
			"google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1",
		} {
			if err := install(pkg); err != nil {
				return err
			}
		}
	}
	cmd := exec.Command(filepath.Join(vendorDir, "protoc", "bin", "protoc"),
		"--go_out=.",
		"--go_opt=paths=source_relative",
		"--go-grpc_out=.",
		"--go-grpc_opt=paths=source_relative",
		"--proto_path=.",
		"pkg/service.proto",
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s", filepath.Join(wd, vendorDir, "bin")))
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func download(uri, target string) error {
	file, err := os.Create(target)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	res, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	if _, err = io.Copy(file, res.Body); err != nil {
		return err
	}
	return nil
}

func downloadAndUnzip(uri, targetZip, dst string) error {
	if err := download(uri, targetZip); err != nil {
		return err
	}
	archive, err := zip.OpenReader(targetZip)
	if err != nil {
		return err
	}
	defer func() { _ = archive.Close() }()
	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)
		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path")
		}
		if f.FileInfo().IsDir() {
			h.Mkdir(filePath)
			continue
		}
		h.Mkdir(filepath.Dir(filePath))
		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		fileInArchive, err := f.Open()
		if err != nil {
			return err
		}
		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return err
		}
		_ = dstFile.Close()
		_ = fileInArchive.Close()
	}
	_ = archive.Close()
	return os.Remove(targetZip)
}

func getOS() string {
	switch runtime.GOOS {
	case "linux":
		return "linux-x86_64"
	case "windows":
		return "win64"
	case "darwin":
		return "osx-x86_64"
	default:
		return "unknown"
	}
}
