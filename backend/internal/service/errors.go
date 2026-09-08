package service

type Error struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields"`
}

func (e *Error) Error() string { return e.Message }

func Invalid(field, message string) error {
	return &Error{Code: "validation_error", Message: "业务参数不合法", Fields: map[string]string{field: message}}
}

func NotFound() error {
	return &Error{Code: "not_found", Message: "资源不存在", Fields: map[string]string{}}
}
