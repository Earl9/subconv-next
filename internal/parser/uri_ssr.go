package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"subconv-next/internal/model"
)

func parseSSR(raw string, source model.SourceInfo) (model.NodeIR, error) {
	body := strings.TrimPrefix(raw, "ssr://")
	if idx := strings.Index(body, "#"); idx >= 0 {
		body = body[:idx]
	}
	decoded, err := DecodeBase64String(body)
	if err != nil {
		return model.NodeIR{}, fmt.Errorf("decode ssr payload: %w", err)
	}

	payload := string(decoded)
	mainPart, queryPart := splitSSRPayload(payload)
	mainPart = strings.TrimSuffix(mainPart, "/")
	parts := strings.SplitN(mainPart, ":", 6)
	if len(parts) != 6 {
		return model.NodeIR{}, fmt.Errorf("invalid ssr payload")
	}

	port, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return model.NodeIR{}, fmt.Errorf("invalid ssr port: %w", err)
	}
	password, err := decodeSSRParam(parts[5])
	if err != nil {
		return model.NodeIR{}, fmt.Errorf("decode ssr password: %w", err)
	}

	values, _ := url.ParseQuery(queryPart)
	node := newBaseNode(model.ProtocolSSR, source)
	node.Server = strings.TrimSpace(parts[0])
	node.Port = port
	node.Auth.Password = password
	node.Name = firstNonEmpty(decodeSSRParamOrEmpty(values.Get("remarks")), node.Server)
	node.UDP = model.Bool(true)
	setRaw(&node, "protocol", strings.TrimSpace(parts[2]))
	setRaw(&node, "method", strings.TrimSpace(parts[3]))
	setRaw(&node, "obfs", strings.TrimSpace(parts[4]))
	setRaw(&node, "protocolParam", decodeSSRParamOrEmpty(values.Get("protoparam")))
	setRaw(&node, "obfsParam", decodeSSRParamOrEmpty(values.Get("obfsparam")))
	return node, nil
}

func splitSSRPayload(payload string) (string, string) {
	if parts := strings.SplitN(payload, "/?", 2); len(parts) == 2 {
		return parts[0], parts[1]
	}
	if parts := strings.SplitN(payload, "?", 2); len(parts) == 2 {
		return strings.TrimSuffix(parts[0], "/"), parts[1]
	}
	return payload, ""
}

func decodeSSRParam(value string) (string, error) {
	value = strings.Trim(strings.TrimSpace(value), "/")
	if value == "" {
		return "", nil
	}
	decoded, err := DecodeBase64String(value)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func decodeSSRParamOrEmpty(value string) string {
	decoded, err := decodeSSRParam(value)
	if err != nil {
		return ""
	}
	return decoded
}
