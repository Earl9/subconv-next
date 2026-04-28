package parser

import (
	"fmt"
	"strings"

	"subconv-next/internal/model"
)

func parseVLESS(raw string, source model.SourceInfo) (model.NodeIR, error) {
	u, err := parseStandardURL(raw)
	if err != nil {
		return model.NodeIR{}, err
	}

	host, port, err := hostPortFromURL(u)
	if err != nil {
		return model.NodeIR{}, err
	}

	node := newBaseNode(model.ProtocolVLESS, source)
	node.Name = parseFragmentName(u)
	node.Server = host
	node.Port = port
	if u.User != nil {
		node.Auth.UUID = u.User.Username()
	}
	node.UDP = model.Bool(true)

	q := u.Query()
	handled := []string{
		"security", "sni", "servername", "fp", "client-fingerprint",
		"pbk", "publicKey", "public-key", "sid", "shortId", "short-id",
		"spx", "spiderX", "spider-x", "type", "network", "path", "host",
		"serviceName", "service-name", "flow", "packetEncoding", "packet-encoding",
		"alpn", "encryption", "mode", "no-grpc-header", "noGrpcHeader",
	}

	security := strings.ToLower(firstQuery(q, "security"))
	if security == "tls" || security == "reality" {
		node.TLS.Enabled = true
	}
	if sni := firstQuery(q, "sni", "servername"); sni != "" {
		node.TLS.Enabled = true
		node.TLS.SNI = sni
	}
	node.TLS.ClientFingerprint = firstQuery(q, "fp", "client-fingerprint")
	node.TLS.ALPN = parseCSV(firstQuery(q, "alpn"))

	publicKey := firstQuery(q, "pbk", "publicKey", "public-key")
	shortID := firstQuery(q, "sid", "shortId", "short-id")
	spiderX := firstQuery(q, "spx", "spiderX", "spider-x")
	if publicKey != "" || shortID != "" || spiderX != "" {
		node.TLS.Enabled = true
		node.TLS.Reality = &model.RealityOptions{
			PublicKey: publicKey,
			ShortID:   shortID,
			SpiderX:   spiderX,
		}
	}

	node.Transport.Network = firstQuery(q, "type", "network")
	node.Transport.Path = firstQuery(q, "path")
	node.Transport.Host = firstQuery(q, "host")
	node.Transport.ServiceName = firstQuery(q, "serviceName", "service-name")
	node.Transport.Mode = firstQuery(q, "mode")
	node.Transport.H2Hosts = parseCSV(firstQuery(q, "host"))
	if hasQuery(q, "no-grpc-header", "noGrpcHeader") {
		value := parseBoolString(firstQuery(q, "no-grpc-header", "noGrpcHeader"))
		node.Transport.NoGRPCHeader = model.Bool(value)
		setRaw(&node, "noGrpcHeader", value)
	}

	setRaw(&node, "flow", firstQuery(q, "flow"))
	setRaw(&node, "encryption", firstNonEmpty(firstQuery(q, "encryption"), "none"))
	setRaw(&node, "packetEncoding", firstQuery(q, "packetEncoding", "packet-encoding"))
	setRaw(&node, "mode", firstQuery(q, "mode"))

	if raw := unknownQueryParams(q, handled...); raw != nil {
		for key, value := range raw {
			setRaw(&node, key, value)
		}
	}

	return node, nil
}

func parseVisionFlow(flow string) string {
	return strings.TrimSpace(flow)
}

func invalidVLESSError(message string) error {
	return fmt.Errorf("vless: %s", message)
}
