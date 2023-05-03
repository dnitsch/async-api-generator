package parser_test

import (
	"testing"

	"github.com/dnitsch/async-api-generator/pkg/parser"
	"github.com/dnitsch/async-api-generator/pkg/token"
)

func TestString(t *testing.T) {
	program := &parser.GenDoc{
		Statements: []parser.Statement{
			&parser.IgnoreStatement{
				Token: token.Token{Type: token.TEXT, Literal: "beginning"},
				Name: &parser.EnclosedIdentifier{
					Token: token.Token{Type: token.SPACE, Literal: " "},
					Value: " ",
				},
				Value: &parser.EnclosedIdentifier{
					Token: token.Token{Type: token.NEW_LINE, Literal: "\n"},
					Value: "\n",
				},
			},
			&parser.GenDocStatement{
				Token: token.Token{Type: token.BEGIN_DOC_GEN, Literal: "//+gendoc", MetaAnnotation: "type=message,subtype=example,consumer=[],producer=[]"},
				Name: &parser.EnclosedIdentifier{
					Token: token.Token{Type: token.CONTENT_DOC_GEN, Literal: "content_gendoc"},
					Value: "",
				},
				Value: &parser.EnclosedIdentifier{
					Token: token.Token{Type: token.TEXT, Literal: "text"},
					Value: "", //
				},
			},
		},
	}

	if program.String() != `beginning 
//+gendoc type=message,subtype=example,consumer=[],producer=[]` {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}
