package renderer

import (
	"bytes"
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestBatchedYAMLEncodingMatchesSingleEncoder(t *testing.T) {
	proxies := sequenceNode()
	groups := sequenceNode()
	rules := sequenceNode()
	var names []string
	for i := 0; i < 1100; i++ {
		name := fmt.Sprintf("🇭🇰 HK node-%d [yes] # \\u1234", i)
		if i%100 == 0 {
			name += "\nsecond line\tquoted: \"value\""
		}
		names = append(names, name)
		proxy := flowMappingNode()
		appendMap(proxy, "name", scalarNode(name))
		appendMap(proxy, "port", scalarNode(443))
		proxies.Content = append(proxies.Content, proxy)
		rules.Content = append(rules.Content, scalarNode("DOMAIN,node.example,"+name))
	}
	for i := 0; i < 4; i++ {
		group := flowMappingNode()
		appendMap(group, "name", scalarNode(fmt.Sprintf("group-%d", i)))
		appendMap(group, "proxies", flowStringSeqNode(names))
		groups.Content = append(groups.Content, group)
	}
	for key, value := range map[string]*yaml.Node{"proxies": proxies, "proxy-groups": groups, "rules": rules, "empty": sequenceNode()} {
		t.Run(key, func(t *testing.T) {
			var got, want bytes.Buffer
			if err := encodeYAMLSection(&got, key, value); err != nil {
				t.Fatal(err)
			}
			doc := mappingNode()
			appendMap(doc, key, value)
			encoder := yaml.NewEncoder(&want)
			encoder.SetIndent(2)
			if err := encoder.Encode(doc); err != nil {
				t.Fatal(err)
			}
			if err := encoder.Close(); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got.Bytes(), bytes.TrimSuffix(want.Bytes(), []byte("\n"))) {
				t.Fatal("batched encoding changed YAML bytes")
			}
		})
	}
}
