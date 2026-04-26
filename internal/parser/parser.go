package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"subconv-next/internal/model"
)

type ParseResult struct {
	Nodes    []model.NodeIR `json:"nodes"`
	Warnings []string       `json:"warnings"`
	Errors   []ParseError   `json:"errors"`
}

type ParseError struct {
	Line    int    `json:"line"`
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

func ParseContent(content []byte, source model.SourceInfo) ParseResult {
	return parseByKind(Detect(content), content, source, 0)
}

func parseByKind(kind InputKind, content []byte, source model.SourceInfo, depth int) ParseResult {
	if depth > 1 {
		return ParseResult{
			Errors: []ParseError{{Kind: "PARSE_DEPTH", Message: "nested content detection exceeded"}},
		}
	}

	switch kind {
	case InputKindURIList:
		return parseURIList(content, source)
	case InputKindBase64:
		decoded, err := DecodeBase64Bytes(content)
		if err != nil {
			return ParseResult{
				Errors: []ParseError{{Kind: "INVALID_BASE64", Message: err.Error()}},
			}
		}
		result := parseByKind(Detect(decoded), decoded, source, depth+1)
		result.Warnings = append([]string{"base64 decoded input"}, result.Warnings...)
		return result
	case InputKindYAML:
		return ParseResult{
			Errors: []ParseError{{Kind: "UNSUPPORTED_YAML", Message: "YAML proxy parsing is not implemented yet"}},
		}
	default:
		node, err := ParseWireGuardConfig(content, source)
		if err != nil {
			return ParseResult{
				Errors: []ParseError{{Kind: "UNKNOWN_INPUT", Message: "unsupported subscription content"}},
			}
		}
		return ParseResult{Nodes: model.NormalizeNodes([]model.NodeIR{node})}
	}
}

func parseURIList(content []byte, source model.SourceInfo) ParseResult {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	result := ParseResult{}
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		node, err := parseURI(line, source)
		if err != nil {
			result.Errors = append(result.Errors, ParseError{
				Line:    lineNo,
				Kind:    "INVALID_URI",
				Message: err.Error(),
			})
			continue
		}

		result.Nodes = append(result.Nodes, node)
	}

	if err := scanner.Err(); err != nil {
		result.Errors = append(result.Errors, ParseError{
			Kind:    "SCAN_ERROR",
			Message: err.Error(),
		})
	}

	result.Nodes = model.NormalizeNodes(result.Nodes)
	return result
}

func parseURI(raw string, source model.SourceInfo) (model.NodeIR, error) {
	schemeEnd := strings.Index(raw, "://")
	if schemeEnd <= 0 {
		return model.NodeIR{}, fmt.Errorf("missing scheme")
	}

	switch strings.ToLower(raw[:schemeEnd]) {
	case string(model.ProtocolSS):
		return parseSS(raw, source)
	case string(model.ProtocolVMess):
		return parseVMess(raw, source)
	case string(model.ProtocolVLESS):
		return parseVLESS(raw, source)
	case string(model.ProtocolTrojan):
		return parseTrojan(raw, source)
	case "hy2", string(model.ProtocolHysteria2):
		return parseHysteria2(raw, source)
	case string(model.ProtocolTUIC):
		return parseTUIC(raw, source)
	case string(model.ProtocolAnyTLS):
		return parseAnyTLS(raw, source)
	case string(model.ProtocolWireGuard):
		return parseWireGuardURI(raw, source)
	default:
		return model.NodeIR{}, fmt.Errorf("unsupported scheme %q", raw[:schemeEnd])
	}
}

func newBaseNode(protocol model.Protocol, source model.SourceInfo) model.NodeIR {
	return model.NodeIR{
		Type:   protocol,
		Source: source,
	}
}

func parseStandardURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}
	if u.Hostname() == "" {
		return nil, fmt.Errorf("missing host")
	}
	return u, nil
}

func hostPortFromURL(u *url.URL) (string, int, error) {
	host := u.Hostname()
	if host == "" {
		return "", 0, fmt.Errorf("missing host")
	}

	port := 0
	if rawPort := u.Port(); rawPort != "" {
		parsedPort, err := strconv.Atoi(rawPort)
		if err != nil {
			return "", 0, fmt.Errorf("invalid port %q", rawPort)
		}
		port = parsedPort
	}

	return host, port, nil
}

func parseFragmentName(u *url.URL) string {
	if u == nil || u.Fragment == "" {
		return ""
	}

	name, err := url.PathUnescape(u.Fragment)
	if err != nil {
		return u.Fragment
	}
	return name
}

func parseBoolString(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func parseIntString(value string) (int, bool) {
	if strings.TrimSpace(value) == "" {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, false
	}
	return n, true
}

func parseCSV(value string) []string {
	parts := strings.Split(value, ",")
	var out []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func firstQuery(values url.Values, keys ...string) string {
	for _, key := range keys {
		for existing, existingValues := range values {
			if strings.EqualFold(existing, key) && len(existingValues) > 0 {
				return existingValues[0]
			}
		}
	}
	return ""
}

func unknownQueryParams(values url.Values, handled ...string) map[string]interface{} {
	handledSet := make(map[string]struct{}, len(handled))
	for _, key := range handled {
		handledSet[strings.ToLower(key)] = struct{}{}
	}

	raw := make(map[string]interface{})
	for key, vals := range values {
		if _, ok := handledSet[strings.ToLower(key)]; ok {
			continue
		}
		if len(vals) == 1 {
			raw[key] = vals[0]
			continue
		}
		copied := append([]string(nil), vals...)
		raw[key] = copied
	}

	if len(raw) == 0 {
		return nil
	}
	return raw
}

func setRaw(node *model.NodeIR, key string, value interface{}) {
	if value == nil {
		return
	}
	switch typed := value.(type) {
	case string:
		if strings.TrimSpace(typed) == "" {
			return
		}
	case []string:
		if len(typed) == 0 {
			return
		}
	}

	if node.Raw == nil {
		node.Raw = make(map[string]interface{})
	}
	node.Raw[key] = value
}

func splitEndpoint(value string) (string, int, error) {
	host, portString, err := net.SplitHostPort(strings.TrimSpace(value))
	if err != nil {
		return "", 0, fmt.Errorf("parse endpoint %q: %w", value, err)
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		return "", 0, fmt.Errorf("invalid endpoint port %q", portString)
	}

	return host, port, nil
}
