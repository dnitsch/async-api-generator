package parser_test

import (
	"errors"
	"os"
	"testing"

	"github.com/dnitsch/async-api-generator/internal/gendoc"
	"github.com/dnitsch/async-api-generator/internal/lexer"
	"github.com/dnitsch/async-api-generator/internal/parser"
	log "github.com/dnitsch/simplelog"
)

var lexerSource = lexer.Source{FileName: "bar", FullPath: "/foo/bar"}

func Test_Initial_GenDocBlocks(t *testing.T) {
	ttests := map[string]struct {
		input              string
		expectedIdentifier string
		expectedValue      any
	}{
		"gendoc found and contains data": {`let x = 42;
//+gendoc category=message type=nameId id=MessageId parent=somechannel
//-gendoc
some other throw away
//+gendoc category=message type=nameId id=someid channelId=somechannel
stuff {
	here string
}
//-gendoc
`, "category=message type=nameId", `stuff {
	here string
}`},
		"should not throw and error when content empty": {`let x = 42;
//+gendoc category=message type=nameId id=someid channelId=somechannel
ignore me
//-gendoc
//+gendoc category=message type=nameId id=someid channelId=somechannel
//-gendoc
`, "category=message type=nameId id=someid", ""},
	}

	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			lexerSource.Input = tt.input
			l := lexer.New(lexerSource)
			p := parser.New(l, &parser.Config{}).WithLogger(log.New(os.Stderr, log.ErrorLvl))
			parsed, errs := p.InitialParse()
			if len(errs) > 0 {
				t.Fatalf("parser had errors, expected <nil>\nerror: %v", errs)
			}

			if len(parsed) != 2 {
				t.Fatalf("program.Statements does not contain 2 statements. got=%d",
					len(parsed))
			}

			stmt := parsed[1]
			if !testHelperGenDocBlock(t, stmt, tt.expectedIdentifier, tt.expectedValue) {
				return
			}
		})
	}
}

func Test_Parse_Id_and_type_of_nodes(t *testing.T) {
	ttests := map[string]struct {
		input        string
		config       *parser.Config
		wantCatType  parser.NodeCategory
		wantUrn      string
		wantId       string
		wantParentId string
	}{
		"service succeed with serviceId inherited in single repo mode": {`let x = 42;
			//+gendoc category=info type=description
			this is some description
			//-gendoc`,
			&parser.Config{ServiceId: "foo"},
			parser.ServiceNode, "urn:::foo", "foo", "",
		},
		"service succeed with serviceId inherited in markdown": {`let x = 42;
		<!-- //+gendoc category=info type=description -->
		this is some description
		<!-- //-gendoc -->`,
			&parser.Config{ServiceId: "foo"},
			parser.ServiceNode, "urn:::foo", "foo", "",
		},
		"service succeed with serviceId specified": {`let x = 42;
			//+gendoc category=server type=description id=baz
			this is some description
			//-gendoc`,
			&parser.Config{ServiceId: "foo"},
			parser.ServiceNode, "urn:::baz", "baz", "",
		},
		"service succeed with serviceId extracted": {`let x = 42;
			//+gendoc category=root type=nameId
			bazquxservice.sample
			//-gendoc`,
			&parser.Config{},
			parser.ServiceNode, "urn:::bazquxservice.sample", "bazquxservice.sample", "",
		},
		"service succeed with serviceId enriched with business domain and bounded domain": {`let x = 42;
		//+gendoc category=root type=nameId
		bazquxservice.sample
		//-gendoc`,
			&parser.Config{BusinessDomain: "dom1", BoundedDomain: "area51"},
			parser.ServiceNode, "urn:dom1:area51:bazquxservice.sample", "bazquxservice.sample", "",
		},
		"channel succeed with id specified": {`let x = 42;
			//+gendoc category=channel type=description id=baz
			this is some description
			//-gendoc`,
			&parser.Config{ServiceId: "foo"},
			parser.ChannelNode, "", "baz", "foo",
		},
		"operation succeed with id specified": {`let x = 42;
			//+gendoc category=pubOperation type=description id=baz parent=parentBaz
			this is some description
			//-gendoc`,
			&parser.Config{ServiceId: "foo"},
			parser.OperationNode, "", "baz", "parentBaz",
		},
		"operation succeed with parent fallback": {`let x = 42;
		//+gendoc category=pubOperation type=description id=baz channelId=topic-bat
		this is some description
		//-gendoc`,
			&parser.Config{ServiceId: "foo"},
			parser.OperationNode, "", "baz", "topic-bat",
		},
		"message succeed with id specified": {`let x = 42;
			//+gendoc category=message type=description id=baz parent=bar
			this is some description
			//-gendoc`,
			&parser.Config{ServiceId: "foo"},
			parser.MessageNode, "", "baz", "bar",
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			lexerSource.Input = tt.input
			l := lexer.New(lexerSource)
			p := parser.New(l, tt.config).WithLogger(log.New(os.Stderr, log.ErrorLvl))
			got, errs := p.InitialParse()
			if len(errs) > 0 {
				t.Fatalf("parser had errors, expected <nil>\nerror: %v", errs)
			}
			if len(got) != 1 {
				t.Fatalf("expected 1 GenDocBlock to come back")
			}
			if got[0].NodeCategory != tt.wantCatType {
				t.Errorf("parser incorrectly converted node category\ngot: %v\nexpected %v\n", got[0].NodeCategory, tt.wantCatType)
			}
			if got[0].Annotation.Id != tt.wantId {
				t.Errorf("parser incorrectly converted node Id\ngot: %v\nexpected %v\n", got[0].Annotation.Id, tt.wantId)
			}
			if got[0].Annotation.ServiceURN != tt.wantUrn {
				t.Errorf("parser incorrectly converted node ServieURN\ngot: %v\nexpected %v\n", got[0].Annotation.ServiceURN, tt.wantUrn)
			}
			if got[0].Annotation.Parent != tt.wantParentId {
				t.Errorf("parser incorrectly converted node Parent\ngot: %v\nexpected %v\n", got[0].Annotation.Parent, tt.wantParentId)
			}
		})
	}
}

func Test_Parse_Id_and_type_of_nodes_should_fail(t *testing.T) {
	ttests := map[string]struct {
		input       string
		config      *parser.Config
		wantErrType error
	}{
		"service without id inherited in single repo mode": {`let x = 42;
			//+gendoc category=info type=description
			this is some description
			//-gendoc`,
			&parser.Config{ServiceId: ""},
			parser.ErrIdRequired,
		},
		"channel with id missing": {`let x = 42;
			//+gendoc category=channel type=description
this is some description
			//-gendoc`,
			&parser.Config{ServiceId: "foo"},
			parser.ErrIdRequired,
		},
		"channel with parent id missing": {`let x = 42;
		//+gendoc category=channel type=description id=bad
this is some description
		//-gendoc`,
			&parser.Config{},
			parser.ErrParentIdRequired,
		},
		"operation with parent id missing": {`let x = 42;
		//+gendoc category=pubOperation type=description id=bad
this is some description
		//-gendoc`,
			&parser.Config{},
			parser.ErrParentIdRequired,
		},
		"operation with id missing": {`let x = 42;
		//+gendoc category=pubOperation type=description
this is some description
		//-gendoc`,
			&parser.Config{},
			parser.ErrIdRequired,
		},
		"message with id missing": {`let x = 42;
		//+gendoc category=message type=description
this is some description
		//-gendoc`,
			&parser.Config{},
			parser.ErrIdRequired,
		},
		"message with description id missing": {`let x = 42;
		//+gendoc category=message id=bad
this is some description
		//-gendoc`,
			&parser.Config{},
			parser.ErrContentTypeRequired,
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			lexerSource.Input = tt.input
			l := lexer.New(lexerSource)
			p := parser.New(l, tt.config).WithLogger(log.New(os.Stderr, log.ErrorLvl))
			_, errs := p.InitialParse()
			if len(errs) == 0 {
				t.Fatalf("parser had NO errors, expected errors to be non nil")
			}
			if !errors.Is(errs[0], tt.wantErrType) {
				t.Errorf("incorrect error type, got: %v, want: %v", errs[0], tt.wantErrType)
			}
		})
	}
}

func Test_ShouldError_when_no_End_tag_found(t *testing.T) {
	input := `let x = 5;
	//+gendoc category=message type=nameId parent=id1 id=id
	`

	lexerSource.Input = input
	l := lexer.New(lexerSource)
	p := parser.New(l, &parser.Config{}).WithLogger(log.New(os.Stderr, log.ErrorLvl))
	_, errs := p.InitialParse()
	if len(errs) != 1 {
		t.Errorf("unexpected number of errors\n got: %v, wanted: 1", errs)
	}
	if !errors.Is(errs[0], parser.ErrNoEndTagFound) {
		t.Errorf("unexpected error type\n got: %T, wanted: %T", errs, parser.ErrNoEndTagFound)
	}
}

func Test_Error_on_unparseable_tag(t *testing.T) {
	input := `let x = 5;
	//+gendoc category=bar type=nameId
	//-gendoc
	`

	lexerSource.Input = input
	l := lexer.New(lexerSource)
	p := parser.New(l, &parser.Config{}).WithLogger(log.New(os.Stderr, log.ErrorLvl))
	_, errs := p.InitialParse()
	if len(errs) != 1 {
		t.Errorf("unexpected number of errors\n got: %v, wanted: 1", errs)
	}
	if !errors.Is(errs[0], gendoc.ErrIncorrectCategory) {
		t.Errorf("unexpected error type\n got: %T, wanted: %T", errs[0], gendoc.ErrIncorrectCategory)
	}
}

func testHelperGenDocBlock(t *testing.T, initialGenDocBlock parser.GenDocBlock, name string, content any) bool {
	if initialGenDocBlock.Token.Literal != "//+gendoc" {
		t.Errorf("got=%q, wanted initialGenDocBlock.TokenLiteral = '//+gendoc'.", initialGenDocBlock.Token.Literal)
		return false
	}

	if initialGenDocBlock.Value != content {
		t.Errorf("initialGenDocBlock.Value. got=%s, wanted=%s", initialGenDocBlock.Value, name)
		return false
	}

	if initialGenDocBlock.EndToken.Literal != "//-gendoc" {
		t.Errorf("initialGenDocBlock.EndToken incorrect\ngot=%v\nwanted non nil",
			initialGenDocBlock.EndToken)
		return false
	}
	return true
}

func Test_ExpandEnvVariables_succeeds(t *testing.T) {
	ttests := map[string]struct {
		input  string
		expect string
		envVar []string
	}{
		"with single var": {
			"some var is $var",
			"some var is foo",
			[]string{"var=foo"},
		},
		"with multiple var": {
			"some var is $var and docs go [here]($DOC_LINK/stuff)",
			"some var is foo and docs go [here](https://somestuff.com/stuff)",
			[]string{"var=foo", "DOC_LINK=https://somestuff.com"},
		},
		"with no vars in content": {
			"some var is foo and docs go [here](foo.com/stuff)",
			"some var is foo and docs go [here](foo.com/stuff)",
			[]string{"var=foo", "DOC_LINK=https://somestuff.com"},
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			defer os.Clearenv()
			got, err := parser.ExpandEnvVariables(tt.input, tt.envVar)
			if err != nil {
				t.Errorf("expected %v to be <nil>", err)
			}
			if got != tt.expect {
				t.Errorf("want: %s, got: %s", got, tt.expect)
			}
		})
	}
}

func Test_ExpandEnvVariables_fails(t *testing.T) {

	ttests := map[string]struct {
		input  string
		setup  func() func()
		envVar []string
	}{
		"with single var": {
			"some var is $var",
			func() func() {
				return func() {
					os.Clearenv()
				}
			},
			[]string{"v=foo"},
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			clear := tt.setup()
			defer clear()
			_, err := parser.ExpandEnvVariables(tt.input, tt.envVar)
			if err == nil {
				t.Errorf("wanted error, got <nil>")
			}
		})
	}
}

func Test_Parse_WithOwnEnviron_passed_in_succeeds(t *testing.T) {
	ttests := map[string]struct {
		input   string
		expect  string
		environ []string
	}{
		"test1": {
			input: `let x = 42;
//+gendoc category=message type=description id=foo
this is some description with $foo
//-gendoc`,
			environ: []string{"foo=bar"},
			expect:  "this is some description with bar",
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			defer os.Clearenv()
			lexerSource.Input = tt.input
			l := lexer.New(lexerSource)
			p := parser.New(l, &parser.Config{}).WithLogger(log.New(os.Stderr, log.ErrorLvl)).WithEnvironment(tt.environ)
			got, errs := p.InitialParse()
			if len(errs) > 0 {
				t.Error(errs)
			}
			if got[0].Value != tt.expect {
				t.Error("")
			}
		})
	}
}

func Test_Parse_WithOwnEnviron_passed_in_fails(t *testing.T) {
	ttests := map[string]struct {
		input   string
		expect  error
		environ []string
	}{
		"if variable is not set": {
			input: `let x = 42;
		//+gendoc category=message type=description id=foo
		this is some description with $foo
		//-gendoc`,
			expect:  parser.ErrUnableToReplaceVarPlaceholder,
			environ: []string{"notfoo=bar"},
		},
		"if variable is not set but empty": {
			input: `let x = 42;
//+gendoc category=message type=description id=foo
this is some description with $foo
//-gendoc`,
			expect:  parser.ErrUnableToReplaceVarPlaceholder,
			environ: []string{"foo="},
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			defer os.Clearenv()
			lexerSource.Input = tt.input
			l := lexer.New(lexerSource)
			p := parser.New(l, &parser.Config{}).WithLogger(log.New(os.Stderr, log.ErrorLvl)).WithEnvironment(tt.environ)
			_, errs := p.InitialParse()

			if len(errs) < 1 {
				t.Error("expected errors to occur")
				t.Fail()
			}
			if !errors.Is(errs[0], tt.expect) {
				t.Errorf("unexpected error type\n got: %T, wanted: %T", errs, tt.expect)
			}
		})
	}
}
