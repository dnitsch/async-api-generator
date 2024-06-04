// Package tools only imports locally any required build/test tools
// will not be part of the final binary
//
// The main idea is to autogenerate the models from the schema however
// as the schema is quite complex there aren't any tools at the moment.
package tools

// import (
// 	_ "github.com/atombender/go-jsonschema/cmd/gojsonschema"
// 	_ "github.com/deepmap/oapi-codegen/cmd/oapi-codegen"
// )

// Testing of go:generate

//go:generate gojsonschema --schema-package=https://asyncapi.com/definitions/2.5.0/asyncapi.json=github.com/dnitsch/async-api-generator/schema --schema-output=https://asyncapi.com/definitions/2.5.0/asyncapi.json=github.com/dnitsch/async-api-generator/schema/schema.go ./asyncapi_schema_2.5.0_.json

//go:generate oapi-codegen --package=tools -generate=types -o ./petstore.gen.go ./oas.asyncapi_2.5.0.json

//go:generate gojsonschema --package=generate -generate=types -o ../pkg/generate/gen.go https://asyncapi.com/definitions/2.5.0/asyncapi.json

// ./oas.asyncapi_2.5.0.json
