package parser

import (
	"encoding/base64"
	"testing"

	"subconv-next/internal/model"
)

func TestDetect(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#node"))
	tests := []struct {
		name    string
		content string
		want    InputKind
	}{
		{name: "uri", content: "vless://uuid@example.com:443#x", want: InputKindURIList},
		{name: "yaml", content: "proxies:\n  - name: demo\n", want: InputKindYAML},
		{name: "base64", content: encoded, want: InputKindBase64},
		{name: "unknown", content: "not-a-subscription", want: InputKindUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Detect([]byte(tt.content)); got != tt.want {
				t.Fatalf("Detect() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseContentVLESSReality(t *testing.T) {
	content := []byte("vless://uuid-1@example.com:443?security=reality&sni=example.com&fp=chrome&pbk=pub&sid=abcd&type=tcp&flow=xtls-rprx-vision#HK Node")
	result := ParseContent(content, model.SourceInfo{Name: "manual", Kind: "inline"})
	if len(result.Errors) != 0 {
		t.Fatalf("Errors = %#v, want none", result.Errors)
	}
	if len(result.Nodes) != 1 {
		t.Fatalf("len(Nodes) = %d, want 1", len(result.Nodes))
	}

	node := result.Nodes[0]
	if node.Type != model.ProtocolVLESS {
		t.Fatalf("Type = %q, want %q", node.Type, model.ProtocolVLESS)
	}
	if node.TLS.Reality == nil || node.TLS.Reality.PublicKey != "pub" {
		t.Fatalf("Reality = %#v, want public key", node.TLS.Reality)
	}
	if flow, ok := node.Raw["flow"].(string); !ok || flow != "xtls-rprx-vision" {
		t.Fatalf("flow raw = %#v, want xtls-rprx-vision", node.Raw["flow"])
	}
	if len(node.Tags) == 0 || node.Tags[0] != "HK" {
		t.Fatalf("Tags = %#v, want HK", node.Tags)
	}
}

func TestParseContentAnyTLSIgnoresReality(t *testing.T) {
	content := []byte("anytls://secret@example.com:443?sni=example.com&client-fingerprint=chrome&pbk=ignored&sid=ignored#anytls")
	result := ParseContent(content, model.SourceInfo{Name: "manual"})
	if len(result.Errors) != 0 {
		t.Fatalf("Errors = %#v, want none", result.Errors)
	}

	node := result.Nodes[0]
	if node.Type != model.ProtocolAnyTLS {
		t.Fatalf("Type = %q, want %q", node.Type, model.ProtocolAnyTLS)
	}
	if len(node.Warnings) != 1 || node.Warnings[0] != "reality parameters ignored for anytls" {
		t.Fatalf("Warnings = %#v, want reality warning", node.Warnings)
	}
	if node.TLS.Reality != nil {
		t.Fatalf("TLS.Reality = %#v, want nil", node.TLS.Reality)
	}
}

func TestParseContentWireGuardURI(t *testing.T) {
	content := []byte("wireguard://private-key@example.com:51820?public-key=server-key&ip=172.16.0.2/32&ipv6=fd00::2/128&allowed-ips=0.0.0.0/0,::/0&reserved=209,98,59&dns=1.1.1.1#wg")
	result := ParseContent(content, model.SourceInfo{Name: "manual"})
	if len(result.Errors) != 0 {
		t.Fatalf("Errors = %#v, want none", result.Errors)
	}

	node := result.Nodes[0]
	if node.WireGuard == nil {
		t.Fatalf("WireGuard = nil")
	}
	if node.Auth.PrivateKey != "private-key" {
		t.Fatalf("PrivateKey = %q, want %q", node.Auth.PrivateKey, "private-key")
	}
	if len(node.WireGuard.Reserved) != 3 {
		t.Fatalf("Reserved = %#v, want 3 values", node.WireGuard.Reserved)
	}
}

func TestParseWireGuardConfig(t *testing.T) {
	content := []byte(`
[Interface]
Address = 172.16.0.2/32, fd00::2/128
PrivateKey = CLIENT_PRIVATE_KEY
DNS = 1.1.1.1
MTU = 1280

[Peer]
PublicKey = SERVER_PUBLIC_KEY
PresharedKey = PSK
AllowedIPs = 0.0.0.0/0, ::/0
Endpoint = example.com:51820
PersistentKeepalive = 25
`)

	node, err := ParseWireGuardConfig(content, model.SourceInfo{Name: "wg"})
	if err != nil {
		t.Fatalf("ParseWireGuardConfig() error = %v", err)
	}
	if node.Type != model.ProtocolWireGuard {
		t.Fatalf("Type = %q, want %q", node.Type, model.ProtocolWireGuard)
	}
	if node.Server != "example.com" || node.Port != 51820 {
		t.Fatalf("endpoint = %s:%d, want example.com:51820", node.Server, node.Port)
	}
	if node.WireGuard == nil || node.WireGuard.IP != "172.16.0.2/32" {
		t.Fatalf("WireGuard = %#v, want IPv4 address", node.WireGuard)
	}
}
