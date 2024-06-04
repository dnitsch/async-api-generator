package storage_test

import (
	"testing"

	"github.com/dnitsch/async-api-generator/internal/storage"
)

func Test_Factory_Correctly_returns(t *testing.T) {
	t.Run("localFs concrete impl ", func(t *testing.T) {
		client, err := storage.ClientFactory(storage.Local, "__")
		if err != nil {
			t.Fatal(err)
		}
		impl, ok := client.(*storage.LocalFS)
		if !ok {
			t.Fatalf("wrong type returned, got: %v, wanted: %v", impl, "storage.LocalFS")
		}

	})
	t.Run("RemoteAzBlob concrete impl ", func(t *testing.T) {
		client, err := storage.ClientFactory(storage.AzBlob, "__")
		if err != nil {
			t.Fatal(err)
		}
		impl, ok := client.(*storage.RemoteAzBlob)
		if !ok {
			t.Fatalf("wrong type returned, got: %v, wanted: %v", impl, "storage.RemoteAzBlob")
		}
	})

	t.Run("Uknown should return an error", func(t *testing.T) {
		_, err := storage.ClientFactory(storage.Uknown, "__")
		if err == nil {
			t.Fatal(err)
		}
	})
}
