package generate

// some of these structs were generated from the 2.5.0 async API schema
// NOTE: these should really be generated from the schema, but struggling to find a reliable enough way

// AsyncAPIRoot
type AsyncAPIRoot struct {
	AsyncAPI           string             `json:"asyncapi" yaml:"asyncapi"`
	ID                 string             `json:"id" yaml:"id"`
	Info               Info               `json:"info" yaml:"info"`
	DefaultContentType string             `json:"defaultContentType,omitempty" yaml:"defaultContentType,omitempty"`
	Servers            map[string]Server  `json:"servers,omitempty" yaml:"servers,omitempty"`
	Channels           map[string]Channel `json:"channels" yaml:"channels"`
	Components         *Components        `json:"components,omitempty" yaml:"components,omitempty"`
	Tags               []Tag              `json:"tags,omitempty" yaml:"tags,omitempty"`
}

type Info struct {
	Title          string  `json:"title" yaml:"title"`
	Version        string  `json:"version" yaml:"version"`
	Description    string  `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string  `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        Contact `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        License `json:"license,omitempty" yaml:"license,omitempty"`
}

type Server struct {
	URL         string     `json:"url" yaml:"url"`
	Description string     `json:"description,omitempty" yaml:"description,omitempty"`
	Protocol    string     `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	Username    string     `json:"username,omitempty" yaml:"username,omitempty"`
	Password    string     `json:"password,omitempty" yaml:"password,omitempty"`
	Variables   []Variable `json:"variables,omitempty" yaml:"variables,omitempty"`
}

type Variable struct {
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     string   `json:"default,omitempty" yaml:"default,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
}

type Channel struct {
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Parameters  map[string]Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Publish     *Operation           `json:"publish,omitempty" yaml:"publish,omitempty"` // Channel will be writeable topic or queue or readable subscription or read from queu yaml:"publish,omitempty"` // Channel will be writeable topic or queue or readable subscription or read from queue
	Subscribe   *Operation           `json:"subscribe,omitempty" yaml:"subscribe,omitempty"`
}

type Parameter struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Schema      Schema `json:"schema" yaml:"schema"`
}

type Operation struct {
	Summary     string        `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string        `json:"description,omitempty" yaml:"description,omitempty"`
	OperationId string        `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Traits      []interface{} `json:"traits,omitempty" yaml:"traits,omitempty"`
	Message     *Message       `json:"message,omitempty" yaml:"message,omitempty"`
	Bindings    []interface{} `json:"bindings,omitempty" yaml:"bindings,omitempty"`
}

type CommonDescription struct {
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Summary     string `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type Message struct {
	// MessageBodyShared `json:"inline" yaml:"inline"`
	Name         string                `json:"name,omitempty" yaml:"name,omitempty"`
	Summary      string                `json:"summary,omitempty" yaml:"summary,omitempty"`
	Payload      any                   `json:"payload,omitempty" yaml:"payload,omitempty"`
	Headers      []map[string]Schema   `json:"headers,omitempty" yaml:"headers,omitempty"`
	Title        string                `json:"title,omitempty" yaml:"title,omitempty"`
	Description  string                `json:"description,omitempty" yaml:"description,omitempty"`
	MessageId    string                `json:"messageId" yaml:"messageId"`
	Examples     []MessageBodyShared   `json:"examples,omitempty" yaml:"examples,omitempty"`
	Tags         []Tag                 `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

// Schema was changed to be an interface type - i.e. any
type Schema map[string]interface{}

type Components struct {
	Messages        map[string]Message        `json:"messages,omitempty" yaml:"messages,omitempty"`
	Schemas         map[string]Schema         `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	Parameters      map[string]Parameter      `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

type Tag struct {
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// currently unused
	// ExternalDocs ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

type ExternalDocumentation struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string `json:"url,omitempty" yaml:"url,omitempty"`
}

type License struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url,omitempty" yaml:"url,omitempty"`
}

type Contact struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

type SecurityScheme struct {
	Type         string `json:"type,omitempty" yaml:"type,omitempty"`
	Description  string `json:"description,omitempty" yaml:"description,omitempty"`
	Name         string `json:"name,omitempty" yaml:"name,omitempty"`
	In           string `json:"in,omitempty" yaml:"in,omitempty"`
	Scheme       string `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormat string `json:"bearer" yaml:"bearer"`
}

type MessageBodyShared struct {
	Name    string              `json:"name,omitempty" yaml:"name,omitempty"`
	Summary string              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Payload any                 `json:"payload,omitempty" yaml:"payload,omitempty"`
	Headers []map[string]Schema `json:"headers,omitempty" yaml:"headers,omitempty"`
}
