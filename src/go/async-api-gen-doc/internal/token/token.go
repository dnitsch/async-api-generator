package token

import (
	"strings"
)

type TokenType string

const (
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	SPACE           TokenType = "SPACE"           // ' '
	TAB             TokenType = "TAB"             // '\t'
	NEW_LINE        TokenType = "NEW_LINE"        // '\n'
	CARRIAGE_RETURN TokenType = "CARRIAGE_RETURN" // '\r'
	CONTROL         TokenType = "CONTROL"

	// Identifiers + literals
	TEXT TokenType = "TEXT"

	EXCLAMATION TokenType = "!"

	BACK_SLASH    TokenType = "BACK_SLASH"    // \
	FORWARD_SLASH TokenType = "FORWARD_SLASH" // `/`
	// Comment Tokens
	DOUBLE_FORWARD_SLASH TokenType = "DOUBLE_FORWARD_SLASH" // `//`
	HASH                 TokenType = "HASH"                 // `#`
	BEGIN_HTML_COMMENT   TokenType = "BEGIN_HTML_COMMENT"   // `<!--`
	END_HTML_COMMENT     TokenType = "END_HTML_COMMENT"     // `-->`

	// DOC_GEN Keywords
	BEGIN_DOC_GEN TokenType = "BEGIN_DOC_GEN"
	META_DOC_GEN  TokenType = "META_DOC_GEN"
	END_DOC_GEN   TokenType = "END_DOC_GEN"

	// Parsed "expressions"
	GEN_DOC_CONTENT TokenType = "GEN_DOC_CONTENT"
	UNUSED_TEXT     TokenType = "UNUSED_TEXT"
	MESSAGE         TokenType = "MESSAGE"
	OPERATION       TokenType = "OPERATION"
	CHANNEL         TokenType = "CHANNEL"
	INFO            TokenType = "INFO"
	SERVER          TokenType = "SERVER"
)

type Source struct {
	File string `json:"file"`
	Path string `json:"path"`
}

// Token is the basic structure of the captured token
type Token struct {
	Type           TokenType `json:"type"`
	Literal        string    `json:"literal"`
	MetaAnnotation string    `json:"annotationLiteral"` //parser.GenDocMetaAnnotation additional info about the captured token
	Line           int       `json:"line"`
	Column         int       `json:"column"`
	Source         Source    `json:"source"`
}

var keywords = map[string]TokenType{
	" ":  SPACE,
	"\n": NEW_LINE,
	"\r": CARRIAGE_RETURN,
	"\t": TAB,
	"\f": CONTROL,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TEXT
}

var typeMapper = map[string]TokenType{
	"MESSAGE":   MESSAGE,
	"OPERATION": OPERATION,
	"CHANNEL":   CHANNEL,
	"INFO":      INFO,
	"SERVER":    SERVER,
}

func LookupType(typ string) TokenType {
	if tok, ok := typeMapper[strings.ToUpper(typ)]; ok {
		return tok
	}
	return ""
}
