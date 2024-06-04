// Package gendoc holds the struct for gendoc annotations
//
// It uses an in-built precence to identify parent to current
// This can be used to generate and store an interim state before the AsyncAPI structure is
package gendoc

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dnitsch/async-api-generator/internal/token"
	log "github.com/dnitsch/simplelog"
)

// ContentType is an indicator of where to place the extracted text in the template
//
// i.e. is it a Description, summary, title, example?
type ContentType string

const (
	JSONSchema  ContentType = "json_schema" // when a schema is embedded
	Example     ContentType = "example"
	Title       ContentType = "title"       // human friendly title of an object will be used in title - e.g. in message or operation
	Summary     ContentType = "summary"     // short description of an object will be used in summary where possible - e.g. message/operation
	Description ContentType = "description" // long description of the object will be used in description
	NameId      ContentType = "nameId"
)

var contentTypeEnum = map[string]ContentType{
	"json_schema": JSONSchema,
	"example":     Example,
	"description": Description,
	"title":       Title,
	"summary":     Summary,
	"nameId":      NameId,
}

// CategoryType is the top level categery for the annotation
// e.g. a server or channel or operation
// will guide as to how the rest of the annotation is unpacked.
//
// Server and Info Blocks are fairly flat in that they only go one level deep.
//
// Channel however consists of parameters and publish or subscribe operation block
// each operation will have a message block
//
// Message Block will have the message definitions and will be bound
// to a payload/schema and examples/samples
type CategoryType string

const (
	RootBlock         CategoryType = "root"
	InfoBlock         CategoryType = "info"
	ServerBlock       CategoryType = "server"
	ChannelBlock      CategoryType = "channel"
	OperationBlock    CategoryType = "operation"
	SubOperationBlock CategoryType = "subOperation"
	PubOperationBlock CategoryType = "pubOperation"
	MessageBlock      CategoryType = "message"
)

var categoryTypeEnum = map[string]CategoryType{
	"server":       ServerBlock,
	"info":         InfoBlock,
	"channel":      ChannelBlock,
	"operation":    OperationBlock,
	"subOperation": SubOperationBlock,
	"pubOperation": PubOperationBlock,
	"message":      MessageBlock,
	"root":         RootBlock,
}

// GenDoc holds all the attributes required to backfill an AsyncAPI
//
// These are user facing properties and should be mapped
// directly to the AsyncAPI standard specifications
type GenDoc struct {
	raw             string
	log             log.Loggeriface
	CategoryType    CategoryType `json:"category" yaml:"category"`
	ContentType     ContentType  `json:"type" yaml:"type"`
	Name            string       `json:"name" yaml:"name"`
	Id              string       `json:"id" yaml:"id"`
	ServiceId       string       `json:"serviceId" yaml:"serviceId"` // for cases when children of services i.e. channels or servers are defined outside of a repo. This property can be in form of an array like string.
	ChannelId       string       `json:"channelId" yaml:"channelId"` // for cases when children of channels i.e. operations are defined outside of a repo. This property can be in form of an array like string.
	Parent          string       `json:"parent" yaml:"parent"`
	ServiceURN      string       `json:"serviceURN" yaml:"serviceURN"`
	ServiceRepoUrl  string       `json:"serviceRepoUrl" yaml:"serviceRepoUrl"`
	ServiceRepoLang string       `json:"serviceRepoLang" yaml:"serviceRepoLang"`
}

func NewFromToken(token token.Token, log log.Loggeriface) (GenDoc, error) {
	g := &GenDoc{raw: token.MetaAnnotation, log: log}
	return g.new()
}

// New returns the GenDoc value
//
// It should ONLY be extended by caller.
func New(annotation string, log log.Loggeriface) (GenDoc, error) {
	l := &log
	g := &GenDoc{raw: annotation, log: *l}
	return g.new()
}

func (g *GenDoc) new() (GenDoc, error) {
	if err := g.unmarshal(); err != nil {
		return *g, err
	}
	return *g, nil
}

func wrapErr(msg string, etyp error) error {
	return fmt.Errorf("%s\n%w", msg, etyp)
}

var (
	// ErrUnparseableTag indicates that the string extracted as a possible key/value cannot be parsed as such.
	ErrUnparseableTag = errors.New("string field cannot be split into a key=value")
	// ErrZeroLengthKeyOrValue means that either the key or the value has 0 length.
	ErrZeroLengthKeyOrValue = errors.New("both key and value must be a non-zero length string")
	// ErrIncorrectCategory indicates that an unknown category has been chosen.
	ErrIncorrectCategory = errors.New("category type incorrect should be one of ['server','info','channel','operation','message','root']")
	// ErrIncorrectType means that wrong type has been specified.
	ErrIncorrectType = errors.New("content type incorrect should be one of ['json_schema','example','mdescription','title','nameId']")
)

func (g *GenDoc) unmarshal() error {
	fields := strings.Fields(string(g.raw))
	for _, keyvalpair := range fields {
		if ignore(keyvalpair) {
			continue
		}
		splitv := strings.Split(keyvalpair, "=")
		if len(splitv) != 2 {
			g.log.Debugf("value in gendoc tag: %s, cannot be split into a key=value", keyvalpair)
			return wrapErr(keyvalpair, ErrUnparseableTag)
		}
		key, val := splitv[0], splitv[1]
		if len(key) < 1 || len(val) < 1 {
			return wrapErr(fmt.Sprintf("key '%s' and value '%s'", key, val), ErrZeroLengthKeyOrValue)
		}
		switch key {
		case "id":
			g.Id = val
		case "parent", "p":
			g.Parent = val
		case "serviceId":
			g.ServiceId = val
		case "channelId":
			g.ChannelId = val
		case "type":
			found, ok := contentTypeEnum[val]
			if !ok {
				return wrapErr(fmt.Sprintf("type: '%s'", val), ErrIncorrectType)
			}
			g.ContentType = found
		case "category", "cat", "c":
			found, ok := categoryTypeEnum[val]
			if !ok {
				return wrapErr(fmt.Sprintf("category: '%s'", val), ErrIncorrectCategory)
			}
			g.CategoryType = found
		default:
			g.log.Debugf("the tag key=value pair '%s' is in correct format, unable to match the key '%s' to an existing case", keyvalpair, key)
			g.log.Debug("skipping...")
		}
	}
	return nil
}

var commentBlock = map[string]bool{
	"-->": true,
	"#":   true,
	"##":  true,
}

func ignore(keyval string) bool {
	if _, ok := commentBlock[keyval]; ok {
		return true
	}
	return false
}
