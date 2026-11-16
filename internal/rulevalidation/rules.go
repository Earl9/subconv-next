package rulevalidation

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ValidateRules checks complete rules written directly into the routing list.
// Each entry includes its target policy. Unlike provider payloads, this list
// may be empty and may reference a RULE-SET. The application's final MATCH is
// generated separately and must not be overridden here.
func ValidateRules(rules []string) error {
	for i, raw := range rules {
		line := prepareLine(raw)
		if line == "" {
			continue
		}
		var err error
		switch {
		case len(line) > maxLineBytes:
			err = fmt.Errorf("单条规则不能超过 64 KiB")
		case !utf8.ValidString(line):
			err = fmt.Errorf("规则包含无效字符编码")
		case strings.ContainsAny(line, "\r\n\x00"):
			err = fmt.Errorf("每行只能填写一条规则，不能包含换行或空字符")
		default:
			err = validateRoutingRule(line)
		}
		if err != nil {
			return &Error{Line: i + 1, Message: err.Error()}
		}
	}
	return nil
}

func validateRoutingRule(line string) error {
	tp, rest, ok := strings.Cut(line, ",")
	tp = strings.ToUpper(strings.TrimSpace(tp))
	if tp == "MATCH" || tp == "FINAL" {
		return fmt.Errorf("默认兜底规则由系统生成，请勿填写 MATCH 或 FINAL")
	}
	if !ok || strings.TrimSpace(rest) == "" {
		return fmt.Errorf("直写规则应为 类型,匹配内容,目标策略")
	}
	if tp == "SUB-RULE" {
		return fmt.Errorf("当前配置未定义 sub-rules，不能使用 SUB-RULE；请使用普通规则或 RULE-SET")
	}
	// Mihomo's ParseRulePayload takes the last item as target for regex and
	// logical rules, preserving all intermediate commas as part of the payload.
	switch tp {
	case "DOMAIN-REGEX", "PROCESS-NAME-REGEX", "PROCESS-PATH-REGEX", "AND", "OR", "NOT":
		payload, target, ok := cutLast(rest)
		if !ok || strings.TrimSpace(payload) == "" {
			return fmt.Errorf("%s 缺少匹配内容或目标策略", tp)
		}
		if err := validatePolicy(target); err != nil {
			return err
		}
		if err := validateClassical(tp+","+strings.TrimSpace(payload), 0); err != nil {
			return err
		}
		return nil
	}
	parts := strings.Split(rest, ",")
	if len(parts) < 2 {
		return fmt.Errorf("%s 缺少目标策略，应为 类型,匹配内容,目标策略", tp)
	}
	operand := strings.TrimSpace(parts[0])
	if operand == "" {
		return fmt.Errorf("%s 的匹配内容不能为空", tp)
	}
	if err := validatePolicy(parts[1]); err != nil {
		return err
	}
	if tp == "RULE-SET" {
		if strings.ContainsFunc(operand, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) || strings.ContainsAny(operand, "()\"'") {
			return fmt.Errorf("RULE-SET 规则集名称格式错误")
		}
		for _, param := range parts[2:] {
			value := strings.ToLower(strings.TrimSpace(param))
			if value != "no-resolve" && value != "src" {
				return fmt.Errorf("RULE-SET 只支持 no-resolve 或 src 附加参数")
			}
		}
		return nil
	}
	// Routing rules have already supplied their policy. Do not let the
	// provider compatibility parser strip a second pasted policy as if it
	// were a supported parameter.
	ipParams := false
	switch tp {
	case "IP-CIDR", "IP-CIDR6", "SRC-IP-CIDR", "IP-SUFFIX", "SRC-IP-SUFFIX", "IP-ASN", "SRC-IP-ASN", "GEOIP", "SRC-GEOIP":
		ipParams = true
	}
	for _, param := range parts[2:] {
		value := strings.ToLower(strings.TrimSpace(param))
		if !ipParams || value != "no-resolve" && value != "src" {
			return fmt.Errorf("%s 包含多余或不支持的附加参数", tp)
		}
	}
	payload := tp + "," + operand
	if len(parts) > 2 {
		payload += "," + strings.Join(parts[2:], ",")
	}
	return validateClassical(payload, 0)
}

func validatePolicy(raw string) error {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fmt.Errorf("目标策略不能为空，请填写 DIRECT、REJECT 或已存在的代理组名称")
	}
	if strings.ContainsAny(value, ",\r\n") || strings.ContainsFunc(value, unicode.IsControl) {
		return fmt.Errorf("目标策略不能包含逗号、换行或控制字符")
	}
	if strings.EqualFold(value, "no-resolve") || strings.EqualFold(value, "src") {
		return fmt.Errorf("缺少目标策略；no-resolve 和 src 应放在目标策略之后")
	}
	return nil
}

func cutLast(value string) (string, string, bool) {
	index := strings.LastIndexByte(value, ',')
	if index < 0 {
		return "", "", false
	}
	return value[:index], value[index+1:], true
}
