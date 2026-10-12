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
		{"POST", "/api/v1/foods", `{"favorite":true}`, 400},
		{"POST", "/api/v1/meal-entries", `{"request_id":"unexpected"}`, 400},
		{"PATCH", "/api/v1/foods/1", `{"favorite":"true"}`, 400},
		{"GET", "/api/v1/foods/nope", ``, 422},
		{"GET", "/api/v1/foods/0", ``, 422},
		{"GET", "/api/v1/foods?page=invalid", ``, 422},
		{"GET", "/api/v1/foods?favorite=invalid", ``, 422},
		{"GET", "/api/v1/foods?favorite=", ``, 422},
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

func TestEntryIdempotencyHeaderErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := New(nil, nil, openapi.Document)
	for _, tc := range []struct {
		name   string
		values []string
	}{
		{"empty", []string{""}},
		{"too long", []string{strings.Repeat("a", 129)}},
		{"whitespace", []string{"key with space"}},
		{"non ASCII", []string{"午餐"}},
		{"multiple headers", []string{"key-1", "key-2"}},
		{"comma separated", []string{"key-1,key-2"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/api/v1/meal-entries", strings.NewReader(`{}`))
			request.Header["Idempotency-Key"] = tc.values
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)
			var envelope struct {
				Error struct {
					Code   string
					Fields map[string]string
				}
			}
			if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
			if w.Code != 400 || envelope.Error.Code != "bad_request" || envelope.Error.Fields["Idempotency-Key"] == "" {
				t.Fatalf("invalid key response=%d %s", w.Code, w.Body)
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
