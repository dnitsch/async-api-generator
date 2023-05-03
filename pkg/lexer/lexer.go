package lexer

import (
	"strings"

	"github.com/dnitsch/async-api-generator/pkg/token"
)

const (
	// Literals
	BEGIN_DOC string = "+gendoc"
	END_DOC   string = "-gendoc"
)

// nonText characters captures all character sets that are _not_ assignable to TEXT
var nonText = map[string]bool{" ": true, "\n": true, "\r": true, "\t": true}

// Lexer
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

// New returnds a Lexer pointer
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

// NextToken advances through the source returning a found token
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	switch l.ch {
	case '/':
		if l.peekChar() == '/' {
			// if next rune is a `/` then we have to consume it from lexer
			l.readChar()
			if l.peekIsDocGenBegin() {
				tok = l.readDocAnnotation(token.Token{Type: token.BEGIN_DOC_GEN, Literal: "//+gendoc"})
				// return l.readDocAnnotation(token.Token{Type: token.BEGIN_DOC_GEN, Literal: "// gendoc"})
			} else if l.peekIsDocGenEnd() {
				tok = token.Token{Type: token.END_DOC_GEN, Literal: "//-gendoc"}
			} else {
				// it is not a doc gen marker assigning double slash as text
				tok = token.Token{Type: token.TEXT, Literal: "//"}
			}
		} else {
			tok = token.Token{Type: token.FORWARD_SLASH, Literal: "/"}
		}
	// want to preserve all indentations and punctuation
	case ' ', '\n', '\r', '\t', '\f':
		tok = l.setTextSeparatorToken()
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isText(l.ch) {
			tok.Literal = l.readText()
			tok.Type = token.TEXT
			return tok
		}
		tok = newToken(token.ILLEGAL, l.ch)
	}

	l.readChar()
	return tok
}

// readChar moves cursor along
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) readText() string {
	position := l.position
	for isText(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) setTextSeparatorToken() token.Token {
	tok := newToken(token.LookupIdent(string(l.ch)), l.ch)
	return tok
}

// readDocAnnotation reads the rest of the line identified by
func (l *Lexer) readDocAnnotation(tok token.Token) token.Token {
	metaTag := ""
	for l.peekChar() != '\n' {
		metaTag += string(l.peekChar())
		l.readChar()
	}
	tok.MetaAnnotation = strings.TrimSpace(metaTag)
	return tok
}

// peekIsDocGenBegin attempts to identify the gendoc keyword after 2 slashes
func (l *Lexer) peekIsDocGenBegin() bool {
	count := 0
	docGen := ""
	for count < 7 {
		count++
		docGen += string(l.peekChar())
		l.readChar()
	}

	if docGen == BEGIN_DOC {
		return true
	}
	l.resetAfterPeek(len(BEGIN_DOC))
	return false
}

// peekIsDocGenEnd
func (l *Lexer) peekIsDocGenEnd() bool {
	count := 0
	docGen := ""
	for count < 7 {
		count++
		docGen += string(l.peekChar())
		l.readChar()
	}
	if strings.EqualFold(docGen, END_DOC) {
		return true
	}
	l.resetAfterPeek(len(END_DOC))
	return false
}

// resetAfterPeek will go back specified amount on the cursor
func (l *Lexer) resetAfterPeek(back int) {
	l.position = l.position - back
	l.readPosition = l.readPosition - back
}

// isText only deals with any
func isText(ch byte) bool {
	return !nonText[string(ch)]
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}
