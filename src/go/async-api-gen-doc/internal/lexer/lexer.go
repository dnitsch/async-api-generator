// Package lexer
//
// Performs lexical analysis on the source files and emits tokens.
package lexer

import (
	"strings"

	"github.com/dnitsch/async-api-generator/internal/token"
)

const (
	// Literals
	BEGIN_DOC string = "+gendoc"
	END_DOC   string = "-gendoc"
)

// nonText characters captures all character sets that are _not_ assignable to TEXT
var nonText = map[string]bool{" ": true, "\n": true, "\r": true, "\t": true}

type Source struct {
	Input    string
	FileName string
	FullPath string
}

// Lexer
type Lexer struct {
	length       int
	source       Source
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int  // current line - start at 1
	column       int  // column of text - gets set to 0 on every new line - start at 0
}

// New returns a Lexer pointer allocation
func New(source Source) *Lexer {

	l := &Lexer{source: source, line: 1, column: 0, length: len(source.Input)}
	l.readChar()
	return l
}

// NextToken advances through the source returning a found token
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	switch l.ch {
	case '/':
		if l.peekChar() == '/' {
			// if next char is a `/` then we have to consume it from lexer
			l.readChar()
			if l.peekIsDocGenBegin() {
				tok = l.readDocAnnotation(token.Token{Type: token.BEGIN_DOC_GEN, Literal: "//+gendoc"})
			} else if l.peekIsDocGenEnd() {
				tok = token.Token{Type: token.END_DOC_GEN, Literal: "//-gendoc"}
			} else {
				// it is not a doc gen marker assigning double slash as text
				tok = token.Token{Type: token.DOUBLE_FORWARD_SLASH, Literal: "//"}
			}
		} else {
			tok = token.Token{Type: token.FORWARD_SLASH, Literal: "/"}
		}
	case '#':
		tok = token.Token{Type: token.HASH, Literal: "#"}
	// check if we are in an MarkDown/HTML comment block
	// potential begin comment
	case '<':
		tok = l.htmlCommentTextToken("<!--", token.BEGIN_HTML_COMMENT)
	// potential end comment
	case '-':
		tok = l.htmlCommentTextToken("-->", token.END_HTML_COMMENT)
	case '\n':
		l.line = l.line + 1
		l.column = 0 // reset column count
		tok = l.setTextSeparatorToken()
		// want to preserve all indentations and punctuation
	case ' ', '\r', '\t', '\f':
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
	// add general properties to each token
	tok.Line = l.line
	tok.Column = l.column
	tok.Source = token.Source{Path: l.source.FullPath, File: l.source.FileName}
	l.readChar()
	return tok
}

// readChar moves cursor along
func (l *Lexer) readChar() {
	if l.readPosition >= l.length {
		l.ch = 0
	} else {
		l.ch = l.source.Input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
	l.column += 1
}

// peekChar reveals next char withouh advancing the cursor along
func (l *Lexer) peekChar() byte {
	if l.readPosition >= l.length {
		return 0
	} else {
		return l.source.Input[l.readPosition]
	}
}

func (l *Lexer) readText() string {
	position := l.position
	for isText(l.ch) && l.readPosition <= l.length {
		l.readChar()
	}
	return l.source.Input[position:l.position]
}

func (l *Lexer) setTextSeparatorToken() token.Token {
	tok := newToken(token.LookupIdent(string(l.ch)), l.ch)
	return tok
}

// readDocAnnotation reads the rest of the line identified by
func (l *Lexer) readDocAnnotation(tok token.Token) token.Token {
	metaTag := ""
Loop:
	for {
		peekChar := l.peekChar()
		// if NEXT CHAR is line break AND the current byte is NOT `0x5c` (`\`)
		// we exit the forever loop as we have reached the end of a meta annotation string
		if peekChar == '\n' && l.ch != 0x5c { //0x5c
			break Loop
		}
		// if current byte is `\` and is followed by line break
		// we continue to build the meta-anotation just broken over
		// many lines but skip adding them to the metatag
		// strip the '\n' and '\\' chars from the metaTag
		if peekChar != '\n' && peekChar != 0x5c {
			// the parser will strip any non key/value identifiers
			metaTag += string(peekChar)
		}
		l.readChar()
	}
	tok.MetaAnnotation = strings.TrimSpace(metaTag)
	return tok
}

// peekIsDocGenBegin attempts to identify the gendoc keyword after 2 slashes
func (l *Lexer) peekIsDocGenBegin() bool {
	count := 0
	docGen := ""
	for count < len(BEGIN_DOC) {
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
	for count < len(END_DOC) {
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

func (l *Lexer) htmlCommentTextToken(commentLiteral string, commentType token.TokenType) token.Token {
	length := len(commentLiteral)
	for i := range commentLiteral {
		if i < length {
			idx := i + 1
			if idx == length { // is last char and already peeked and matched
				break
			}
			if l.peekChar() == commentLiteral[idx] {
				l.readChar()
			} else {
				return token.Token{Type: token.TEXT, Literal: commentLiteral[0:idx]}
			}
		}
	}
	// special case: swallow all whitespace until next token
	for l.peekChar() == ' ' {
		l.readChar()
	}
	return token.Token{Type: commentType, Literal: commentLiteral}
}

// resetAfterPeek will go back specified amount on the cursor
func (l *Lexer) resetAfterPeek(back int) {
	l.position = l.position - back
	l.readPosition = l.readPosition - back
}

// isText only deals with any text characters defined as
// outside of the capture group
func isText(ch byte) bool {
	return !nonText[string(ch)]
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}
