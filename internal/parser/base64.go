package parser

import (
	"encoding/base64"
	"fmt"
	"strings"
)

func DecodeBase64Bytes(data []byte) ([]byte, error) {
	return DecodeBase64String(string(data))
}

func DecodeBase64String(value string) ([]byte, error) {
	compact := strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\n', '\r', '\t':
			return -1
		default:
			return r
		}
	}, strings.TrimSpace(value))
	if compact == "" {
		return nil, fmt.Errorf("empty base64 input")
	}

	candidates := []string{compact}
	switch len(compact) % 4 {
	case 2:
		candidates = append(candidates, compact+"==")
	case 3:
		candidates = append(candidates, compact+"=")
	}

	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}

	var lastErr error
	for _, candidate := range candidates {
		for _, enc := range encodings {
			decoded, err := enc.DecodeString(candidate)
			if err == nil {
				return decoded, nil
			}
			lastErr = err
		}
	}

	return nil, fmt.Errorf("decode base64: %w", lastErr)
}
