package parser

import (
	"fmt"

	"github.com/dnitsch/async-api-generator/pkg/lexer"
	"github.com/dnitsch/async-api-generator/pkg/token"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // array[index]
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) Parse() *GenDoc {
	program := &GenDoc{}
	program.Statements = []Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() Statement {
	switch p.curToken.Type {
	case token.BEGIN_DOC_GEN:
		return p.parseBeginGenDocStatement()
	default:
		return p.parseIgnoreStatement()
	}
}

func (p *Parser) parseBeginGenDocStatement() *GenDocStatement {
	stmt := &GenDocStatement{Token: p.curToken}
	// do some parsing here perhaps of the name and file name/location etc...
	stmt.Name = &EnclosedIdentifier{Token: token.Token{Type: token.BEGIN_DOC_GEN, Literal: "//+gendoc", MetaAnnotation: p.curToken.MetaAnnotation}, Value: "//+gendoc"} //p.curToken.MetaTags}
	// move past gendoc token
	p.nextToken()

	genDocValue := ""
	// should exit the loop if no end doc tag found
	for {
		genDocValue += p.curToken.Literal
		if p.peekTokenIs(token.END_DOC_GEN) {
			p.nextToken()
			break
		}
		p.nextToken()
	}

	stmt.Value = &EnclosedIdentifier{Token: token.Token{Type: token.CONTENT_DOC_GEN, Literal: genDocValue, MetaAnnotation: stmt.Token.MetaAnnotation}, Value: genDocValue}
	// skip end doc
	p.nextToken()
	return stmt
}

func (p *Parser) parseIgnoreStatement() *IgnoreStatement {
	stmt := &IgnoreStatement{Token: p.curToken}

	ignoreValue := ""
	for !p.curTokenIs(token.EOF) {
		ignoreValue += p.curToken.Literal
		if p.peekTokenIs(token.BEGIN_DOC_GEN) {
			break
		}
		p.nextToken()
	}
	stmt.Value = &UnusedIdentifier{Token: token.Token{Type: token.NEW_LINE, Literal: ignoreValue}, Value: ignoreValue}
	return stmt
}

func (p *Parser) parseExpression(precedence int) Expression {
	return &EnclosedIdentifier{}
}
