package openapi

import _ "embed"

// Document 嵌入手工维护的接口契约；接口变更时需同步 openapi.json。
//
//go:embed openapi.json
var Document []byte
