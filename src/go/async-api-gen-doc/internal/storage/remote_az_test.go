package storage_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/dnitsch/async-api-generator/internal/storage"
)

func Test_Remote_Client_AzBlob(t *testing.T) {
	t.Run("new client", func(t *testing.T) {
		_, err := storage.NewBlobClient("stdevsandboxeuwdev")
		if err != nil {
			t.Fatal(err)
		}
	})
}

type mockAzClient struct {
	download    func(ctx context.Context, containerName string, blobName string, o *azblob.DownloadStreamOptions) (azblob.DownloadStreamResponse, error)
	upload      func(ctx context.Context, containerName string, blobName string, body io.Reader, o *azblob.UploadStreamOptions) (azblob.UploadStreamResponse, error)
	listSegment func(containerName string, o *azblob.ListBlobsFlatOptions) *runtime.Pager[azblob.ListBlobsFlatResponse]
}

func (m mockAzClient) DownloadStream(ctx context.Context, containerName string, blobName string, o *azblob.DownloadStreamOptions) (azblob.DownloadStreamResponse, error) {
	return m.download(ctx, containerName, blobName, o)
}
func (m mockAzClient) UploadStream(ctx context.Context, containerName string, blobName string, body io.Reader, o *azblob.UploadStreamOptions) (azblob.UploadStreamResponse, error) {
	return m.upload(ctx, containerName, blobName, body, o)
}

func (m mockAzClient) NewListBlobsFlatPager(containerName string, o *azblob.ListBlobsFlatOptions) *runtime.Pager[azblob.ListBlobsFlatResponse] {
	return m.listSegment(containerName, o)
}

func Test_Write_to_remote_az(t *testing.T) {
	t.Run("succeeds with correct input", func(t *testing.T) {
		mc := &mockAzClient{
			upload: func(ctx context.Context, containerName, blobName string, body io.Reader, o *azblob.UploadStreamOptions) (azblob.UploadStreamResponse, error) {
				if blobName != "foo.json" {
					t.Fatalf("incorrect blob key passed in")
				}
				return azblob.UploadStreamResponse{}, nil
			},
		}
		sc := storage.NewRemoteAzBlob(mc)
		err := sc.Upload(context.TODO(), &storage.StorageUploadRequest{ContainerName: "bar", BlobKey: "foo.json"})
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("fails with remote error", func(t *testing.T) {
		mc := &mockAzClient{
			upload: func(ctx context.Context, containerName, blobName string, body io.Reader, o *azblob.UploadStreamOptions) (azblob.UploadStreamResponse, error) {
				return azblob.UploadStreamResponse{}, fmt.Errorf("unable to write to blob path('%s)", blobName)
			},
		}
		sc := storage.NewRemoteAzBlob(mc)
		err := sc.Upload(context.TODO(), &storage.StorageUploadRequest{ContainerName: "bar", BlobKey: "foo.json"})
		if err == nil {
			t.Fatal(err)
		}
	})
}

func Test_fetch_from_remote_az(t *testing.T) {

	t.Run("succeeds with correct input", func(t *testing.T) {
		mc := &mockAzClient{
			download: func(ctx context.Context, containerName, blobName string, o *azblob.DownloadStreamOptions) (azblob.DownloadStreamResponse, error) {
				return azblob.DownloadStreamResponse{DownloadResponse: blob.DownloadResponse{
					Body: io.NopCloser(strings.NewReader("returned from azblob")),
				}}, nil
			},
			listSegment: func(containerName string, o *azblob.ListBlobsFlatOptions) *runtime.Pager[azblob.ListBlobsFlatResponse] {
				if containerName != "bar" {
					t.Fatal("incorrect containername")
				}
				if !o.Include.Deleted {
					t.Fatal("incorrect option passed in INclude Deleted blobs, got <false> wanted <true>")
				}
				count := 0
				respList := azblob.ListBlobsFlatResponse{}
				handler := runtime.PagingHandler[azblob.ListBlobsFlatResponse]{}
				handler.More = func(cclbfsr azblob.ListBlobsFlatResponse) bool {
					count++
					return false
				}
				err := runtime.UnmarshalAsJSON(&http.Response{
					Body: io.NopCloser(bytes.NewReader([]byte(`{"Segment": {"Blobs": []}}`))),
				}, &respList.ListBlobsFlatSegmentResponse)
				if err != nil {
					t.Fatal(err)
				}
				blobName := "foo.json"
				handler.Fetcher = func(ctx context.Context, cclbfsr *azblob.ListBlobsFlatResponse) (azblob.ListBlobsFlatResponse, error) {
					respList.ListBlobsFlatSegmentResponse.Segment.BlobItems = append(respList.ListBlobsFlatSegmentResponse.Segment.BlobItems, &container.BlobItem{Name: &blobName})
					return respList, nil
				}
				return runtime.NewPager(handler)
			},
		}
		w := &bytes.Buffer{}
		sc := storage.NewRemoteAzBlob(mc)
		upReq := &storage.StorageFetchRequest{ContainerName: "bar", BlobKey: "bax.json", Destination: "", Writer: w}
		if err := sc.Fetch(context.TODO(), upReq); err != nil {
			t.Fatal(err)
		}
		writtenBytes, _ := io.ReadAll(w)

		if len(writtenBytes) < 1 {
			t.Fatal("nothing written to writer")
		}
		if string(writtenBytes) != "returned from azblob" {
			t.Fatal("incorrect data written to writer")
		}
	})
}

func Test_az_blob_client_should_not_error_on_create(t *testing.T) {
	_, err := storage.NewBlobClient("account_name")
	if err != nil {
		t.Fatal(err)
	}
}
