package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dailymeal/backend/openapi"
	"github.com/gin-gonic/gin"
)

func TestTransportErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := New(nil, func(context.Context) error { return errors.New("postgres secret database details") }, openapi.Document)
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{
		{"POST", "/api/v1/foods", `null`, 400},
		{"POST", "/api/v1/foods", `[]`, 400},
		{"POST", "/api/v1/foods", `{`, 400},
		{"POST", "/api/v1/foods", `{} {}`, 400},
		{"POST", "/api/v1/foods", `{"name":123}`, 400},
		{"POST", "/api/v1/foods", `{"unknown":true}`, 400},
		{"PATCH", "/api/v1/foods/1", `{"basis_unit":"ml"}`, 400},
		{"GET", "/api/v1/foods/nope", ``, 422},
		{"GET", "/api/v1/foods/0", ``, 422},
		{"GET", "/api/v1/foods?page=invalid", ``, 422},
		{"GET", "/missing", ``, 404},
		{"GET", "/healthz", ``, 500},
	} {
		t.Run(tc.method+tc.path+tc.body, func(t *testing.T) {
			response := httptest.NewRecorder()
			r.ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
			if response.Code != tc.status {
				t.Fatalf("got %d: %s", response.Code, response.Body)
			}
			var envelope map[string]json.RawMessage
			if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil || envelope["error"] == nil {
				t.Fatalf("not an error envelope: %s", response.Body)
			}
			if strings.Contains(response.Body.String(), "secret") {
				t.Fatal("leaked database details")
			}
		})
	}
}

func TestHealthAndOpenAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := New(nil, func(context.Context) error { return nil }, openapi.Document)
	for _, path := range []string{"/healthz", "/openapi.json"} {
		response := httptest.NewRecorder()
		r.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != 200 || !json.Valid(response.Body.Bytes()) {
			t.Fatalf("%s: %d %s", path, response.Code, response.Body)
		}
	}
}
