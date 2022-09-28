package create

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/skamenetskiy/grpcapp/grpcapp/h"
)

func Run(args []string) {
	if len(args) == 0 {
		h.Die("application name not specified")
	}
	name := strings.TrimSpace(args[0])
	if name == "" {
		h.Die("invalid application name")
	}
	dirExists, dirEmpty := h.DirInfo(name)
	if !dirEmpty {
		h.Die("directory '%s' exists and is not empty", name)
	}
	if !dirExists {
		h.Mkdir(name)
	}
	goModData := fmt.Sprintf("module %s\n\ngo 1.19\n\nrequire (\n\t"+
		"github.com/skamenetskiy/grpcapp v0.0.3\n)", name)
	if err := os.WriteFile(filepath.Join(name, "go.mod"), []byte(goModData), filePerm); err != nil {
		h.Die("failed to write go.mod: %s", err)
	}
	gitIgnoreData := ".idea\n.vscode\n.proto\n"
	if err := os.WriteFile(filepath.Join(name, ".gitignore"), []byte(gitIgnoreData), filePerm); err != nil {
		h.Die("failed to write .gitignore: %s", err)
	}
	if err := extract(name); err != nil {
		h.Die("failed to extract application files: %s", err)
	}
}

const (
	filePerm = 0664
)

type localFS interface {
	fs.FS
	fs.ReadDirFS
}

//go:embed skeleton
var skeletonFS embed.FS

func extract(name string) error {
	return extractDir(name, "skeleton", name, skeletonFS)
}

func extractDir(appName, name, target string, f localFS) error {
	dir, err := f.ReadDir(name)
	if err != nil {
		return err
	}
	closers := make([]func() error, 0)
	defer func() {
		for _, closer := range closers {
			_ = closer()
		}
	}()
	data := &struct {
		Name string
	}{
		Name: appName,
	}
	for _, entry := range dir {
		if entry.IsDir() {
			currentDir := filepath.Join(name, entry.Name())
			targetDir := filepath.Join(target, entry.Name())
			h.Mkdir(targetDir)
			if err = extractDir(appName, currentDir, targetDir, f); err != nil {
				return err
			}
		} else {
			currentFile := filepath.Join(name, entry.Name())
			targetFile := filepath.Join(target, strings.TrimSuffix(entry.Name(), "template"))
			memoryFile, err := f.Open(currentFile)
			if err != nil {
				return err
			}
			closers = append(closers, memoryFile.Close)
			diskFile, err := os.Create(targetFile)
			if err != nil {
				return err
			}
			closers = append(closers, diskFile.Close)
			if strings.HasSuffix(currentFile, "template") {
				b, err := io.ReadAll(memoryFile)
				if err != nil {
					return err
				}
				tpl, err := template.New("file").Parse(string(b))
				if err != nil {
					return err
				}
				if err = tpl.Execute(diskFile, data); err != nil {
					return err
				}
			} else {
				if _, err = io.Copy(diskFile, memoryFile); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
