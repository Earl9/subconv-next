package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"subconv-next/internal/model"
)

func parseVMess(raw string, source model.SourceInfo) (model.NodeIR, error) {
	body := strings.TrimPrefix(raw, "vmess://")
	decoded, err := DecodeBase64String(body)
	if err != nil {
		return model.NodeIR{}, fmt.Errorf("decode vmess payload: %w", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return model.NodeIR{}, fmt.Errorf("decode vmess json: %w", err)
	}

	node := newBaseNode(model.ProtocolVMess, source)
	node.Name = stringField(payload, "ps")
	node.Server = stringField(payload, "add")
	if port, ok := parseIntString(stringField(payload, "port")); ok {
		node.Port = port
	}
	node.Auth.UUID = stringField(payload, "id")
	node.Transport.Network = stringField(payload, "net")
	node.Transport.Host = stringField(payload, "host")
	node.Transport.Path = stringField(payload, "path")
	node.Transport.H2Hosts = parseCSV(stringField(payload, "host"))
	node.TLS.SNI = stringField(payload, "sni")
	node.TLS.ALPN = parseCSV(stringField(payload, "alpn"))
	node.TLS.ClientFingerprint = stringField(payload, "fp")
	node.UDP = model.Bool(true)

	if tlsValue := strings.ToLower(stringField(payload, "tls")); tlsValue != "" && tlsValue != "none" && tlsValue != "0" && tlsValue != "false" {
		node.TLS.Enabled = true
	}

	setRaw(&node, "alterId", stringField(payload, "aid"))
	setRaw(&node, "headerType", stringField(payload, "type"))
	setRaw(&node, "cipher", firstNonEmpty(stringField(payload, "scy"), stringField(payload, "cipher")))
	return node, nil
}

func stringField(values map[string]interface{}, key string) string {
	value, ok := values[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return fmt.Sprintf("%.0f", typed)
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}
