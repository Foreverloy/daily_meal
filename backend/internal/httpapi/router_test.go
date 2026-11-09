package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dailymeal/backend/openapi"
	"github.com/gin-gonic/gin"
)

// TestTransportErrors 验证请求解析、路由和健康检查失败时返回的 HTTP 状态码。
func TestTransportErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := New(nil, func(context.Context) error { return errors.New("postgres secret database details") }, openapi.Document)
	for _, tc := range []struct {
		// method、path、body 分别是请求方法、接口路径和 JSON 请求体文本。
		method, path, body string
		// status 是预期返回的 HTTP 状态码，如 400 表示请求解析失败。
		status int
	}{
		{"POST", "/api/v1/foods", `[]`, 400},
		{"POST", "/api/v1/foods", `{`, 400},
		{"POST", "/api/v1/foods", `{} {}`, 400},
		{"POST", "/api/v1/foods", `{"name":123}`, 400},
		{"PATCH", "/api/v1/foods/1", `{"favorite":"true"}`, 400},
		{"GET", "/api/v1/foods/nope", ``, 400},
		{"GET", "/api/v1/foods?page=invalid", ``, 400},
		{"GET", "/api/v1/foods?favorite=invalid", ``, 400},
		{"GET", "/api/v1/foods?favorite=", ``, 400},
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

func TestResponseEncodingFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/result", func(c *gin.Context) {
		respond(c, http.StatusOK, gin.H{"total": math.Inf(1)}, nil)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/result", nil))
	if w.Code != http.StatusInternalServerError || !json.Valid(w.Body.Bytes()) {
		t.Fatalf("encoding failure returned success or invalid JSON: %d %s", w.Code, w.Body)
	}
}
