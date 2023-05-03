package cmdutil_test

import (
	"os"
	"strings"
	"testing"

	"github.com/dnitsch/async-api-generator/internal/cmdutil"
)

func Test_ListFiles_in_mock_directort(t *testing.T) {
	baseDir := "./test/samples"
	// debug test purposes only
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(cwd, "internal/cmdutil") {
		baseDir = "../../test/samples"
	}
	// END DEBUG
	got, err := cmdutil.ListFiles(baseDir)
	if err != nil {
		t.Fatalf("%v", err)
	}
	ttests := []struct {
		expectName string
		expectPath string
		expectType string
	}{
		{"sample.cs", "../../test/samples/business/sample.cs", "cs"},
		{"index.md", "../../test/samples/index.md", "md"},
		{"sample.tf", "../../test/samples/infra/sample.tf", "tf"},
	}
	for idx, tt := range ttests {
		gotLf := got[idx]
		if gotLf.Name != tt.expectName {
			t.Errorf("expected names of files to be equal. got: %s, wanted: %s", gotLf.Name, tt.expectName)
		}
		if gotLf.Path != tt.expectPath {
			t.Errorf("expected path of files to be equal. got: %s, wanted: %s", gotLf.Path, tt.expectPath)
		}
		if gotLf.Type != tt.expectType {
			t.Errorf("expected types of files to be equal. got: %s, wanted: %s", gotLf.Type, tt.expectType)
		}
	}
}
