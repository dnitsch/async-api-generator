package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"golang.org/x/sync/errgroup"
)

// RemoteAzBlob implements the StorageClient interface for use with AzureBlob Storage
type RemoteAzBlob struct {
	client BlobApi
}

type BlobListSegmentPager interface {
	More() bool
	NextPage(ctx context.Context) (*azblob.ListBlobsFlatResponse, error)
}

// BlobApi defines the Az Blob methods we care about
type BlobApi interface {
	NewListBlobsFlatPager(containerName string, o *azblob.ListBlobsFlatOptions) *runtime.Pager[azblob.ListBlobsFlatResponse]
	DownloadStream(ctx context.Context, containerName string, blobName string, o *azblob.DownloadStreamOptions) (azblob.DownloadStreamResponse, error)
	UploadStream(ctx context.Context, containerName string, blobName string, body io.Reader, o *azblob.UploadStreamOptions) (azblob.UploadStreamResponse, error)
}

// NewRemoteAzBlob returns an instance of StorageClient with AZ concrete impl
func NewRemoteAzBlob(client BlobApi) *RemoteAzBlob {
	return &RemoteAzBlob{
		client: client,
	}
}

// NewBlobClient used inside the client factory
func NewBlobClient(account string) (*azblob.Client, error) {
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", account)
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	return azblob.NewClient(serviceURL, cred, nil)
}

// Fetch downloads and stores stream from remote AZBlob
func (fs *RemoteAzBlob) Fetch(ctx context.Context, p *StorageFetchRequest) error {
	ctx_, cancel := context.WithCancel(ctx)
	defer cancel()

	pager := fs.client.NewListBlobsFlatPager(p.ContainerName, &azblob.ListBlobsFlatOptions{
		Include: container.ListBlobsInclude{Deleted: true, Versions: true},
	})

	blobs := []string{}
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return err // if err is not nil, break the loop.
		}

		for _, _blob := range resp.Segment.BlobItems {
			blobs = append(blobs, *_blob.Name)
		}
	}

	// Download from AZ concurrently
	g := new(errgroup.Group)

	for _, blob := range blobs {
		blob := blob // initializing a per iteration value for `blob`
		g.Go(func() error {
			return fs.fetchSingleAzBlob(ctx_, blob, *p)
		})
	}

	return g.Wait()
}

func (fs *RemoteAzBlob) fetchSingleAzBlob(ctx context.Context, blobName string, fr StorageFetchRequest) error {

	get, err := fs.client.DownloadStream(ctx, fr.ContainerName, blobName, &azblob.DownloadStreamOptions{})
	if err != nil {
		return err
	}

	downloadedData := bytes.Buffer{}

	retryReader := get.NewRetryReader(ctx, &azblob.RetryReaderOptions{})

	if _, err := downloadedData.ReadFrom(retryReader); err != nil {
		return err
	}

	if err := retryReader.Close(); err != nil {
		return err
	}

	lfs := &LocalFS{}

	upReq := &StorageUploadRequest{Destination: filepath.Join(fr.EmitPath, filepath.Base(blobName)), Reader: &downloadedData}
	upReq.Writer = fr.Writer
	// store in the interim directory
	return lfs.Upload(ctx, upReq)
}

// Upload
func (fs *RemoteAzBlob) Upload(ctx context.Context, p *StorageUploadRequest) error {
	ctx_, cancel := context.WithCancel(ctx)
	defer cancel()
	// p.BlobKey in this case is base path where the uploads will be placed
	if _, err := fs.client.UploadStream(ctx_, p.ContainerName, p.BlobKey, p.Reader, &azblob.UploadStreamOptions{
		BlockSize:   int64(1024), // 1Mib
		Concurrency: 1,           // most files should only ever be less than 1Mib so no need to concurrent chunking
	}); err != nil {
		return err
	}
	return nil
}
