package parser_test

import (
	"testing"

	"github.com/dnitsch/async-api-generator/pkg/lexer"
	"github.com/dnitsch/async-api-generator/pkg/parser"
)

func Test_GenDocStatements(t *testing.T) {
	tests := map[string]struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		"simple": {`let x = 5;
//+gendoc foo
stuff {
	here string
}
//-gendoc
`, "//+gendoc", `stuff {
	here string
}`},
		// "bool":   {"let y = true;", "y", true},
		// "rand":   {"let foobar = y;", "foobar", "y"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		parsed := p.Parse()
		checkParserErrors(t, p)

		if len(parsed.Statements) != 2 {
			t.Fatalf("program.Statements does not contain 2 statements. got=%d",
				len(parsed.Statements))
		}

		stmt := parsed.Statements[1]
		if !testGenDocStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func testGenDocStatement(t *testing.T, s parser.Statement, name string) bool {
	if s.TokenLiteral() != "//+gendoc" {
		t.Errorf("got=%q, wanted s.TokenLiteral = '//+gendoc'.", s.TokenLiteral())
		return false
	}

	genStmt, ok := s.(*parser.GenDocStatement)
	if !ok {
		t.Errorf("s not *parser.GenDocStatement. got=%T", s)
		return false
	}

	if genStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, genStmt.Name.Value)
		return false
	}

	if genStmt.Name.TokenLiteral() != name {
		t.Errorf("letStmt.Name.TokenLiteral() not '%s'. got=%s",
			name, genStmt.Name.TokenLiteral())
		return false
	}

	return true
}

func checkParserErrors(t *testing.T, p *parser.Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
