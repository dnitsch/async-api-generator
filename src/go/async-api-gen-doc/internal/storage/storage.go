package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
)

// StorageClient defines the IFace for the storage clients behaviour
//
// Usually this would be defined on the consumer but in this case,
// like with io.Reader and others it makes sense to define it on the producer
type StorageClient interface {
	Fetch(ctx context.Context, p *StorageFetchRequest) error
	Upload(ctx context.Context, p *StorageUploadRequest) error
}

var ErrClientUnknown = errors.New("unknown client")

func ClientFactory(typ StorageType, dest string) (StorageClient, error) {
	switch typ {
	case Local:
		return NewLocalFS(dest)
	case AzBlob:
		rc, err := NewBlobClient(dest)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize the Blob Storage Client: %v", err)
		}
		return NewRemoteAzBlob(rc), nil
	default:
		return nil, fmt.Errorf("client type not recognized\n%w", ErrClientUnknown)
	}
}

// StorageUploadRequest
type StorageUploadRequest struct {
	ContainerName string
	BlobKey       string
	Destination   string
	Reader        io.Reader // readerObj //
	Writer        io.Writer
}

// StorageFetchRequest
type StorageFetchRequest struct {
	Destination   string
	ContainerName string
	BlobKey       string
	EmitPath      string
	Reader        io.Reader
	Writer        io.Writer
}
