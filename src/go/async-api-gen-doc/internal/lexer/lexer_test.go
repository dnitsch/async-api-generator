package lexer_test

import (
	"testing"

	"github.com/dnitsch/async-api-generator/internal/lexer"
	"github.com/dnitsch/async-api-generator/internal/token"
)

func Test_simple_with_gendoc_found(t *testing.T) {
	input := `foo stuyfsdfsf
/* som multiline comment
//+gendoc type=message subtype=example consumer=[] producer=[] \
multiline=val2
endmultilinecommentand_successfully_extract_tag
*/
class {
	stuff string {get; set;}
}
//-gendoc

/// <summary> ignorethis
# another comment
`
	ttests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.TEXT, "foo"},
		{token.SPACE, " "},
		{token.TEXT, "stuyfsdfsf"},
		{token.NEW_LINE, "\n"},
		{token.FORWARD_SLASH, "/"},
		{token.TEXT, "*"},
		{token.SPACE, " "},
		{token.TEXT, "som"},
		{token.SPACE, " "},
		{token.TEXT, "multiline"},
		{token.SPACE, " "},
		{token.TEXT, "comment"},
		{token.NEW_LINE, "\n"},
		{token.BEGIN_DOC_GEN, "//+gendoc"},
		{token.NEW_LINE, "\n"},
		{token.TEXT, "endmultilinecommentand_successfully_extract_tag"},
		{token.NEW_LINE, "\n"},
		{token.TEXT, "*/"},
		{token.NEW_LINE, "\n"},
		{token.TEXT, "class"},
		{token.SPACE, " "},
		{token.TEXT, "{"},
		{token.NEW_LINE, "\n"},
		{token.TAB, "\t"},
		{token.TEXT, "stuff"},
		{token.SPACE, " "},
		{token.TEXT, "string"},
		{token.SPACE, " "},
		{token.TEXT, "{get;"},
		{token.SPACE, " "},
		{token.TEXT, "set;}"},
		{token.NEW_LINE, "\n"},
		{token.TEXT, "}"},
		{token.NEW_LINE, "\n"},
		{token.END_DOC_GEN, "//-gendoc"},
		{token.NEW_LINE, "\n"},
		{token.NEW_LINE, "\n"},
		{token.DOUBLE_FORWARD_SLASH, "//"},
		{token.FORWARD_SLASH, "/"},
		{token.SPACE, " "},
		{token.TEXT, "<"},
		{token.TEXT, "summary>"},
		{token.SPACE, " "},
		{token.TEXT, "ignorethis"},
		{token.NEW_LINE, "\n"},
		{token.HASH, "#"},
		{token.SPACE, " "},
		{token.TEXT, "another"},
		{token.SPACE, " "},
		{token.TEXT, "comment"},
		{token.NEW_LINE, "\n"},
		{token.EOF, ""},
	}
	l := lexer.New(lexer.Source{Input: input, FullPath: "/foo/bar", FileName: "bar"})

	for i, tt := range ttests {

		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. got=%q, expected=%q",
				i, tok.Type, tt.expectedType)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. got=%q, expected=%q",
				i, tok.Literal, tt.expectedLiteral)
		}
		if tok.Type == token.BEGIN_DOC_GEN {
			metaWant := "type=message subtype=example consumer=[] producer=[] multiline=val2"
			if len(tok.MetaAnnotation) < 1 || tok.MetaAnnotation != metaWant {
				t.Errorf("gendoc meta annotation:\ngot: %s\nwanted %s", tok.MetaAnnotation, metaWant)
			}
		}
	}
}

func Test_simple_with_gendoc_with_comment(t *testing.T) {
	input := `
<!-- //+gendoc type=message subtype=example consumer=[] producer=[] \
-->
<!-- //-gendoc
<!-- dont care text -->
`
	ttests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.NEW_LINE, "\n"},
		{token.BEGIN_HTML_COMMENT, "<!--"},
		{token.BEGIN_DOC_GEN, "//+gendoc"},
		{token.NEW_LINE, "\n"},
		{token.BEGIN_HTML_COMMENT, "<!--"},
		{token.END_DOC_GEN, "//-gendoc"},
		{token.NEW_LINE, "\n"},
		{token.BEGIN_HTML_COMMENT, "<!--"},
		{token.TEXT, "dont"},
		{token.SPACE, " "},
		{token.TEXT, "care"},
		{token.SPACE, " "},
		{token.TEXT, "text"},
		{token.SPACE, " "},
		{token.END_HTML_COMMENT, "-->"},
		{token.NEW_LINE, "\n"},
		{token.EOF, ""},
	}
	l := lexer.New(lexer.Source{Input: input, FullPath: "/foo/bar", FileName: "bar"})

	for i, tt := range ttests {

		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. got=%q, expected=%q",
				i, tok.Type, tt.expectedType)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. got=%q, expected=%q",
				i, tok.Literal, tt.expectedLiteral)
		}
		if tok.Type == token.BEGIN_DOC_GEN {
			metaWant := "type=message subtype=example consumer=[] producer=[] -->"
			if len(tok.MetaAnnotation) < 1 || tok.MetaAnnotation != metaWant {
				t.Errorf("gendoc meta annotation:\ngot: %s\nwanted %s", tok.MetaAnnotation, metaWant)
			}
		}
	}
}

func Test_simple_with_incpomplete_html_comment(t *testing.T) {
	input := `
<!- //+gendoc type=message subtype=example consumer=[] producer=[]
->
<! //-gendoc - <
<!-- dontcaretext -->
`
	ttests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.NEW_LINE, "\n"},
		{token.TEXT, "<!-"},
		{token.SPACE, " "},
		{token.BEGIN_DOC_GEN, "//+gendoc"},
		{token.NEW_LINE, "\n"},
		{token.TEXT, "-"},
		{token.TEXT, ">"},
		{token.NEW_LINE, "\n"},
		{token.TEXT, "<!"},
		{token.SPACE, " "},
		{token.END_DOC_GEN, "//-gendoc"},
		{token.SPACE, " "},
		{token.TEXT, "-"},
		{token.SPACE, " "},
		{token.TEXT, "<"},
		{token.NEW_LINE, "\n"},
		{token.BEGIN_HTML_COMMENT, "<!--"},
		{token.TEXT, "dontcaretext"},
		{token.SPACE, " "},
		{token.END_HTML_COMMENT, "-->"},
		{token.NEW_LINE, "\n"},
		{token.EOF, ""},
	}
	l := lexer.New(lexer.Source{Input: input, FullPath: "/foo/bar", FileName: "bar"})

	for i, tt := range ttests {

		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. got=%q, expected=%q",
				i, tok.Type, tt.expectedType)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. got=%q, expected=%q",
				i, tok.Literal, tt.expectedLiteral)
		}
		if tok.Type == token.BEGIN_DOC_GEN {
			metaWant := "type=message subtype=example consumer=[] producer=[]"
			if len(tok.MetaAnnotation) < 1 || tok.MetaAnnotation != metaWant {
				t.Errorf("gendoc meta annotation:\ngot: %s\nwanted %s", tok.MetaAnnotation, metaWant)
			}
		}
	}
}

func Test_empty_file(t *testing.T) {
	input := ``
	l := lexer.New(lexer.Source{Input: input, FullPath: "/foo/bar", FileName: "bar"})
	tok := l.NextToken()
	if tok.Type != token.EOF {
		t.Fatal("expected EOF")
	}
}

// var inputNoEOF = `# Only trigger builds for main or tags (release).

//   # Directory and file paths
//   - name: helm_directory
// 	value: $(system.defaultWorkingDirectory)/deploy/helm

// 				resources.requests.memory=$(memory_request)
// 				resources.limits.cpu=$(cpu_limit)
// 				resources.limits.memory=$(memory_limit)
// 				autoscaling.minReplicas=$(min_service_replicas)
// 				autoscaling.maxReplicas=$(max_service_replicas)
// 				autoscaling.cpuTargetUtilisation=$(cpu_target_percentage)`

// func Test_input_with_no_EOF(t *testing.T) {
// 	l := lexer.New(lexer.Source{Input: inputNoEOF, FullPath: "/foo/bar", FileName: "bar"})
// 	// tok := l.NextToken()
// 	for (l.NextToken()).Type != token.EOF {
// 		fmt.Printf("token: %s\n", tok.Type)
// 	}

// 	t.Fatal("expected EOF")
// 	// if tok.Type == token.EOF {
// 	// 	t.Fatal("expected EOF")
// 	// }
// }
