package storage_test

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/dnitsch/async-api-generator/internal/storage"
)

func Test_Write_to_lfs_successfully(t *testing.T) {
	want := []byte(`{"foo":"bar"}`)

	t.Run("when passing in a writer", func(t *testing.T) {

		f := &bytes.Buffer{}
		sc, err := storage.NewLocalFS("tmpFile")
		if err != nil {
			t.Fatal(err)
		}
		if err := sc.Upload(context.TODO(), &storage.StorageUploadRequest{Reader: bytes.NewReader(want), Writer: f}); err != nil {
			t.Fatal(err)
		}

		b, _ := io.ReadAll(f)
		if string(b) != string(want) {
			t.Fatal("incorrect data written to writer")
		}
	})

	t.Run("when creating the writer within the upload LFS", func(t *testing.T) {
		dir := filepath.Join(os.TempDir(), "storage-test-lfs")
		destPath := filepath.Join(dir, "out.json")
		defer os.RemoveAll(dir)
		sc, err := storage.NewLocalFS(destPath)
		if err != nil {
			t.Fatal(err)
		}

		if err := sc.Upload(context.TODO(), &storage.StorageUploadRequest{Reader: bytes.NewReader(want), Destination: destPath}); err != nil {
			t.Fatal(err)
		}
		b, _ := os.ReadFile(destPath)
		if string(b) != string(want) {
			t.Fatal("incorrect data written to writer")
		}
	})
}

func Test_Write_to_lfs_fails(t *testing.T) {
	t.Skip()
	tmpFile, _ := os.CreateTemp(os.TempDir(), "storage-test")
	sc, err := storage.NewLocalFS(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	os.Remove(tmpFile.Name())
	if err := sc.Upload(context.TODO(), &storage.StorageUploadRequest{Reader: bytes.NewReader([]byte(`{"foo":"bar"}`))}); err == nil {
		t.Fatal("expected err to not be nil")
	}
}

func Test_Fetch_LocalFS_succeeds(t *testing.T) {
	t.Run("when dir exists with multiple items", func(t *testing.T) {
		sourceDir := filepath.Join(os.TempDir(), "storage-test-lfs")
		emitDir := filepath.Join(os.TempDir(), "emit-out-test-lfs")
		os.MkdirAll(emitDir, 0o777)
		defer os.RemoveAll(emitDir)

		dirCh1 := filepath.Join(sourceDir, "child1")
		dirCh2 := filepath.Join(sourceDir, "child2")

		defer os.RemoveAll(sourceDir)

		if err := os.MkdirAll(dirCh1, 0o777); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(dirCh2, 0o777); err != nil {
			t.Fatal(err)
		}
		f1, _ := os.CreateTemp(dirCh1, "some1")
		f1.Write([]byte(`{"json":true}`))
		f2, _ := os.CreateTemp(dirCh2, "some2")
		f2.Write([]byte(`{"json":true}`))
		// sc.
		sc, err := storage.NewLocalFS("tmpFile")
		if err != nil {
			t.Fatal(err)
		}
		fetchReq := &storage.StorageFetchRequest{Destination: sourceDir, ContainerName: "", EmitPath: emitDir}
		if err := sc.Fetch(context.TODO(), fetchReq); err != nil {
			t.Fatal(err)
		}
		found := []string{}
		filepath.WalkDir(emitDir, func(path string, d fs.DirEntry, err error) error {
			found = append(found, path)
			return nil
		})
		if len(found) < 5 {
			t.Fatalf("no items transferred over, got: %d\n", len(found))
		}
	})

}
