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
