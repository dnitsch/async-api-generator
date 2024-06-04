package parser

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/a8m/envsubst"
	"github.com/dnitsch/async-api-generator/internal/gendoc"
	"github.com/dnitsch/async-api-generator/internal/lexer"
	"github.com/dnitsch/async-api-generator/internal/token"
	log "github.com/dnitsch/simplelog"
)

func wrapErr(file string, line, position int, etyp error) error {
	return fmt.Errorf("\n - [%s:%d:%d] %w", file, line, position, etyp)
}

var (
	ErrNoEndTagFound                 = errors.New("no corresponding //-gendoc end tag found")
	ErrUnableToReplaceVarPlaceholder = errors.New("variable specified in the content was not found in the environment")
	ErrUnableToConvertCategory       = errors.New("unable to convert category to node category")
	ErrIdRequired                    = errors.New("id must be specified")
	ErrContentTypeRequired           = errors.New("content type must be specified")
	ErrParentIdRequired              = errors.New("parent must be specified")
)

type Parser struct {
	l         *lexer.Lexer
	errors    []error
	log       log.Loggeriface
	curToken  token.Token
	peekToken token.Token
	config    *Config
	environ   []string
}

func New(l *lexer.Lexer, c *Config) *Parser {
	p := &Parser{
		l:       l,
		log:     log.New(os.Stderr, log.ErrorLvl),
		errors:  []error{},
		config:  c,
		environ: os.Environ(),
	}

	// Read two tokens, so curToken and peekToken are both set
	// first one sets the curToken to the value of peekToken -
	// which at this point is just the first upcoming token
	p.nextToken()
	// second one sets the curToken to the actual value of the first upcoming
	// token and peekToken is the actual second upcoming token
	p.nextToken()

	return p
}

func (p *Parser) WithEnvironment(environ []string) *Parser {
	p.environ = environ
	return p
}

type AnalysisMode string

const (
	Validate   AnalysisMode = "validate" //analysis only
	SingleRepo AnalysisMode = "single_context"
	AllRepo    AnalysisMode = "global_context"
)

var AnalysisModeEnum = map[string]AnalysisMode{
	"validate":       Validate,
	"single_context": SingleRepo,
	"global_context": AllRepo,
}

// Config holds the parser config
type Config struct {
	ServiceId       string
	ServiceRepoUrl  string
	ServiceLanguage string
	BusinessDomain  string // Business level domain i.e. warehouse
	BoundedDomain   string // BoundDomain within a business domain
	// Note: other properties can go here
	// perhaps better to use the options pattern
	// ...apply(opt)
}

func (p *Parser) WithLogger(logger log.Loggeriface) *Parser {
	p.log = nil //speed up GC
	p.log = logger
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) currentTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// InitialParse creates a flat list of GenDocBlock
// and parsed annotations as "expressions"
//
// Currently, there will only ever be a single level nesting
// i.e. GenDoc statement will only ever contain a single expression
// this intermediate output will enable the
// advanced parsing of building an N-Ary tree based on precedence.
//
// Returns values as they are not meant to be mutated past this point.
func (p *Parser) InitialParse() ([]GenDocBlock, []error) {
	genDocStms := []GenDocBlock{}

	for !p.currentTokenIs(token.EOF) {
		if p.currentTokenIs(token.BEGIN_DOC_GEN) {
			// parseGenDocBlocks will advance the token until
			// it hits the END_DOC_GEN token
			if stmt := p.parseGenDocBlocks(); stmt != nil {
				genDocStms = append(genDocStms, *stmt)
			}
		}
		p.nextToken()
	}

	return genDocStms, p.errors
}

// parseGenDocBlocks throws away all other content other
// than what is inside //+gendoc tags
// parses any annotation and creates GenDocBlock
// for later analysis
func (p *Parser) parseGenDocBlocks() *GenDocBlock {
	genDocToken := p.curToken
	stmt := &GenDocBlock{Token: genDocToken}
	// do some parsing here perhaps of the name and file name/location etc...
	genDocMeta, err := gendoc.New(genDocToken.MetaAnnotation, p.log)
	if err != nil {
		p.errors = append(p.errors, wrapErr(genDocToken.Source.File, genDocToken.Line, genDocToken.Column, err))
		return nil
	}

	// move past gendoc token
	p.nextToken()
	// move past \n
	p.nextToken()

	contentVal := ""
	// should exit the loop if no end doc tag found
	notFoundEnd := true
	// stop on end of file
	for !p.peekTokenIs(token.EOF) {

		// for cases where the body is empty
		if p.currentTokenIs(token.END_DOC_GEN) {
			stmt.EndToken = p.curToken
			notFoundEnd = false
			break
		}

		// if currToken is some sort of comment character and peekTokenIs endDocGen
		if p.peekTokenIs(token.END_DOC_GEN) && (p.currentTokenIs(token.END_HTML_COMMENT) || p.currentTokenIs(token.BEGIN_HTML_COMMENT)) {
			notFoundEnd = false
			p.nextToken()
			stmt.EndToken = p.curToken
			break
		}

		// when next token is end doc
		// we skip assigning to the literal
		// we consume it and move on
		if p.peekTokenIs(token.END_DOC_GEN) {
			notFoundEnd = false
			p.nextToken()
			stmt.EndToken = p.curToken
			break
		}

		// TODO: add to built string if it isn't
		// a comment character after line break
		// for cases when CONTENT_DOC_GEN is inside comment blocks
		// \n# or \n//
		//
		// BUG: the only way for now to have valid multiline annotation
		// is if it's wrapped in langauge specific multiline comments
		// e.g. /* ... */
		contentVal += p.curToken.Literal
		p.nextToken()
	}

	if notFoundEnd {
		p.errors = append(p.errors, wrapErr(genDocToken.Source.File, genDocToken.Line, genDocToken.Column, ErrNoEndTagFound))
		return nil
	}

	val, err := ExpandEnvVariables(contentVal, p.environ)

	if err != nil {
		p.errors = append(p.errors, wrapErr(genDocToken.Source.File, genDocToken.Line, genDocToken.Column, fmt.Errorf("%v - %w", err, ErrUnableToReplaceVarPlaceholder)))
	}

	stmt.Value = val

	// NOTE: This will never faile unless categories are extended
	// and the mapping is not extended - GenDoc tests will catch that!
	nodeCat := nodeCatConverter[string(genDocMeta.CategoryType)]

	stmt.NodeCategory = nodeCat
	ant, err := p.parseAnnotation(genDocMeta, nodeCat, stmt)
	if err != nil {
		p.errors = append(p.errors, wrapErr(genDocToken.Source.File, genDocToken.Line, genDocToken.Column, err))
		return nil
	}
	stmt.Annotation = ant

	// skip end doc
	p.nextToken()
	return stmt
}

// parseAnnotation performs business logic tasks as part of parsing of annotations
func (p *Parser) parseAnnotation(gd gendoc.GenDoc, cat NodeCategory, docBlock *GenDocBlock) (gendoc.GenDoc, error) {
	a := attemptIdExtract(gd, docBlock)

	switch cat {
	case ServiceNode:
		if a.Id == "" {
			// fallback on the top level folder name
			// i.e. the result of git clone
			if p.config.ServiceId != "" {
				a.Id = p.config.ServiceId
			} else {
				return a, fmt.Errorf("service annotation parser error, id cannot be deduced: %w", ErrIdRequired)
			}
		}
		a.ServiceRepoLang = p.config.ServiceLanguage
		a.ServiceRepoUrl = p.config.ServiceRepoUrl
		a.ServiceURN = p.serviceUrn(a.Id)
	case ChannelNode:
		err := "channel annotation parse error"
		// must be defined
		if a.Id == "" {
			return a, fmt.Errorf("%s: %w", err, ErrIdRequired)
		}
		if a.Parent == "" && p.config.ServiceId == "" {
			return a, fmt.Errorf("%s: %w", err, ErrParentIdRequired)
		}
		if a.Parent == "" && p.config.ServiceId != "" {
			a.Parent = p.config.ServiceId
		}
		//Operation must specify a parent -> Channel
	case OperationNode:
		err := "operation annotation parse error"
		// must be defined
		if a.Id == "" {
			return a, fmt.Errorf("%s: %w", err, ErrIdRequired)
		}
		if a.Parent == "" && a.ChannelId == "" {
			return a, fmt.Errorf("%s: %w", err, ErrParentIdRequired)
		}
		if a.Parent == "" && a.ChannelId != "" {
			a.Parent = a.ChannelId
		}
	case MessageNode:
		err := "message annotation parse error"
		// must be defined
		if a.Id == "" {
			return a, fmt.Errorf("%s: %w", err, ErrIdRequired)
		}

		if a.Parent == "" {
			// the id of a message and the parent (i.e. an operation must be the same)
			a.Parent = a.Id
		}

		if a.ContentType == "" {
			return a, fmt.Errorf("%s: %w", err, ErrContentTypeRequired)
		}
	}

	// normalize name to be same as Id
	a.Name = a.Id

	return a, nil
}

// serviceUrn sets the id to be AsyncAPI compliant URN
// uses the following format urn:$BUSINESS_DOMAIN:$BOUNDED_CTX_NAME:$SERVICE_ID
func (p *Parser) serviceUrn(id string) string {
	var idTpl = "urn:%s:%s:%s"
	return fmt.Sprintf(idTpl, p.config.BusinessDomain, p.config.BoundedDomain, id)
}

// attemptIdExtract tries to grab the Id from the marker content
//
// It will use this regex `[a-zA-Z0-9~\-|#._]+` to ascertain a valid Id value.
// It will skip any characters not in that group, and it will use the first match.
//
// Example:
//
// # Marker in Terraform
// When you define a list of topics/queues in you can extract its name using the `nameId` type
//
// ```
//
//	 main.tf
//	 ...
//		locals {
//			topics = [
//			  //+gendoc category=channel type=nameId
//			  "domain-demand~demand-cancelled-domain-event",
//			  //-gendoc
//			  "domain-demand~demand-updated-domain-event",
//			   ...
//			]
//
// ...
// ```
//
// The extracted Id will be `domain-demand~demand-cancelled-domain-event`
//
// If an Id is successfully extracted it is assigned to the Id value and it is returned else the object is returned unchanged
func attemptIdExtract(a gendoc.GenDoc, docBlock *GenDocBlock) gendoc.GenDoc {
	if a.ContentType == gendoc.NameId {
		// sanitize
		r := regexp.MustCompile(`[a-zA-Z0-9~\-|#._]+`)
		id := r.FindString(docBlock.Value)
		if id != "" {
			a.Id = id
		}
	}
	return a
}

// ExpandEnvVariables expands the env vars inside DocContent
// to their environment var values.
//
// Failing when a variable is either not set or set but empty.
func ExpandEnvVariables(input string, vars []string) (string, error) {
	for _, v := range vars {
		kv := strings.Split(v, "=")
		key, value := kv[0], kv[1] // kv[1] will be an empty string = ""
		os.Setenv(key, value)
	}

	return envsubst.StringRestrictedNoDigit(input, true, true, false)
}
