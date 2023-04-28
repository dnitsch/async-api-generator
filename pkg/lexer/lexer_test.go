package lexer_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dnitsch/async-api-generator/pkg/lexer"
	"github.com/dnitsch/async-api-generator/pkg/token"
)

func Test(t *testing.T) {
	input := `foo stuyfsdfsf
// gendoc type=message,subtype=example,consumer=[],producer=[]
class {
	stuff string {get; set;}
}
// !gendoc

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
		{token.BEGIN_DOC_GEN, "// gendoc"},
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
		{token.END_DOC_GEN, "// !gendoc"},
		{token.NEW_LINE, "\n"},
		{token.NEW_LINE, "\n"},
		{token.TEXT, "//"},
		{token.FORWARD_SLASH, "/"},
		{token.SPACE, " "},
		{token.TEXT, "<summary>"},
		{token.SPACE, " "},
		{token.TEXT, "ignorethis"},
		{token.NEW_LINE, "\n"},
		{token.TEXT, "#"},
		{token.SPACE, " "},
		{token.TEXT, "another"},
		{token.SPACE, " "},
		{token.TEXT, "comment"},
		{token.NEW_LINE, "\n"},
	}
	l := lexer.New(input)

	for i, tt := range ttests {

		tok := l.NextToken()
		fmt.Println(tok)
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. got=%q, expected=%q",
				i, tok.Type, tt.expectedType)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. got=%q, expected=%q",
				i, tok.Literal, tt.expectedLiteral)
		}
		if tok.Type == token.BEGIN_DOC_GEN && len(tok.MetaTags) < 1 && strings.EqualFold(tok.MetaTags, " type=message,subtype=example,consumer=[],producer=[]\n") {
			// if tok.MetaTags
			t.Errorf("gendoc token should include ")
		}
	}
}
