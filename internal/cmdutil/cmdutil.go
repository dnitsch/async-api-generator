package cmdutil

import (
	"io/fs"
	"path/filepath"
	"strings"
)

type FileList struct {
	Name string // fullname of file
	Path string // full path to file - either relative or full
	Type string // type of file e.g. schema json, CS, TF, K8sYaml, HelmYaml
}

func ListFiles(baseDir string) ([]*FileList, error) {
	// init a
	files := []*FileList{}
	if err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			files = append(files, &FileList{Name: d.Name(), Path: path, Type: strings.TrimPrefix(filepath.Ext(path), ".")})
			return nil
		}
		return nil
	}); err != nil {
		// return nil - even if some files were processed
		return nil, err
	}
	return files, nil
}
