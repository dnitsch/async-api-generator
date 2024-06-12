package generate_test

import (
	"bytes"
	"testing"

	"github.com/dnitsch/async-api-generator/internal/fshelper"
	"github.com/dnitsch/async-api-generator/internal/generate"
	"github.com/dnitsch/async-api-generator/internal/parser"
	log "github.com/dnitsch/simplelog"
)

var baseDir = "test/foo.sample"

func Test_SAMPLE_Generate_ProcessedInputs_from_directory_input(t *testing.T) {

	inputs, _ := fshelper.ListFiles(fshelper.DebugDirHelper(t, baseDir, "internal/generate", "../../"))
	g := generate.New(&generate.Config{ParserConfig: parser.Config{ServiceId: "bazquxsample", ServiceRepoUrl: "https://github.com/asynapi-gen"}}, log.New(&bytes.Buffer{}, log.ErrorLvl))
	g.LoadInputsFromFiles(inputs)

	err := g.GenDocBlox()
	if err != nil {
		t.Fatal(err)
	}
	if len(*g.Processed()) <= 5 {
		t.Fatalf("got length: %d, wanted number of statements to be >= 5", len(*g.Processed()))
	}
}

func Test_SAMPLE_Error_in_single_file_should_fail_all(t *testing.T) {
	got, _ := fshelper.ListFiles(fshelper.DebugDirHelper(t, baseDir, "internal/generate", "../../"))

	g := generate.New(&generate.Config{}, log.New(&bytes.Buffer{}, log.ErrorLvl))
	g.LoadInputsFromFiles(got)

	err := g.GenDocBlox()
	if err == nil {
		t.Fatal(err)
	}
	if g.Processed() != nil {
		t.Fatalf("got: %v, wanted: <nil>", g.Processed())
	}
}
