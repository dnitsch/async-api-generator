package fshelper

import (
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type FSHelper struct {
	// interim      string
	// processed    string
	// remoteConfig remote.AzBlobRequest
	// remoteType   StorageType // remoteAddr string
	// rc           remote.RemoteClient
	// ctx          context.Context
}

// // New returns a new "instance" of the FSHelper
// func New(ctx context.Context, interim, processed string, client remote.RemoteClient, rmtConf remote.AzBlobRequest) *FSHelper {
// 	return &FSHelper{
// 		ctx:          ctx,
// 		interim:      interim,
// 		processed:    processed,
// 		rc:           client,
// 		remoteConfig: rmtConf,
// 		remoteType:   AzBlob, // hardcode AzBlob for now
// 	}
// }

type FileList struct {
	Name string // fullname of file
	Path string // full path to file - either relative or full
	Type string // type of file e.g. schema json, CS, TF, K8sYaml, HelmYaml
	// Mu   sync.Mutex // TODO: make mu private
}

var (
	skipFile = map[string]bool{".DS_Store": true, ".dockerignore": true, ".gitignore": true}
	skipDir  = map[string]bool{"bin": true, "dist": true, "node_modules": true, ".cache": true, ".terraform": true, ".git": true, "obj": true}
)

// ListFiles prepares a list of all files in lexical order we care about.
// Skipping any binary files like .DS_Store, compiled outputs, like
func ListFiles(baseDir string) ([]*FileList, error) {
	// init a
	files := []*FileList{}
	if err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// skip entire directory
		if d.IsDir() && skipDir[d.Name()] {
			return filepath.SkipDir
		}

		if !d.IsDir() {
			if skipFile[d.Name()] {
				return nil
			}
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

// TODO: is this is actually required?

// FlushJson writes a file to disk using JSON struct tags to convert
func (fs *FSHelper) FlushJson(w io.Writer, v any) error {
	// helper function
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if _, err := w.Write(b); err != nil {
		return err
	}
	return nil
}

// DebugDirHelper is passed a baseDir as if it's run from a makefile
// e.g. `./test/samples`
func DebugDirHelper(t *testing.T, targetDir, pkgDir, segmentLocation string) string {
	// debug test purposes only
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(cwd, pkgDir) {
		targetDir = filepath.Join(cwd, segmentLocation, targetDir)
	}
	return targetDir
}
