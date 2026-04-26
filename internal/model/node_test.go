package model

import "testing"

func TestNormalizeNode(t *testing.T) {
	node := NodeIR{
		Type:   ProtocolVLESS,
		Name:   "  香港 HK\t\n",
		Server: "Example.COM",
		Port:   443,
		Auth: Auth{
			UUID: "uuid-1",
		},
	}

	got := NormalizeNode(node)
	if got.Name != "香港 HK" {
		t.Fatalf("Name = %q, want %q", got.Name, "香港 HK")
	}
	if got.Server != "example.com" {
		t.Fatalf("Server = %q, want %q", got.Server, "example.com")
	}
	if len(got.Tags) != 1 || got.Tags[0] != "HK" {
		t.Fatalf("Tags = %#v, want %#v", got.Tags, []string{"HK"})
	}
	if got.ID == "" {
		t.Fatalf("ID is empty")
	}
}

func TestDedupeNodes(t *testing.T) {
	nodes := []NodeIR{
		NormalizeNode(NodeIR{
			Type:   ProtocolTrojan,
			Name:   "node-a",
			Server: "example.com",
			Port:   443,
			Auth:   Auth{Password: "secret"},
			Tags:   []string{"US"},
			Source: SourceInfo{Name: "first"},
		}),
		NormalizeNode(NodeIR{
			Type:   ProtocolTrojan,
			Name:   "node-b",
			Server: "EXAMPLE.COM",
			Port:   443,
			Auth:   Auth{Password: "secret"},
			Tags:   []string{"HK"},
			Source: SourceInfo{Name: "second"},
		}),
	}

	got := DedupeNodes(nodes)
	if len(got) != 1 {
		t.Fatalf("len(DedupeNodes) = %d, want 1", len(got))
	}
	if len(got[0].Tags) != 2 {
		t.Fatalf("Tags = %#v, want merged tags", got[0].Tags)
	}
	if len(got[0].Warnings) != 1 || got[0].Warnings[0] != "duplicate skipped from source second" {
		t.Fatalf("Warnings = %#v, want duplicate warning", got[0].Warnings)
	}
}
