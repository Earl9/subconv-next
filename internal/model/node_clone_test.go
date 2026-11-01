package model

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"unicode"
)

func TestCloneNodeDoesNotShareMutableFields(t *testing.T) {
	original := NodeIR{
		ID: "node", Name: "node", Type: ProtocolWireGuard, UDP: Bool(true),
		Tags: []string{"HK"}, Sources: []SourceInfo{{Name: "source"}}, Warnings: []string{"warning"},
		TLS:       TLSOptions{ALPN: []string{"h2"}, Reality: &RealityOptions{PublicKey: "key"}, ECH: &ECHOptions{Config: "ech"}},
		Transport: TransportOptions{NoGRPCHeader: Bool(true), H2Hosts: []string{"host"}, Headers: map[string]string{"Host": "host"}},
		Raw: map[string]interface{}{
			"nested":  map[string]interface{}{"items": []interface{}{map[string]interface{}{"value": "original"}}},
			"strings": []string{"original"}, "ints": []int{1}, "bytes": []byte{1},
			"headers": map[string]string{"key": "original"},
		},
		WireGuard: &WireGuardOptions{
			AllowedIPs: []string{"0.0.0.0/0"}, Reserved: []int{1, 2, 3}, DNS: []string{"1.1.1.1"},
			Peers:     []WGPeer{{AllowedIPs: []string{"::/0"}, Reserved: []int{1, 2, 3}}},
			AmneziaWG: map[string]interface{}{"jc": []interface{}{float64(3)}},
		},
	}
	before, _ := json.Marshal(original)
	copy := CloneNode(original)
	copied, _ := json.Marshal(copy)
	if !bytes.Equal(before, copied) {
		t.Fatal("cloning changed serialized node contents")
	}
	*copy.UDP = false
	copy.Tags[0], copy.Sources[0].Name, copy.Warnings[0] = "changed", "changed", "changed"
	copy.TLS.ALPN[0], copy.TLS.Reality.PublicKey, copy.TLS.ECH.Config = "changed", "changed", "changed"
	*copy.Transport.NoGRPCHeader = false
	copy.Transport.H2Hosts[0], copy.Transport.Headers["Host"] = "changed", "changed"
	copy.Raw["nested"].(map[string]interface{})["items"].([]interface{})[0].(map[string]interface{})["value"] = "changed"
	copy.Raw["strings"].([]string)[0] = "changed"
	copy.Raw["ints"].([]int)[0] = 2
	copy.Raw["bytes"].([]byte)[0] = 2
	copy.Raw["headers"].(map[string]string)["key"] = "changed"
	copy.WireGuard.AllowedIPs[0], copy.WireGuard.DNS[0] = "changed", "changed"
	copy.WireGuard.Reserved[0], copy.WireGuard.Peers[0].Reserved[0] = 2, 2
	copy.WireGuard.Peers[0].AllowedIPs[0] = "changed"
	copy.WireGuard.AmneziaWG["jc"].([]interface{})[0] = float64(5)
	after, _ := json.Marshal(original)
	if !bytes.Equal(before, after) {
		t.Fatal("mutating the clone changed the original node")
	}
}

func FuzzSanitizeTextMatchesNormalization(f *testing.F) {
	for _, seed := range []string{"", "node-1.example", " HK node ", "foo\tbar\n baz", "a\u00a0b\u2003c", "a\x00b", "a\xffb", "香港🇭🇰"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		var normalized strings.Builder
		for _, r := range strings.TrimSpace(input) {
			if r == '\n' || r == '\r' || r == '\t' {
				normalized.WriteByte(' ')
			} else if !unicode.IsControl(r) {
				normalized.WriteRune(r)
			}
		}
		want := strings.Join(strings.Fields(normalized.String()), " ")
		if got := sanitizeText(input); got != want {
			t.Fatalf("sanitizeText(%q) = %q, want %q", input, got, want)
		}
	})
}
