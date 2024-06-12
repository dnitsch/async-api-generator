package fshelper_test

import (
	"bytes"
	"io"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/dnitsch/async-api-generator/internal/fshelper"
)

func Test_ListFiles_in_mock_directory(t *testing.T) {
	baseDir := "test/foo.sample"

	got, err := fshelper.ListFiles(fshelper.DebugDirHelper(t, baseDir, "internal/fshelper", "../../"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	ttests := []struct {
		expectName string
		expectPath string
		expectType string
	}{
		{"someeventpoco.cs", "test/foo.sample/src/someeventpoco.cs", "cs"},
		{"index.md", "test/foo.sample/index.md", "md"},
		{"sample.tf", "test/foo.sample/infra/sample.tf", "tf"},
	}
	var found *fshelper.FileList = nil
	for _, tt := range ttests {
		for _, gotLF := range got {
			if tt.expectName == gotLF.Name {
				found = gotLF
				break
			}
		}
		if found == nil {
			t.Error("not found test files")
			break
		}
		if found.Name != tt.expectName {
			t.Errorf("expected names of files to be equal. got: %s, wanted: %s", found.Name, tt.expectName)
		}
		if !strings.Contains(found.Path, tt.expectPath) {
			t.Errorf("expected path of files to be equal. got: %s, wanted: %s", found.Path, tt.expectPath)
		}
		if found.Type != tt.expectType {
			t.Errorf("expected types of files to be equal. got: %s, wanted: %s", found.Type, tt.expectType)
		}
	}
}

func Test_Flush_success(t *testing.T) {

	ttests := map[string]struct {
		byteWriter *bytes.Buffer
		want       string
		input      any
	}{
		"with annotated property": {
			byteWriter: &bytes.Buffer{},
			want:       `{"foo":"bar"}`,
			input: interface{}(struct {
				Foo string `json:"foo"`
			}{Foo: "bar"}),
		},
		"withOUT annotated property": {
			byteWriter: &bytes.Buffer{},
			want:       `{"Foo":"bar"}`,
			input: interface{}(struct {
				Foo string
			}{Foo: "bar"}),
		},
		"should pass with binary num": {
			byteWriter: &bytes.Buffer{},
			want:       "0",
			input:      interface{}(int(0b0)),
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			fsh := &fshelper.FSHelper{}
			err := fsh.FlushJson(tt.byteWriter, tt.input)

			if err != nil {
				t.Fatal(err)
			}
			got := tt.byteWriter.String()
			if got != tt.want {
				t.Fatalf("got: %v, wanted: %v", got, tt.want)
			}
		})
	}
}

func Test_Flush_error(t *testing.T) {

	ttests := map[string]struct {
		byteWriter func() io.Writer
		input      any
	}{
		"with unsuportedtype error": {
			byteWriter: func() io.Writer { return &bytes.Buffer{} },
			input:      make(chan int),
		},
		"with UnsupportedValueError": {
			byteWriter: func() io.Writer { return &bytes.Buffer{} },
			input:      math.Inf(1),
		},
		"with writer closed": {
			byteWriter: func() io.Writer {
				w, _ := os.CreateTemp(os.TempDir(), "foo")
				defer w.Close()
				return w
			},
			input: `string`,
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			fsh := &fshelper.FSHelper{}
			err := fsh.FlushJson(tt.byteWriter(), tt.input)
			if err == nil {
				t.Fatalf("got: <nil>, wanted: %v", err)
			}
		})
	}
}
