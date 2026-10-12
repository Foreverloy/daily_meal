package httpapi

import (
	"regexp"

	"dailymeal/backend/internal/service"
	"github.com/gin-gonic/gin"
)

var idempotencyKeyPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,128}$`)

func entryIdempotencyKey(c *gin.Context) (string, bool) {
	values := c.Request.Header.Values("Idempotency-Key")
	if len(values) == 0 {
		return "", true
	}
	if len(values) != 1 || !idempotencyKeyPattern.MatchString(values[0]) {
		writeError(c, &service.Error{
			Code: "bad_request", Message: "请求标识格式错误",
			Fields: map[string]string{"Idempotency-Key": "须为单个 1～128 字符的标识，仅允许字母、数字及 . _ : -"},
		})
		return "", false
	}
	return values[0], true
}
