package token

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
	TEXT            TokenType = "TEXT"            //
	CONTENT_DOC_GEN TokenType = "CONTENT_DOC_GEN" //

	EXCLAMATION TokenType = "!"

	BACK_SLASH           TokenType = "BACK_SLASH"           // \
	FORWARD_SLASH        TokenType = "FORWARD_SLASH"        // `/`
	DOUBLE_FORWARD_SLASH TokenType = "DOUBLE_FORWARD_SLASH" // `//`

	// Keywords
	BEGIN_DOC_GEN TokenType = "BEGIN_DOC_GEN"
	END_DOC_GEN   TokenType = "END_DOC_GEN"
)

// Token is the basic structure of the captured token
type Token struct {
	Type           TokenType
	Literal        string
	MetaAnnotation string // additional info about the captured token
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
