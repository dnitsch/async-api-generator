package main

import (
	"context"

	asyncapigendoc "github.com/dnitsch/async-api-generator/cmd/async-api-gen-doc"
)

func main() {
	asyncapigendoc.Execute(context.Background())
}
