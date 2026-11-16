package rulevalidation

import (
	"errors"
	"testing"
)

func TestValidateRulesAcceptsCompleteRoutingRules(t *testing.T) {
	for _, line := range []string{
		"DOMAIN,example.com,DIRECT", "DOMAIN-SUFFIX,example.com,我的 代理组", "DOMAIN-KEYWORD,example,REJECT",
		"IP-CIDR,192.0.2.0/24,DIRECT,no-resolve", "IP-CIDR6,2001:db8::/32,代理,src,no-resolve",
		"DST-PORT,80/443/8000-9000,PROXY", "NETWORK,UDP,REJECT", "DSCP,10,自定义策略",
		`DOMAIN-REGEX,^foo,(bar|baz)\.example$,我的代理`, `PROCESS-NAME-REGEX,^(foo)\1$,DIRECT`,
		"AND,((DOMAIN,example.com),(NETWORK,TCP)),我的代理", "NOT,((NETWORK,UDP)),REJECT",
		"RULE-SET,custom-provider,我的代理", "RULE-SET,custom-provider,DIRECT,no-resolve", "RULE-SET,custom_provider,PROXY,src",
		"- 'DOMAIN-SUFFIX,example.com,DIRECT' # comment",
	} {
		if err := ValidateRules([]string{line}); err != nil {
			t.Errorf("valid routing rule %q: %v", line, err)
		}
	}
	if err := ValidateRules([]string{"", "# no extra rules"}); err != nil {
		t.Fatal(err)
	}
}

func TestValidateRulesRejectsMalformedRoutingRulesAtOriginalLine(t *testing.T) {
	for _, line := range []string{
		"DOMAIN,example.com", "DOMAIN,,DIRECT", "DOMAIN,example.com,", "DOMAIN,example.com,DIRECT,extra", "DOMAIN,example.com,DIRECT,REJECT", "IP-CIDR,192.0.2.0/24,DIRECT,PROXY",
		"DOMIAN,example.com,DIRECT", "DOMAIN,https://example.com,DIRECT", "IP-CIDR,192.0.2.0/33,DIRECT",
		"IP-CIDR,192.0.2.0/24,no-resolve", "DST-PORT,99999,我的代理", "NETWORK,QUIC,DIRECT",
		"DOMAIN-REGEX,[,DIRECT", "DOMAIN-REGEX,^foo$", "AND,((DOMAIN,example.com),(NETWORK,TCP),DIRECT",
		"AND,((DOMAIN,example.com)),DIRECT,REJECT", "NOT,((DOMAIN,example.com,DIRECT)),REJECT",
		"RULE-SET,provider", "RULE-SET,,DIRECT", "RULE-SET,provider,,no-resolve", "RULE-SET,provider,DIRECT,unknown",
		"MATCH,DIRECT", "FINAL,DIRECT", "SUB-RULE,(NETWORK,TCP),sub-rule-name",
	} {
		err := ValidateRules([]string{"# comment", "", line})
		var detail *Error
		if !errors.As(err, &detail) || detail.Line != 3 {
			t.Errorf("invalid routing rule accepted or wrong line %q: %v", line, err)
		}
	}
}
