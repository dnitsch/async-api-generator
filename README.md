# async-api-generator

This tool parses source code - of any type and looks for tags / annotations to extract. 

_Limitations by design:_

- requires a closing tag as the source can be of any type and formatting is not guaranteed. Essentially using things like `{` or cursor column indicator, wouldn't be reliable enough

## Flow Overview

walk the directory and create a list of all source files - excluding specified such as bin, dist, vendor, node_modules, etc...

running each file as a source through a lexer i.e. performing a lexical analysis, where we identify tokens and build an AST (Abstract Syntax Tree)

> each file can be done in a separate go routine as there is no intrinsic relationship between them at this stage

### Lexer

TOKENs are kept to the following types - `token.TokenType`  - not listed as the list is likely to change/grow. 

Special consideration will need to be given to files that are not able to contain comments or anything outside of their existing syntax - e.g. .json most commonly containing schemas. 

> these cases a convention will need to be followed where by the name of the message that it is describing must be in the name of the file.

### Parser

Not using an existing parser generator like CGF, BNF, or EBNF is on purposes as the input source will only ever really be composed of parts we care about i.e. `gendoc` markers their beginning and end and what they enclose within them as text.

We'll use the overly simplified Pratt Parser (top down method) as we'll have no need to for expression parsing onyl statement node creation with the associated/helper attributes.

## AsyncAPI standard

The current [AsyncAPI standard spec](https://www.asyncapi.com/docs/reference/specification/v2.6.0) is at version `2.6.0`.

The tool will deal with all the relevant sections to be able to build an AsyncAPI spec file from within a single repo.

The asyncAPI is built from the `Application` - i.e. service down, each service will have a toplevel description - `info` key, which will in turn include 

### Important Properties

- `id` name of the service. Will default to parent folder name - unless overridden
    - format:  `urn:$business_domain:$bounded_context_domain:$service_name` => `urn:whds:packing:whds.packing.app`
- `application` 
    this is info about the service/application including descriptions and titles
- `channels` outlines the topics/subscriptions or queues the application produces or is subscribed to.
    - `topic/queue/subscription`
        will each contain a message summary description, schema, any traits - i.e. re-useable components - such as the envelop for common parameters

### EventCatalog binding

The translation of AsyncAPI into an EventCatalog set up. Whilst there are fairly standard mappings between the 2 processes - there are some nuances and requirements.
