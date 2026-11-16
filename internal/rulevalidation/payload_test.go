package rulevalidation

import (
	"errors"
	"strings"
	"testing"
)

func TestValidatePayloadAcceptsMihomoSyntax(t *testing.T) {
	tests := []struct {
		behavior string
		payload  []string
	}{
		{"domain", []string{"# comment", "example.com", "+.example.net", "*.example.org", "sub.*.example.org", ".example.io", "localhost", "例子.中国", "_service.example", "*"}},
		{"domain", []string{"- 'example.com'", "- \"+.example.net\"", "example.org # note", "- 'quoted.example' # note", "- # ignored"}},
		{"ipcidr", []string{"192.0.2.0/24", "2001:db8::/32", "0.0.0.0/0", "::/0"}},
		{"classical", []string{
			"DOMAIN-SUFFIX,example.com", "DOMAIN,example.org,DIRECT", "DOMAIN-KEYWORD,tracker", "DOMAIN-WILDCARD,*.example.*",
			"IP-CIDR,192.0.2.0/24,no-resolve", "IP-CIDR,2001:db8::/32,src", "IP-CIDR6,2001:db8::/32",
			"SRC-IP-CIDR,192.0.2.0/24", "IP-SUFFIX,0.0.0.1/8", "SRC-IP-SUFFIX,::1/16",
			"IP-ASN,13335,no-resolve", "SRC-IP-ASN,64512", "GEOSITE,geolocation-!cn", "GEOIP,LAN,src,no-resolve", "SRC-GEOIP,CN",
			"NETWORK,TCP", "NETWORK,udp", "DST-PORT,80/443/8000-9000", "SRC-PORT,0-65535", "DST-PORT,443-80", "IN-PORT,7890", "DSCP,0/10-20/63", "DSCP,*", "UID,0-4294967295",
			"IN-TYPE,HTTP/SOCKS/TUN", "IN-NAME,http-in/tun-in", "IN-USER,alice/bob", "REMATCH-NAME,one/two",
			"PROCESS-NAME,Firefox.exe", "PROCESS-PATH,C:\\Program Files\\Example\\app.exe", "PROCESS-NAME-WILDCARD,firefox*", "PROCESS-PATH-WILDCARD,/usr/*/app",
			`DOMAIN-REGEX,^foo,(bar|baz)\.example$`, `DOMAIN-REGEX,^(?!ads\.).*\.example\.com$`, `PROCESS-NAME-REGEX,^(fire)\1$`,
			"AND,((DOMAIN-SUFFIX,example.com),(NETWORK,TCP))", "OR,((DOMAIN,one.example),(NOT,((DOMAIN,two.example))))", "NOT,((NETWORK,UDP))",
			`AND,((DOMAIN-REGEX,^(foo|bar)\.example$),(NETWORK,TCP))`,
			`NOT,((DOMAIN-REGEX,^[()]foo\(bar\)$))`,
			"IP-CIDR,192.0.2.0/24,DIRECT,no-resolve",
		}},
	}
	for _, tc := range tests {
		if err := ValidatePayload(tc.behavior, tc.payload); err != nil {
			t.Errorf("%s: %v", tc.behavior, err)
		}
	}
}

func TestValidatePayloadRejectsMalformedEntriesAtTheirOriginalLine(t *testing.T) {
	tests := []struct{ behavior, line string }{
		{"domain", "https://example.com/path"}, {"domain", "example..com"}, {"domain", "example.com."}, {"domain", "exam ple.com"}, {"domain", "example.com:443"},
		{"domain", "a*b.example.com"}, {"domain", "foo.+.example.com"}, {"domain", "+"}, {"domain", "<html>"}, {"domain", "example.com,foo"}, {"domain", "-invalid.example"},
		{"ipcidr", "192.0.2.0/33"}, {"ipcidr", "999.1.1.0/24"}, {"ipcidr", "2001:db8::/129"}, {"ipcidr", "192.0.2.1"},
		{"classical", "DOMIAN-SUFFIX,example.com"}, {"classical", "DOMAIN-SUFFIX,"}, {"classical", "DOMAIN-SUFFIX"}, {"classical", "DOMAIN-SUFFIX,example.com,invalid-target"},
		{"classical", "IP-CIDR,192.0.2.0/33"}, {"classical", "IP-CIDR6,192.0.2.0/24"}, {"classical", "SRC-IP-SUFFIX,::1/129"},
		{"classical", "NETWORK,QUIC"}, {"classical", "DST-PORT,65536"}, {"classical", "SRC-PORT,-1"}, {"classical", "DST-PORT,80-"}, {"classical", "DST-PORT,80//443"},
		{"classical", "UID,4294967296"}, {"classical", "DSCP,64"}, {"classical", "IP-ASN,AS13335"}, {"classical", "IP-ASN,0"}, {"classical", "IN-TYPE,TYPO"},
		{"classical", "IN-NAME,one/"}, {"classical", "PROCESS-NAME,"}, {"classical", "DOMAIN-REGEX,["}, {"classical", "PROCESS-PATH-REGEX,("},
		{"classical", "AND,((DOMAIN,example.com),(NETWORK,TCP)"}, {"classical", "NOT,((DOMAIN,a),(DOMAIN,b))"}, {"classical", "AND,()"},
		{"classical", "AND,(DOMAIN,example.com)"}, {"classical", "OR,((DOMAIN,a),)"}, {"classical", "MATCH,DIRECT"}, {"classical", "RULE-SET,provider"}, {"classical", "SUB-RULE,name"},
	}
	for _, tc := range tests {
		t.Run(tc.behavior+"/"+tc.line, func(t *testing.T) {
			err := ValidatePayload(tc.behavior, []string{"# ignored", "", tc.line})
			var invalid *Error
			if !errors.As(err, &invalid) || invalid.Line != 3 || invalid.Message == "" {
				t.Fatalf("invalid entry accepted or wrong line: %v", err)
			}
		})
	}
}

func TestEffectiveBehaviorRecognizesAdvancedClassicalRules(t *testing.T) {
	for _, tp := range []string{"DOMAIN", "IN-PORT", "DSCP", "PROCESS-NAME-WILDCARD", "REMATCH-NAME", "MATCH"} {
		line := tp + ",value"
		if !IsClassicalLine(line) || EffectiveBehavior("domain", []string{line}) != "classical" {
			t.Fatalf("did not recognize %s", tp)
		}
	}
	if EffectiveBehavior("IPCIDR", []string{"192.0.2.0/24"}) != "ipcidr" || IsClassicalLine("example.com") {
		t.Fatal("plain entries classified as classical")
	}
}

func TestInvalidRegexDoesNotEchoItsContents(t *testing.T) {
	err := ValidatePayload("classical", []string{"DOMAIN-REGEX,https://example.com/private?token=sensitive-value["})
	if err == nil || strings.Contains(err.Error(), "sensitive-value") || !strings.Contains(err.Error(), "正则表达式") {
		t.Fatalf("expected a useful regex error without the private expression: %v", err)
	}
}

func TestNormalizePastedQuotesAndComments(t *testing.T) {
	for input, want := range map[string]string{
		`- "DOMAIN-SUFFIX,example.com" # comment "quoted"`: "DOMAIN-SUFFIX,example.com",
		"- 'DOMAIN-SUFFIX,example.com'\t# note 'quoted'":   "DOMAIN-SUFFIX,example.com",
		"DOMAIN-SUFFIX,example.com\t# comment":             "DOMAIN-SUFFIX,example.com",
		`'DOMAIN-REGEX,^say "hi"$' # "quoted" comment`:     `DOMAIN-REGEX,^say "hi"$`,
		`DOMAIN-REGEX,^say "hi"$`:                          `DOMAIN-REGEX,^say "hi"$`,
	} {
		if got := NormalizeLine(input); got != want {
			t.Errorf("NormalizeLine(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestValidatePayloadProtectsSizeAndLogicDepth(t *testing.T) {
	if ValidatePayload("domain", []string{"# comments only", ""}) == nil {
		t.Fatal("empty payload accepted")
	}
	if ValidatePayload("domain", []string{strings.Repeat("a", maxLineBytes+1)}) == nil {
		t.Fatal("oversized entry accepted")
	}
	line := "DOMAIN,example.com"
	for i := 0; i < maxLogicDepth+2; i++ {
		line = "NOT,((" + line + "))"
	}
	if ValidatePayload("classical", []string{line}) == nil {
		t.Fatal("excessive logical nesting accepted")
	}
	if ValidatePayload("invalid", []string{"example.com"}) == nil {
		t.Fatal("invalid behavior accepted")
	}
}
