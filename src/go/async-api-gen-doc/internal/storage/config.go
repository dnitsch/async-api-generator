package storage

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
)

type StorageType struct {
	typ string
}

func (s StorageType) Typ() StorageType {
	switch s.typ {
	case "local":
		return Local
	case "azblob":
		return AzBlob
	default:
		return Uknown
	}
}

// Another way of making an ENUM type with compile time checks
var (
	Uknown = StorageType{""}
	Local  = StorageType{"local"}
	AzBlob = StorageType{"azblob"}
)

type Conf struct {
	Typ            StorageType
	Destination    string // Account (Blob) or Bucket(S3) or Folder(FS)
	TopLevelFolder string // container in AzBlob or part of the path in GCP/AWS/Alicloud/etc...
}

var (
	ErrStorageProtocol            = errors.New("protocol error, incorrect format of protocol marker - should be in `://` form")
	ErrUnsupportedStorageProtocol = errors.New("unsupported protocol error, must be one of ['local://','azblob://']")
	ErrStorageSegment             = errors.New("segment error, must include at least 1 segment separation")
	ErrStorageOutputZeroLength    = errors.New("output zero length error")
)

// ParseStorageOutputConfig should always be set either as the default or user supplied
// will default to local://$HOME/.gendoc
func ParseStorageOutputConfig(out string) (*Conf, error) {
	if len(out) == 0 {
		return nil, fmt.Errorf("output: '%s'\n%w", out, ErrStorageOutputZeroLength)
	}
	s := strings.Split(out, "://")

	if len(s) != 2 {
		return nil, fmt.Errorf("output string('%s')\n%w", out, ErrStorageProtocol)
	}
	typ := StorageType{s[0]}.Typ()
	if typ == Uknown {
		return nil, fmt.Errorf("protocol: %s\n%w", s[0], ErrUnsupportedStorageProtocol)
	}

	conf := &Conf{
		Typ: typ,
	}

	if typ == Local {
		if len(s[1]) < 1 {
			return nil, fmt.Errorf("at least one file system segment must be provided: ('%s'\n%w", s[1], ErrStorageSegment)
		}
		conf.Destination = filepath.Join(s[1:]...)
		conf.TopLevelFolder = ""
		if s[1] == "$HOME/.gendoc" {
			// convert to default
			home, err := homedir.Dir()
			if err != nil {
				return nil, err
			}
			conf.Destination = home
			conf.TopLevelFolder = ".gendoc"
		}

		return conf, nil
	}

	restS := strings.Split(strings.TrimPrefix(s[1], "/"), "/")
	if len(restS) < 2 {
		return nil, fmt.Errorf("specified path ('%s')\n%w", strings.TrimPrefix(s[1], "/"), ErrStorageSegment)
	}
	// in case a triple slash was used as in posix compliant system file protocol
	// this is either the bucket/blob container/fspath parent
	conf.Destination = restS[0]
	conf.TopLevelFolder = restS[1]
	// conf.TopLevelFolder = strings.Join(restS[1:], "/")
	return conf, nil
}
