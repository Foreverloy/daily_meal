package openapi

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestDocument(t *testing.T) {
	document, err := openapi3.NewLoader().LoadFromData(Document)
	if err != nil {
		t.Fatal(err)
	}
	if err := document.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
}
