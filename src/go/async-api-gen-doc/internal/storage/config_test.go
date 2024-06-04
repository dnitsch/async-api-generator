package storage_test

import (
	"errors"
	"testing"

	"github.com/dnitsch/async-api-generator/internal/storage"
	"github.com/mitchellh/go-homedir"
)

func Test_ParseStorageOutputConfig_succeeds_with(t *testing.T) {
	ttests := map[string]struct {
		input  string
		expect func() (storage.StorageType, string, string)
	}{
		"local default config parsed OK": {
			input: "local://$HOME/.gendoc",
			expect: func() (storage.StorageType, string, string) {
				home, _ := homedir.Dir()
				return storage.Local, home, ".gendoc"
			},
		},
		"local user supplied config parsed OK": {
			input: "local:///some/path/.gendoc",
			expect: func() (storage.StorageType, string, string) {
				return storage.Local, "/some/path/.gendoc", ""
			},
		},
		"local user relative path": {
			input: "local://.some/path/another",
			expect: func() (storage.StorageType, string, string) {
				return storage.Local, ".some/path/another", ""
			},
		},
		"az user supplied config parsed OK": {
			input: "azblob://account/container",
			expect: func() (storage.StorageType, string, string) {
				return storage.AzBlob, "account", "container"
			},
		},
		"az user supplied config with additional paths parsed OK": {
			input: "azblob://account/container/some/ignore/path",
			expect: func() (storage.StorageType, string, string) {
				return storage.AzBlob, "account", "container"
			},
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			conf, err := storage.ParseStorageOutputConfig(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			expTyp, expDest, expTLF := tt.expect()
			checkStorageOutConfig(t, *conf, expTyp, expDest, expTLF)
		})
	}
}

func checkStorageOutConfig(t *testing.T, conf storage.Conf, wantTyp storage.StorageType, wantDest, wantTopLF string) {
	if conf.Typ != wantTyp {
		t.Errorf("storage type wrong, got: %v\nwanted: %v", conf.Typ, storage.Local.Typ())
	}
	if conf.Destination != wantDest {
		t.Errorf("Destination wrong, got: %v\nwanted: %v", conf.Destination, wantDest)
	}
	if conf.TopLevelFolder != wantTopLF {
		t.Errorf("Destination wrong, got: %v\nwanted: %v", conf.TopLevelFolder, wantTopLF)
	}
}

func Test_ParseStorageOutputConfig_fails_with(t *testing.T) {
	ttests := map[string]struct {
		input     string
		expectErr error
	}{
		"empty string user supplied": {
			input:     "",
			expectErr: storage.ErrStorageOutputZeroLength,
		},
		"incorrect string user supplied local proto": {
			input:     "local:/",
			expectErr: storage.ErrStorageProtocol,
		},
		"incorrect string user supplied azblob proto": {
			input:     "azblob:/",
			expectErr: storage.ErrStorageProtocol,
		},
		"incorrect string user supplied local path": {
			input:     "local://",
			expectErr: storage.ErrStorageSegment,
		},
		"incorrect string user supplied azblob destination": {
			input:     "azblob://blob_account_name",
			expectErr: storage.ErrStorageSegment,
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			_, err := storage.ParseStorageOutputConfig(tt.input)
			if err == nil {
				t.Fatal("got <nil>, wanted: not nil")
			}
			if !errors.Is(err, tt.expectErr) {
				t.Fatalf("incorrect error returned, got: %v, wanted: %v", err, tt.expectErr)
			}
		})
	}
}
