package parser

import (
	"bytes"
	"strings"
)

type InputKind string

const (
	InputKindURIList InputKind = "uri_list"
	InputKindBase64  InputKind = "base64"
	InputKindYAML    InputKind = "yaml"
	InputKindUnknown InputKind = "unknown"
)

func Detect(content []byte) InputKind {
	trimmed := bytes.TrimSpace(content)
	if len(trimmed) == 0 {
		return InputKindUnknown
	}

	lower := strings.ToLower(string(trimmed))
	if strings.Contains(lower, "proxies:") {
		return InputKindYAML
	}
	if strings.Contains(lower, "://") {
		return InputKindURIList
	}

	if decoded, err := DecodeBase64Bytes(trimmed); err == nil {
		decodedKind := Detect(decoded)
		if decodedKind != InputKindUnknown {
			return InputKindBase64
		}
	}

	return InputKindUnknown
}
