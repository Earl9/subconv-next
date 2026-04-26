package parser

import (
	"fmt"
	"net/url"
	"strings"

	"subconv-next/internal/model"
)

func parseSS(raw string, source model.SourceInfo) (model.NodeIR, error) {
	body := strings.TrimPrefix(raw, "ss://")
	name := ""
	if idx := strings.Index(body, "#"); idx >= 0 {
		name = body[idx+1:]
		body = body[:idx]
	}

	if !strings.Contains(body, "@") {
		decoded, err := DecodeBase64String(body)
		if err != nil {
			return model.NodeIR{}, fmt.Errorf("decode ss payload: %w", err)
		}
		raw = "ss://" + string(decoded)
		if name != "" {
			raw += "#" + name
		}
	}

	u, err := parseStandardURL(raw)
	if err != nil {
		return model.NodeIR{}, err
	}

	host, port, err := hostPortFromURL(u)
	if err != nil {
		return model.NodeIR{}, err
	}
	if u.User == nil {
		return model.NodeIR{}, fmt.Errorf("missing ss userinfo")
	}

	method := u.User.Username()
	password, hasPassword := u.User.Password()
	if !hasPassword {
		decoded, err := DecodeBase64String(method)
		if err != nil {
			return model.NodeIR{}, fmt.Errorf("decode ss credential: %w", err)
		}
		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) != 2 {
			return model.NodeIR{}, fmt.Errorf("invalid ss credential")
		}
		method = parts[0]
		password = parts[1]
	}

	node := newBaseNode(model.ProtocolSS, source)
	node.Name = parseFragmentName(u)
	node.Server = host
	node.Port = port
	node.Auth.Password = password
	node.UDP = model.Bool(true)
	setRaw(&node, "method", method)

	return node, nil
}

func escapeSSUserInfo(method, password string) string {
	return url.UserPassword(method, password).String()
}
