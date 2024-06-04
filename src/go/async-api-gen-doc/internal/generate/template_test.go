package generate_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dnitsch/async-api-generator/internal/fshelper"
	"github.com/dnitsch/async-api-generator/internal/generate"
	"github.com/dnitsch/async-api-generator/internal/parser"
	"github.com/dnitsch/async-api-generator/internal/storage"
	log "github.com/dnitsch/simplelog"
	"gopkg.in/yaml.v3"
)

var setupAsyncApiRoot = generate.AsyncAPIRoot{
	AsyncAPI:           "2.5.0",
	ID:                 "urn:business_domain:bounded_context:service_name",
	DefaultContentType: "application/json",
	Info: generate.Info{
		Title:   "some title for service",
		Version: "1.0.0",
		Description: `some long description here with unicode chars
* random 🌃
* cool  😎
* charts 📈
`,
	},
	Servers: map[string]generate.Server{
		"dev": {
			URL:         "foo.bar.dev",
			Description: "Dev servicebus",
			Protocol:    "amqp",
		},
		"preprod": {
			URL:         "foo.bar.preprod",
			Description: "pre servicebus",
			Protocol:    "amqp",
		},
	},
	Channels: map[string]generate.Channel{
		"test_topic_publish_channel": {
			Description: "test description of topic",
			Parameters:  map[string]generate.Parameter{},
			Publish: &generate.Operation{
				Summary:     "channel sumary",
				OperationId: "someid",
				Traits:      []interface{}{},
				Message: &generate.Message{
					Name:    "message",
					Summary: "m summary",
					Headers: []map[string]generate.Schema{},
					Payload: `{"properties":{
												"foo":{
													"default": "bar",
													"type":"string"
											}
											},
											"type": "object"
						}`,
					Title:        "mtitle",
					Description:  "m desc",
					MessageId:    "msg_id",
					Tags:         []generate.Tag{{Name: "version", Description: "0.0.1"}},
					ExternalDocs: generate.ExternalDocumentation{},
				},
				Bindings: []interface{}{},
			},
		},
	},
	Components: &generate.Components{},
	Tags:       []generate.Tag{},
}

func Test_Generate_From_asyncapi_root(t *testing.T) {
	w := &bytes.Buffer{}
	dc, err := generate.NewTemplateProcessor()

	if err != nil {
		t.Fatal(err)
	}

	dc.GenerateFromRoot(w, setupAsyncApiRoot)

	got := w.String()

	if len(got) < 1 {
		t.Fatalf("got %d, wanted a non empty buffer", len(got))
	}

	unmarhslldGot := &generate.AsyncAPIRoot{}
	if err := yaml.Unmarshal([]byte(got), unmarhslldGot); err != nil {
		t.Fatalf(`input:
--------
%s
--------
failed: %v`, got, err)
	}

	ttests := map[string]struct {
		expect func() string
		got    func() string
	}{
		"descriptions match": {
			func() string { return setupAsyncApiRoot.Info.Description },
			func() string { return unmarhslldGot.Info.Description },
		},
		"payloads match": {
			func() string {
				return strings.Join(strings.Fields(fmt.Sprintf("%s", setupAsyncApiRoot.Channels["test_topic_publish_channel"].Publish.Message.Payload)), "")
			},
			func() string {
				// expecting a valid schema to have been unmarshalled
				b, err := json.Marshal(unmarhslldGot.Channels["test_topic_publish_channel"].Publish.Message.Payload)
				if err != nil {
					t.Fatalf("failed to parse schema back to string: %v", err)
				}
				return strings.TrimSpace(string(b))
			},
		},
	}
	for _, tt := range ttests {
		if !strings.EqualFold(tt.got(), tt.expect()) {
			t.Errorf("expected the outputs to be equal, got: %s, wanted: %s", tt.got(), tt.expect())
		}
	}
}

func Test_BuildAsyncAPIRoot_from_tree(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "throw-away")
	err := os.MkdirAll(testDir, 0o777)
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(testDir)

	baseDir := "test/interim-generated"

	input, _ := fshelper.ListFiles(fshelper.DebugDirHelper(t, baseDir, "internal/generate", "../../"))

	globalConf := &generate.Config{ParserConfig: parser.Config{},
		InterimOutputDir: testDir,
		SearchDirName:    fshelper.DebugDirHelper(t, baseDir, "internal/generate", "../../"),
		Output: &storage.Conf{
			Destination:    testDir,
			Typ:            storage.Local,
			TopLevelFolder: "",
		},
	}

	g := generate.New(globalConf, log.New(os.Stdout, log.DebugLvl))
	g.LoadInputsFromFiles(input)

	if err := g.ConvertProcessed(); err != nil {
		t.Fatal(err)
	}

	if err := g.BuildContextTree(); err != nil {
		t.Fatal(err)
	}

	if err := g.AsyncAPIFromProcessedTree(); err != nil {
		t.Fatal(err)
	}

	got, _ := fshelper.ListFiles(testDir)

	if len(got) != 2 {
		t.Error("wanted 2 services written out")
	}

	// for _, file := range got {
	// 	b, _ := os.ReadFile(file.Path)
	// 	// fmt.Printf("file: %s\n\n%s\n\n", file.Path, string(b))
	// }
}
