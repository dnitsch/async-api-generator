package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"
)

// LocalFS implements StorageClient iface for use with FS
type LocalFS struct {
}

func NewLocalFS(dest string) (*LocalFS, error) {
	return &LocalFS{}, nil
}

// Upload writes to the provided os.File writer
// Checks and creates the dir if necessary
func (ls *LocalFS) Upload(ctx context.Context, p *StorageUploadRequest) error {
	// Ensure dir exists
	if err := os.MkdirAll(filepath.Dir(p.Destination), 0o766); err != nil {
		return err
	}
	// set writer to Destination if empty
	if p.Writer == nil {
		f, err := os.Create(p.Destination)
		if err != nil {
			return err
		}
		p.Writer = f
	}

	b, err := io.ReadAll(p.Reader)
	if err != nil {
		return err
	}

	_, err = p.Writer.Write(b)
	return err
}

// Fetch in LocalFS takes source path and copies into the Interim EmitPath
// EmitPath in most cases will be the interim `DownloadDir`.
func (ls *LocalFS) Fetch(ctx context.Context, p *StorageFetchRequest) error {
	return cp.Copy(filepath.Join(p.Destination, p.ContainerName), p.EmitPath)
}
