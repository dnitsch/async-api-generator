package ast

import (
	"bytes"
	"fmt"

	"github.com/dnitsch/async-api-generator/pkg/token"
)

// Node
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement must implement Node interface
// e.g. DOC_GEN statement
type Statement interface {
	Node
	statementNode()
}

// Expression
// e.g. TEXT inside the
type Expression interface {
	Node
	expressionNode()
}

type GenDoc struct {
	Statements []Statement
}

func (p *GenDoc) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *GenDoc) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

/*
 * Statements will encapsulate expressions so they can be later evaluated
 */
// GenDocStatement
type GenDocStatement struct {
	Token token.Token // token.BEGIN_GEN_DOC token
	Name  *EnclosedIdentifier
	Value Expression
}

func (ls *GenDocStatement) statementNode()       {}
func (ls *GenDocStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *GenDocStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral())
	out.WriteString(fmt.Sprintf("%s %s", ls.Name.String(), ls.Token.MetaTags))

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	return out.String()
}

// IgnoreStatement
type IgnoreStatement struct {
	Token token.Token // the token.LET token
	Name  *EnclosedIdentifier
	Value Expression
}

func (ls *IgnoreStatement) statementNode()       {}
func (ls *IgnoreStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *IgnoreStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral())
	out.WriteString(ls.Name.String())
	// out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	// out.WriteString(";")
	return out.String()
}

// Expressions
type EnclosedIdentifier struct {
	Token token.Token // the token.IDENT token
	Value string
}

func (i *EnclosedIdentifier) expressionNode()      {}
func (i *EnclosedIdentifier) TokenLiteral() string { return i.Token.Literal }
func (i *EnclosedIdentifier) String() string       { return i.Value }

type UnusedIdentifier struct {
	Token token.Token // the token.IDENT token
	Value string
}

func (i *UnusedIdentifier) expressionNode()      {}
func (i *UnusedIdentifier) TokenLiteral() string { return i.Token.Literal }
func (i *UnusedIdentifier) String() string       { return i.Value }
