package parser_test

import (
	"testing"

	"github.com/dnitsch/async-api-generator/pkg/ast"
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
// gendoc foo
stuff {
	here string
}
// !gendoc
`, "// gendoc", `stuff {
	here string
}`},
		// "bool":   {"let y = true;", "y", true},
		// "rand":   {"let foobar = y;", "foobar", "y"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.Parse()
		checkParserErrors(t, p)

		if len(program.Statements) != 2 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d",
				len(program.Statements))
		}

		stmt := program.Statements[1]
		if !testGenDocStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func testGenDocStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "// gendoc" {
		t.Errorf("got=%q, wanted s.TokenLiteral = '// gendoc'.", s.TokenLiteral())
		return false
	}

	genStmt, ok := s.(*ast.GenDocStatement)
	if !ok {
		t.Errorf("s not *ast.GenDocStatement. got=%T", s)
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
