package main

import (
	"os"

	"github.com/skamenetskiy/grpcapp/grpcapp/create"
	"github.com/skamenetskiy/grpcapp/grpcapp/generate"
	"github.com/skamenetskiy/grpcapp/grpcapp/help"
)

func main() {
	var (
		args    = os.Args[1:]
		command string
	)
	if len(args) == 0 {
		command = "help"
	} else {
		command = args[0]
		args = args[1:]
	}
	var cmd func([]string)
	switch command {
	case "create":
		cmd = create.Run
	case "generate":
		cmd = generate.Run
	case "help":
		cmd = help.Run
	default:
		cmd = help.Run
	}
	cmd(args)
}

////go:embed create/skeleton
//var skeletonFS embed.FS

//const (
//	filePerm = 0664
//	dirPerm  = 0775
//
//	vendorDir     = ".proto"
//	protocVersion = "21.6"
//)

//func create(args []string) {
//	if len(args) == 0 {
//		h.Die("application name not specified")
//	}
//	name := strings.TrimSpace(args[0])
//	if name == "" {
//		h.Die("invalid application name")
//	}
//	dirExists, dirEmpty := dirInfo(name)
//	if !dirEmpty {
//		h.Die("directory '%s' exists and is not empty", name)
//	}
//	if !dirExists {
//		mkdir(name)
//	}
//	goModData := fmt.Sprintf("module %s\n\ngo 1.19\n\nrequire (\n\t"+
//		"github.com/skamenetskiy/grpcapp v0.0.0-20220927150441-ce3a333817c0\n)", name)
//	if err := os.WriteFile(filepath.Join(name, "go.mod"), []byte(goModData), filePerm); err != nil {
//		h.Die("failed to write go.mod: %s", err)
//	}
//	if err := extract(name); err != nil {
//		h.Die("failed to extract application files: %s", err)
//	}
//}

//func generate(_ []string) {
//	if err := generateProto(); err != nil {
//		h.Die("failed to generate proto: %s", err)
//	}
//}

//func generateProto() error {
//	detectedOs := getOS()
//	wd, err := os.Getwd()
//	if err != nil {
//		return err
//	}
//	if exists, empty := dirInfo(vendorDir); !exists || empty {
//		mkdir(vendorDir)
//	}
//	if exists, empty := dirInfo(filepath.Join(vendorDir, "protoc")); !exists || empty {
//		mkdir(filepath.Join(vendorDir, "protoc"))
//		protocUri := fmt.Sprintf(
//			"https://github.com/protocolbuffers/protobuf/releases/download/v%s/protoc-%s-%s.zip",
//			protocVersion, protocVersion, detectedOs)
//		if err = downloadAndUnzip(protocUri,
//			filepath.Join(vendorDir, "protoc.zip"),
//			filepath.Join(vendorDir, "protoc")); err != nil {
//			return err
//		}
//	}
//	if exists, empty := dirInfo(filepath.Join(vendorDir, "bin")); !exists || empty {
//		mkdir(filepath.Join(vendorDir, "bin"))
//		install := func(pkg string) error {
//			cmd := exec.Command("go", "install", pkg)
//			cmd.Env = append(os.Environ(), fmt.Sprintf("GOBIN=%s", filepath.Join(wd, vendorDir, "bin")))
//			cmd.Stderr = os.Stderr
//			if err := cmd.Run(); err != nil {
//				return err
//			}
//			return nil
//		}
//		for _, pkg := range []string{
//			"google.golang.org/protobuf/cmd/protoc-gen-go@v1.26",
//			"google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1",
//		} {
//			if err := install(pkg); err != nil {
//				return err
//			}
//		}
//	}
//	cmd := exec.Command(filepath.Join(vendorDir, "protoc", "bin", "protoc"),
//		"--go_out=.",
//		"--go_opt=paths=source_relative",
//		"--go-grpc_out=.",
//		"--go-grpc_opt=paths=source_relative",
//		"--proto_path=.",
//		"pkg/service.proto",
//	)
//	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s", filepath.Join(wd, vendorDir, "bin")))
//	cmd.Stderr = os.Stderr
//	return cmd.Run()
//}

//func download(uri, target string) error {
//	file, err := os.Create(target)
//	if err != nil {
//		return err
//	}
//	defer func() { _ = file.Close() }()
//	res, err := http.Get(uri)
//	if err != nil {
//		return err
//	}
//	defer func() { _ = res.Body.Close() }()
//	if _, err = io.Copy(file, res.Body); err != nil {
//		return err
//	}
//	return nil
//}

//func downloadAndUnzip(uri, targetZip, dst string) error {
//	if err := download(uri, targetZip); err != nil {
//		return err
//	}
//	archive, err := zip.OpenReader(targetZip)
//	if err != nil {
//		return err
//	}
//	defer func() { _ = archive.Close() }()
//	for _, f := range archive.File {
//		filePath := filepath.Join(dst, f.Name)
//		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
//			return fmt.Errorf("invalid file path")
//		}
//		if f.FileInfo().IsDir() {
//			mkdir(filePath)
//			continue
//		}
//		mkdir(filepath.Dir(filePath))
//		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
//		if err != nil {
//			return err
//		}
//		fileInArchive, err := f.Open()
//		if err != nil {
//			return err
//		}
//		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
//			return err
//		}
//		_ = dstFile.Close()
//		_ = fileInArchive.Close()
//	}
//	_ = archive.Close()
//	return os.Remove(targetZip)
//}

//func getOS() string {
//	switch runtime.GOOS {
//	case "linux":
//		return "linux-x86_64"
//	case "windows":
//		return "win64"
//	case "darwin":
//		return "osx-x86_64"
//	default:
//		return "unknown"
//	}
//}
//
//func dirInfo(dirName string) (bool, bool) {
//	dir, err := os.ReadDir(dirName)
//	if err != nil {
//		if os.IsNotExist(err) {
//			return false, true
//		}
//		h.Die("failed to read dir: %s", err)
//	}
//	if len(dir) != 0 {
//		return true, false
//	}
//	return true, true
//}

//func extract(name string) error {
//	return extractDir("skeleton", name, skeletonFS)
//}

//type localFS interface {
//	fs.FS
//	fs.ReadDirFS
//}

//func extractDir(name string, target string, f localFS) error {
//	dir, err := f.ReadDir(name)
//	if err != nil {
//		return err
//	}
//	closers := make([]func() error, 0)
//	defer func() {
//		for _, closer := range closers {
//			_ = closer()
//		}
//	}()
//	for _, entry := range dir {
//		if entry.IsDir() {
//			currentDir := filepath.Join(name, entry.Name())
//			targetDir := filepath.Join(target, entry.Name())
//			mkdir(targetDir)
//			if err = extractDir(currentDir, targetDir, f); err != nil {
//				return err
//			}
//		} else {
//			currentFile := filepath.Join(name, entry.Name())
//			targetFile := filepath.Join(target, entry.Name())
//			memoryFile, err := f.Open(currentFile)
//			if err != nil {
//				return err
//			}
//			closers = append(closers, memoryFile.Close)
//			diskFile, err := os.Create(targetFile)
//			if err != nil {
//				return err
//			}
//			closers = append(closers, diskFile.Close)
//			if _, err = io.Copy(diskFile, memoryFile); err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}

//func mkdir(name string) {
//	if err := os.MkdirAll(name, dirPerm); err != nil {
//		h.Die("failed to create directory: %s", err)
//	}
//}

//func h.Die(msg string, v ...any) {
//	fmt.Printf("error: %s\n", fmt.Sprintf(msg, v...))
//	os.Exit(1)
//}
