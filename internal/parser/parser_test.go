package parser

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func TestParseContentVLESSXHTTP(t *testing.T) {
	content := []byte("vless://uuid-1@example.com:443?type=xhttp&security=reality&sni=example.com&fp=chrome&pbk=pub&sid=abcd&path=%2Fdemo&mode=auto&no-grpc-header=false#xhttp")
	result := ParseContent(content, model.SourceInfo{Name: "manual"})
	if len(result.Errors) != 0 {
		t.Fatalf("Errors = %#v, want none", result.Errors)
	}

	node := result.Nodes[0]
	if node.Transport.Network != "xhttp" {
		t.Fatalf("Transport.Network = %q, want xhttp", node.Transport.Network)
	}
	if node.Transport.Path != "/demo" {
		t.Fatalf("Transport.Path = %q, want /demo", node.Transport.Path)
	}
	if node.Transport.Mode != "auto" {
		t.Fatalf("Transport.Mode = %q, want auto", node.Transport.Mode)
	}
	if node.Transport.NoGRPCHeader == nil || *node.Transport.NoGRPCHeader != false {
		t.Fatalf("Transport.NoGRPCHeader = %#v, want false pointer", node.Transport.NoGRPCHeader)
	}
	if got, _ := node.Raw["encryption"].(string); got != "none" {
		t.Fatalf("raw encryption = %#v, want none", node.Raw["encryption"])
	}
}

func TestParseContentVMessCipher(t *testing.T) {
	content := []byte("vmess://eyJ2IjoiMiIsInBzIjoidm1lc3Mtbm9kZSIsImFkZCI6ImV4YW1wbGUuY29tIiwicG9ydCI6IjQ0MyIsImlkIjoiMDAwMDAwMDAtMDAwMC0wMDAwLTAwMDAtMDAwMDAwMDAwMDAwIiwiYWlkIjoiMCIsInNjeSI6ImF1dG8iLCJuZXQiOiJ3cyIsInR5cGUiOiJub25lIiwiaG9zdCI6ImNkbi5leGFtcGxlLmNvbSIsInBhdGgiOiIvd3MiLCJ0bHMiOiJ0bHMiLCJzbmkiOiJleGFtcGxlLmNvbSIsImZwIjoiY2hyb21lIn0=")
	result := ParseContent(content, model.SourceInfo{Name: "vmess"})
	if len(result.Errors) != 0 {
		t.Fatalf("Errors = %#v, want none", result.Errors)
	}
	if len(result.Nodes) != 1 {
		t.Fatalf("len(Nodes) = %d, want 1", len(result.Nodes))
	}
	if got := fmt.Sprint(result.Nodes[0].Raw["cipher"]); got != "auto" {
		t.Fatalf("raw cipher = %q, want %q", got, "auto")
	}
}

func TestParseContentSSR(t *testing.T) {
	content := []byte("ssr://ZXhhbXBsZS5jb206NDQzOmF1dGhfc2hhMV92NDphZXMtMjU2LWNmYjp0bHMxLjJfdGlja2V0X2F1dGg6Y0dGemN3Lz9yZW1hcmtzPWMzTnlMVzV2WkdVJm9iZnNwYXJhbT1ZMlJ1TG1WNFlXMXdiR1V1WTI5dCZwcm90b3BhcmFtPQ==")
	result := ParseContent(content, model.SourceInfo{Name: "ssr"})
	if len(result.Errors) != 0 {
		t.Fatalf("Errors = %#v, want none", result.Errors)
	}
	if len(result.Nodes) != 1 {
		t.Fatalf("len(Nodes) = %d, want 1", len(result.Nodes))
	}
	node := result.Nodes[0]
	if node.Type != model.ProtocolSSR {
		t.Fatalf("Type = %q, want %q", node.Type, model.ProtocolSSR)
	}
	if node.Auth.Password != "pass" {
		t.Fatalf("Password = %q, want pass", node.Auth.Password)
	}
	if got := fmt.Sprint(node.Raw["protocol"]); got != "auth_sha1_v4" {
		t.Fatalf("protocol = %q, want auth_sha1_v4", got)
	}
	if got := fmt.Sprint(node.Raw["obfs"]); got != "tls1.2_ticket_auth" {
		t.Fatalf("obfs = %q, want tls1.2_ticket_auth", got)
	}
}

func TestParseProtocolCompatibilityFixture(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "testdata", "subscriptions", "protocols.txt"))
	if err != nil {
		t.Fatalf("ReadFile(protocols.txt) error = %v", err)
	}
	expectedBytes, err := os.ReadFile(filepath.Join("..", "..", "testdata", "expected", "protocols.txt"))
	if err != nil {
		t.Fatalf("ReadFile(expected protocols) error = %v", err)
	}

	result := ParseContent(content, model.SourceInfo{Name: "protocols", Kind: "fixture"})
	if len(result.Errors) != 0 {
		t.Fatalf("Errors = %#v, want none", result.Errors)
	}

	expected := strings.Fields(string(expectedBytes))
	if len(result.Nodes) != len(expected) {
		t.Fatalf("len(Nodes) = %d, want %d", len(result.Nodes), len(expected))
	}
	for index, want := range expected {
		node := result.Nodes[index]
		parts := strings.Split(want, ":")
		if string(node.Type) != parts[0] {
			t.Fatalf("node[%d].Type = %q, want %q", index, node.Type, parts[0])
		}
		if len(parts) > 1 && node.Transport.Network != parts[1] {
			t.Fatalf("node[%d].Transport.Network = %q, want %q", index, node.Transport.Network, parts[1])
		}
		if len(parts) > 2 && parts[2] == "reality" && node.TLS.Reality == nil {
			t.Fatalf("node[%d].TLS.Reality = nil, want reality", index)
		}
	}
}

func TestParseContentMihomoYAML(t *testing.T) {
	content := []byte(`
proxies:
  - name: "[anytls]US LAS Buyvm"
    type: anytls
    server: buyvm-las-01.telecom.moe
    port: 8444
    password: secret-pass
    udp: true
    sni: buyvm-las-01.telecom.moe
    client-fingerprint: chrome
    skip-cert-verify: false
  - name: "[vless]JP Osaka Oracle"
    type: vless
    server: oracle-osa-01.telecom.moe
    port: 5444
    uuid: uuid-1
    network: xhttp
    tls: true
    servername: cas-bridge.xethub.hf.co
    client-fingerprint: chrome
    encryption: none
    reality-opts:
      public-key: pub-key
      short-id: short-id
      spider-x: /
    xhttp-opts:
      path: /NevernessToEverness
      mode: auto
`)
	result := ParseContent(content, model.SourceInfo{Name: "yaml", Kind: "file"})
	if len(result.Errors) != 0 {
		t.Fatalf("Errors = %#v, want none", result.Errors)
	}
	if len(result.Nodes) != 2 {
		t.Fatalf("len(Nodes) = %d, want 2", len(result.Nodes))
	}

	anytls := result.Nodes[0]
	if anytls.Type != model.ProtocolAnyTLS {
		t.Fatalf("first Type = %q, want %q", anytls.Type, model.ProtocolAnyTLS)
	}
	if anytls.Auth.Password != "secret-pass" {
		t.Fatalf("anytls password = %q, want secret-pass", anytls.Auth.Password)
	}
	if anytls.TLS.ClientFingerprint != "chrome" {
		t.Fatalf("anytls client fingerprint = %q, want chrome", anytls.TLS.ClientFingerprint)
	}

	vless := result.Nodes[1]
	if vless.Type != model.ProtocolVLESS {
		t.Fatalf("second Type = %q, want %q", vless.Type, model.ProtocolVLESS)
	}
	if vless.Transport.Network != "xhttp" || vless.Transport.Path != "/NevernessToEverness" || vless.Transport.Mode != "auto" {
		t.Fatalf("vless transport = %#v, want xhttp path/mode", vless.Transport)
	}
	if vless.TLS.Reality == nil || vless.TLS.Reality.PublicKey != "pub-key" {
		t.Fatalf("vless reality = %#v, want parsed reality", vless.TLS.Reality)
	}
	if got := fmt.Sprint(vless.Raw["encryption"]); got != "none" {
		t.Fatalf("vless raw encryption = %q, want none", got)
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
