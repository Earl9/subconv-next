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

func TestParseContentMieruURI(t *testing.T) {
	content := []byte("mieru://user:secret@example.com:2999?transport=tcp&udp=1&multiplexing=MULTIPLEXING_LOW&handshake-mode=HANDSHAKE_STANDARD#mieru-node")
	result := ParseContent(content, model.SourceInfo{Name: "manual"})
	if len(result.Errors) != 0 {
		t.Fatalf("Errors = %#v, want none", result.Errors)
	}
	if len(result.Nodes) != 1 {
		t.Fatalf("len(Nodes) = %d, want 1", len(result.Nodes))
	}

	node := result.Nodes[0]
	if node.Type != model.ProtocolMieru {
		t.Fatalf("Type = %q, want %q", node.Type, model.ProtocolMieru)
	}
	if node.Auth.Username != "user" || node.Auth.Password != "secret" {
		t.Fatalf("auth = %#v, want username/password", node.Auth)
	}
	if got := fmt.Sprint(node.Raw["transport"]); got != "TCP" {
		t.Fatalf("transport = %q, want TCP", got)
	}
	if node.UDP == nil || !*node.UDP {
		t.Fatalf("UDP = %#v, want true", node.UDP)
	}
}

func TestParseContentMieruPortRange(t *testing.T) {
	content := []byte("mieru://example.com?username=user&password=secret&port-range=2090-2099&transport=UDP#mieru-range")
	result := ParseContent(content, model.SourceInfo{Name: "manual"})
	if len(result.Errors) != 0 {
		t.Fatalf("Errors = %#v, want none", result.Errors)
	}

	node := result.Nodes[0]
	if node.Port != 0 {
		t.Fatalf("Port = %d, want 0 for port-range", node.Port)
	}
	if got := fmt.Sprint(node.Raw["portRange"]); got != "2090-2099" {
		t.Fatalf("portRange = %q, want 2090-2099", got)
	}
}

func TestParseContentHTTPAndSOCKS5URI(t *testing.T) {
	content := []byte(strings.Join([]string{
		"http://user:secret@example.com:8080#http-node",
		"https://secure:pass@example.org:8443?sni=proxy.example.org&skip-cert-verify=1#https-node",
		"socks5://sock:s3cr3t@socks.example.net:1080?udp=0#socks-node",
	}, "\n"))
	result := ParseContent(content, model.SourceInfo{Name: "manual"})
	if len(result.Errors) != 0 {
		t.Fatalf("Errors = %#v, want none", result.Errors)
	}
	if len(result.Nodes) != 3 {
		t.Fatalf("len(Nodes) = %d, want 3", len(result.Nodes))
	}

	httpNode := result.Nodes[0]
	if httpNode.Type != model.ProtocolHTTP || httpNode.Auth.Username != "user" || httpNode.Auth.Password != "secret" {
		t.Fatalf("http node = %#v, want http with username/password", httpNode)
	}
	if httpNode.TLS.Enabled {
		t.Fatalf("http TLS.Enabled = true, want false")
	}

	httpsNode := result.Nodes[1]
	if httpsNode.Type != model.ProtocolHTTP || !httpsNode.TLS.Enabled || !httpsNode.TLS.Insecure || httpsNode.TLS.SNI != "proxy.example.org" {
		t.Fatalf("https node = %#v, want http type with TLS/SNI/insecure", httpsNode)
	}

	socksNode := result.Nodes[2]
	if socksNode.Type != model.ProtocolSOCKS5 || socksNode.UDP == nil || *socksNode.UDP {
		t.Fatalf("socks5 node = %#v, want socks5 with udp false", socksNode)
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

func TestParseContentVLESSPacketEncodingAlias(t *testing.T) {
	content := []byte("vless://uuid-1@example.com:443?type=tcp&security=tls&sni=example.com&packet_encoding=XUDP#xudp")
	result := ParseContent(content, model.SourceInfo{Name: "manual"})
	if len(result.Errors) != 0 {
		t.Fatalf("Errors = %#v, want none", result.Errors)
	}
	if len(result.Nodes) != 1 {
		t.Fatalf("len(Nodes) = %d, want 1", len(result.Nodes))
	}

	node := result.Nodes[0]
	if got, _ := node.Raw["packetEncoding"].(string); got != "xudp" {
		t.Fatalf("raw packetEncoding = %#v, want xudp", node.Raw["packetEncoding"])
	}
	if _, ok := node.Raw["packet_encoding"]; ok {
		t.Fatalf("raw packet_encoding should be normalized away: %#v", node.Raw)
	}
}

func TestParseContentVLESSInsecure(t *testing.T) {
	content := []byte("vless://uuid-1@example.com:443?type=tcp&security=tls&sni=example.com&allowInsecure=1#insecure")
	result := ParseContent(content, model.SourceInfo{Name: "manual"})
	if len(result.Errors) != 0 {
		t.Fatalf("Errors = %#v, want none", result.Errors)
	}
	if len(result.Nodes) != 1 {
		t.Fatalf("len(Nodes) = %d, want 1", len(result.Nodes))
	}
	if !result.Nodes[0].TLS.Insecure {
		t.Fatalf("TLS.Insecure = false, want true")
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
    skip-cert-verify: true
    encryption: none
    reality-opts:
      public-key: pub-key
      short-id: short-id
      spider-x: /
    xhttp-opts:
      path: /NevernessToEverness
      mode: auto
  - name: "mieru yaml"
    type: mieru
    server: mieru.example.com
    port-range: 2090-2099
    transport: TCP
    udp: true
    username: user
    password: secret
    multiplexing: MULTIPLEXING_LOW
  - name: "http yaml"
    type: http
    server: http.example.com
    port: 8080
    username: http-user
    password: http-pass
    tls: true
    sni: http.example.com
    skip-cert-verify: true
  - name: "socks yaml"
    type: socks5
    server: socks.example.com
    port: 1080
    username: socks-user
    password: socks-pass
    udp: false
  - name: "hysteria yaml"
    type: hysteria
    server: hysteria.example.com
    port: 443
    auth-str: hy-secret
    protocol: udp
    obfs: salamander
    obfs-param: obfs-secret
    up: 100
    down: 100
    sni: hysteria.example.com
    skip-cert-verify: true
    custom-hy-flag: keep-v1
  - name: "hy2 yaml"
    type: hysteria2
    server: hkt03ddns.poke-mon.xyz
    port: 20000
    ports: 20000-50000
    mport: 20000-50000
    udp: true
    password: hy2-secret
    sni: www.bing.com
    skip-cert-verify: false
    custom-hy2-flag: keep-me
`)
	result := ParseContent(content, model.SourceInfo{Name: "yaml", Kind: "file"})
	if len(result.Errors) != 0 {
		t.Fatalf("Errors = %#v, want none", result.Errors)
	}
	if len(result.Nodes) != 7 {
		t.Fatalf("len(Nodes) = %d, want 7", len(result.Nodes))
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
	if !vless.TLS.Insecure {
		t.Fatalf("vless TLS.Insecure = false, want true from skip-cert-verify")
	}
	if got := fmt.Sprint(vless.Raw["encryption"]); got != "none" {
		t.Fatalf("vless raw encryption = %q, want none", got)
	}

	mieru := result.Nodes[2]
	if mieru.Type != model.ProtocolMieru {
		t.Fatalf("third Type = %q, want %q", mieru.Type, model.ProtocolMieru)
	}
	if mieru.Auth.Username != "user" || mieru.Auth.Password != "secret" {
		t.Fatalf("mieru auth = %#v, want username/password", mieru.Auth)
	}
	if got := fmt.Sprint(mieru.Raw["portRange"]); got != "2090-2099" {
		t.Fatalf("mieru portRange = %q, want 2090-2099", got)
	}

	httpNode := result.Nodes[3]
	if httpNode.Type != model.ProtocolHTTP || httpNode.Auth.Username != "http-user" || !httpNode.TLS.Enabled || !httpNode.TLS.Insecure {
		t.Fatalf("http yaml node = %#v, want auth and TLS parsed", httpNode)
	}

	socksNode := result.Nodes[4]
	if socksNode.Type != model.ProtocolSOCKS5 || socksNode.Auth.Password != "socks-pass" || socksNode.UDP == nil || *socksNode.UDP {
		t.Fatalf("socks yaml node = %#v, want socks5 auth and udp false", socksNode)
	}

	hy := result.Nodes[5]
	if hy.Type != model.ProtocolHysteria {
		t.Fatalf("sixth Type = %q, want %q", hy.Type, model.ProtocolHysteria)
	}
	if hy.Auth.Password != "hy-secret" || !hy.TLS.Insecure || hy.TLS.SNI != "hysteria.example.com" {
		t.Fatalf("hysteria yaml node = %#v, want auth and TLS parsed", hy)
	}
	hyFields, ok := hy.Raw["_mihomoProxyFields"].(map[string]interface{})
	if !ok {
		t.Fatalf("hysteria raw mihomo fields = %#v, want map", hy.Raw["_mihomoProxyFields"])
	}
	for key, want := range map[string]string{
		"auth-str":       "hy-secret",
		"protocol":       "udp",
		"obfs-param":     "obfs-secret",
		"up":             "100",
		"down":           "100",
		"custom-hy-flag": "keep-v1",
	} {
		if got := fmt.Sprint(hyFields[key]); got != want {
			t.Fatalf("hysteria original field %s = %q, want %q", key, got, want)
		}
	}

	hy2 := result.Nodes[6]
	if hy2.Type != model.ProtocolHysteria2 {
		t.Fatalf("seventh Type = %q, want %q", hy2.Type, model.ProtocolHysteria2)
	}
	fields, ok := hy2.Raw["_mihomoProxyFields"].(map[string]interface{})
	if !ok {
		t.Fatalf("hy2 raw mihomo fields = %#v, want map", hy2.Raw["_mihomoProxyFields"])
	}
	for key, want := range map[string]string{
		"ports":           "20000-50000",
		"mport":           "20000-50000",
		"custom-hy2-flag": "keep-me",
	} {
		if got := fmt.Sprint(fields[key]); got != want {
			t.Fatalf("hy2 original field %s = %q, want %q", key, got, want)
		}
	}
}

func TestParseContentMihomoYAMLFlexibleScalarTypes(t *testing.T) {
	content := []byte(`
proxies:
  - name: "http string fields"
    type: http
    server: http.example.com
    port: "8200"
    username: user
    password: pass
    tls: "true"
    skip-cert-verify: "true"
    alpn: h2,http/1.1
  - name: "vless scalar transport"
    type: vless
    server: vless.example.com
    port: "443"
    uuid: uuid-1
    tls: "true"
    network: ws
    ws-opts:
      path: /ws
      headers: vless.example.com
  - name: "h2 scalar host"
    type: trojan
    server: trojan.example.com
    port: "443"
    password: secret
    network: h2
    h2-opts:
      host: trojan.example.com
      path: /h2
  - name: "wg string fields"
    type: wireguard
    server: wg.example.com
    port: "51820"
    private-key: private
    public-key: public
    allowed-ips: 0.0.0.0/0,::/0
    dns: 1.1.1.1,8.8.8.8
    mtu: "1280"
    persistent-keepalive: "25"
    remote-dns-resolve: "true"
    peers:
      - server: peer.example.com
        port: "51821"
        public-key: peer-public
        allowed-ips: 10.0.0.0/8
`)

	result := ParseContent(content, model.SourceInfo{Name: "yaml", Kind: "file"})
	if len(result.Errors) != 0 {
		t.Fatalf("Errors = %#v, want none", result.Errors)
	}
	if len(result.Nodes) != 4 {
		t.Fatalf("len(Nodes) = %d, want 4", len(result.Nodes))
	}

	httpNode := result.Nodes[0]
	if httpNode.Port != 8200 || !httpNode.TLS.Enabled || !httpNode.TLS.Insecure {
		t.Fatalf("http node = %#v, want string port/bool parsed", httpNode)
	}
	if got := strings.Join(httpNode.TLS.ALPN, ","); got != "h2,http/1.1" {
		t.Fatalf("http ALPN = %q, want h2,http/1.1", got)
	}

	vless := result.Nodes[1]
	if vless.Transport.Host != "vless.example.com" {
		t.Fatalf("vless ws host = %q, want scalar headers host", vless.Transport.Host)
	}

	trojan := result.Nodes[2]
	if got := strings.Join(trojan.Transport.H2Hosts, ","); got != "trojan.example.com" {
		t.Fatalf("trojan h2 hosts = %q, want scalar host", got)
	}

	wg := result.Nodes[3]
	if wg.Port != 0 || wg.WireGuard == nil || len(wg.WireGuard.Peers) != 1 {
		t.Fatalf("wireguard node = %#v, want parsed peer form", wg)
	}
	if wg.WireGuard.MTU != 1280 || wg.WireGuard.PersistentKeepalive != 25 || !wg.WireGuard.RemoteDNSResolve {
		t.Fatalf("wireguard options = %#v, want string numeric/bool parsed", wg.WireGuard)
	}
	if got := strings.Join(wg.WireGuard.DNS, ","); got != "1.1.1.1,8.8.8.8" {
		t.Fatalf("wireguard DNS = %q, want parsed csv", got)
	}
	if wg.WireGuard.Peers[0].Port != 51821 || strings.Join(wg.WireGuard.Peers[0].AllowedIPs, ",") != "10.0.0.0/8" {
		t.Fatalf("wireguard peer = %#v, want string fields parsed", wg.WireGuard.Peers[0])
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
