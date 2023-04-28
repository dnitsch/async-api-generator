package ast_test

import (
	"testing"

	"github.com/dnitsch/async-api-generator/pkg/ast"
	"github.com/dnitsch/async-api-generator/pkg/token"
)

func TestString(t *testing.T) {
	program := &ast.GenDoc{
		Statements: []ast.Statement{
			&ast.IgnoreStatement{
				Token: token.Token{Type: token.TEXT, Literal: "beginning"},
				Name: &ast.EnclosedIdentifier{
					Token: token.Token{Type: token.SPACE, Literal: " "},
					Value: " ",
				},
				Value: &ast.EnclosedIdentifier{
					Token: token.Token{Type: token.NEW_LINE, Literal: "\n"},
					Value: "\n",
				},
			},
			&ast.GenDocStatement{
				Token: token.Token{Type: token.BEGIN_DOC_GEN, Literal: "// gendoc", MetaTags: "type=message,subtype=example,consumer=[],producer=[]"},
				Name: &ast.EnclosedIdentifier{
					Token: token.Token{Type: token.CONTENT_DOC_GEN, Literal: "content_gendoc"},
					Value: "",
				},
				Value: &ast.EnclosedIdentifier{
					Token: token.Token{Type: token.TEXT, Literal: "text"},
					Value: "", //
				},
			},
		},
	}

	if program.String() != `beginning 
// gendoc type=message,subtype=example,consumer=[],producer=[]` {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}
