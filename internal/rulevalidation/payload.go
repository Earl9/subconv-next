// Package rulevalidation checks the rule-provider payload syntax accepted by
// Mihomo without loading geodata or fetching remote content.
package rulevalidation

import (
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dlclark/regexp2"
)

// Error identifies a bad entry using its one-based index in the supplied payload.
type Error struct {
	Line    int
	Message string
}

func (e *Error) Error() string { return fmt.Sprintf("第 %d 行：%s", e.Line, e.Message) }

const maxLineBytes = 64 * 1024
const maxLogicDepth = 32

// IsClassicalLine reports recognized classical syntax, including rule types
// forbidden inside a provider so that they receive an explicit validation error.
func IsClassicalLine(line string) bool {
	line = prepareLine(line)
	tp, _, ok := strings.Cut(line, ",")
	if !ok {
		return false
	}
	switch strings.ToUpper(strings.TrimSpace(tp)) {
	case "DOMAIN", "DOMAIN-SUFFIX", "DOMAIN-KEYWORD", "DOMAIN-WILDCARD", "DOMAIN-REGEX",
		"GEOSITE", "GEOIP", "SRC-GEOIP", "IP-CIDR", "IP-CIDR6", "SRC-IP-CIDR", "IP-SUFFIX", "SRC-IP-SUFFIX", "IP-ASN", "SRC-IP-ASN",
		"SRC-PORT", "DST-PORT", "IN-PORT", "DSCP", "UID", "NETWORK", "IN-TYPE", "IN-NAME", "IN-USER", "REMATCH-NAME",
		"PROCESS-NAME", "PROCESS-PATH", "PROCESS-NAME-REGEX", "PROCESS-PATH-REGEX", "PROCESS-NAME-WILDCARD", "PROCESS-PATH-WILDCARD",
		"AND", "OR", "NOT", "MATCH", "RULE-SET", "SUB-RULE":
		return true
	}
	return false
}

// EffectiveBehavior preserves the application's existing automatic promotion
// of a provider containing classical entries to classical behavior.
func EffectiveBehavior(behavior string, payload []string) string {
	for _, line := range payload {
		if IsClassicalLine(line) {
			return "classical"
		}
	}
	return strings.ToLower(strings.TrimSpace(behavior))
}

// ValidatePayload validates entries without modifying them. Blank lines and
// comments are ignored; YAML list markers and paired quotes are accepted.
func ValidatePayload(behavior string, payload []string) error {
	behavior = strings.ToLower(strings.TrimSpace(behavior))
	switch behavior {
	case "domain", "ipcidr", "classical":
	default:
		return &Error{Line: 1, Message: "规则行为必须为 domain、ipcidr 或 classical"}
	}
	count := 0
	for i, raw := range payload {
		line := NormalizePayloadLine(raw)
		if line == "" {
			continue
		}
		count++
		var err error
		switch {
		case len(line) > maxLineBytes:
			err = fmt.Errorf("单条规则不能超过 64 KiB")
		case !utf8.ValidString(line):
			err = fmt.Errorf("规则包含无效字符编码")
		case strings.ContainsAny(line, "\r\n\x00"):
			err = fmt.Errorf("每行只能填写一条规则，不能包含换行或空字符")
		default:
			switch behavior {
			case "domain":
				err = validateDomain(line, true)
			case "ipcidr":
				err = validatePrefix(line)
			case "classical":
				err = validateClassical(line, 0)
			}
		}
		if err != nil {
			return &Error{Line: i + 1, Message: err.Error()}
		}
	}
	if count == 0 {
		return &Error{Line: 1, Message: "至少填写一条有效规则"}
	}
	return nil
}

func prepareLine(raw string) string { return NormalizeLine(raw) }

// NormalizeLine removes pasted YAML list markers, scalar quotes, and comments.
// Callers split physical lines before calling it so a comment cannot hide the
// following rule. Keep this shared by validation and rendering.
func NormalizeLine(raw string) string {
	line := strings.TrimSpace(raw)
	if strings.HasPrefix(line, "#") {
		return ""
	}
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "-\t") {
		line = strings.TrimSpace(line[1:])
	}
	if len(line) >= 2 && (line[0] == '\'' || line[0] == '"') {
		for end := 1; end < len(line); end++ {
			if line[0] == '"' && line[end] == '\\' {
				end++
				continue
			}
			if line[end] != line[0] {
				continue
			}
			if line[0] == '\'' && end+1 < len(line) && line[end+1] == '\'' {
				end++
				continue
			}
			tail := strings.TrimSpace(line[end+1:])
			if tail == "" || strings.HasPrefix(tail, "#") {
				return strings.TrimSpace(line[1:end])
			}
		}
	}
	for i := 1; i < len(line); i++ {
		if line[i] == '#' && (line[i-1] == ' ' || line[i-1] == '\t') {
			line = strings.TrimSpace(line[:i])
			break
		}
	}
	// Comments may follow a quoted YAML scalar. Remove its quotes again
	// after removing the comment so these common pasted entries stay valid.
	if len(line) >= 2 && ((line[0] == '\'' && line[len(line)-1] == '\'') || (line[0] == '"' && line[len(line)-1] == '"')) {
		line = line[1 : len(line)-1]
	}
	if strings.HasPrefix(line, "#") {
		return ""
	}
	return line
}

func validateDomain(value string, patterns bool) error {
	if value == "" {
		return fmt.Errorf("域名不能为空")
	}
	if strings.ContainsAny(value, "/\\:,;@?#<>=()[]{}\"'") || strings.ContainsFunc(value, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) {
		return fmt.Errorf("域名格式错误，请填写域名而非 URL、IP 网段或完整规则")
	}
	if utf8.RuneCountInString(value) > 253 {
		return fmt.Errorf("域名过长")
	}
	labels := strings.Split(value, ".")
	for i, label := range labels {
		if label == "" {
			if patterns && i == 0 && len(labels) > 1 {
				continue
			}
			return fmt.Errorf("域名包含空标签或多余的点")
		}
		if patterns && (label == "*" || label == "+" && i == 0 && len(labels) > 1) {
			continue
		}
		if strings.ContainsAny(label, "*+") {
			return fmt.Errorf("域名通配符必须独占标签；+ 只能用于开头的 +.域名")
		}
		if utf8.RuneCountInString(label) > 63 {
			return fmt.Errorf("域名标签不能超过 63 个字符")
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return fmt.Errorf("域名标签不能以连字符开头或结尾")
		}
		for _, r := range label {
			if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '-' && r != '_' {
				return fmt.Errorf("域名包含不允许的字符")
			}
		}
	}
	return nil
}

func validatePrefix(value string) error {
	if _, err := netip.ParsePrefix(value); err != nil {
		return fmt.Errorf("IP 网段格式错误，请填写有效地址和掩码，例如 192.0.2.0/24 或 2001:db8::/32")
	}
	return nil
}

func validateClassical(line string, depth int) error {
	if depth > maxLogicDepth {
		return fmt.Errorf("逻辑规则嵌套过深，最多 32 层")
	}
	tp, rest, ok := strings.Cut(line, ",")
	tp = strings.ToUpper(strings.TrimSpace(tp))
	if !ok || strings.TrimSpace(rest) == "" {
		return fmt.Errorf("规则缺少类型或匹配内容，应为 类型,内容")
	}
	switch tp {
	case "MATCH", "RULE-SET", "SUB-RULE":
		return fmt.Errorf("%s 不能用于 classical 规则集", tp)
	}
	switch tp {
	case "DOMAIN-REGEX", "PROCESS-NAME-REGEX", "PROCESS-PATH-REGEX":
		if _, err := regexp2.Compile(rest, regexp2.IgnoreCase); err != nil {
			// regexp2's error includes the whole expression, which can contain
			// private URL parameters. Keep diagnostics tied to the rule's line.
			return fmt.Errorf("%s 正则表达式错误，请检查括号、转义或量词", tp)
		}
		return nil
	case "AND", "OR", "NOT":
		return validateLogic(tp, rest, depth)
	}
	parts := strings.Split(rest, ",")
	operand := strings.TrimSpace(parts[0])
	if operand == "" {
		return fmt.Errorf("%s 的匹配内容不能为空", tp)
	}
	ipParams := false
	switch tp {
	case "IP-CIDR", "IP-CIDR6", "SRC-IP-CIDR", "IP-SUFFIX", "SRC-IP-SUFFIX", "IP-ASN", "SRC-IP-ASN", "GEOIP", "SRC-GEOIP":
		ipParams = true
	}
	for _, param := range parts[1:] {
		p := strings.ToLower(strings.TrimSpace(param))
		if !ipParams || p != "no-resolve" && p != "src" {
			return fmt.Errorf("%s 包含多余或不支持的参数；规则集内无需填写目标策略", tp)
		}
	}
	switch tp {
	case "DOMAIN", "DOMAIN-SUFFIX":
		return validateDomain(operand, false)
	case "DOMAIN-KEYWORD":
		if strings.ContainsAny(operand, "/\\:<>") || strings.ContainsFunc(operand, unicode.IsSpace) {
			return fmt.Errorf("DOMAIN-KEYWORD 应填写域名关键词")
		}
	case "DOMAIN-WILDCARD":
		if strings.ContainsAny(operand, "/\\:<>") || strings.ContainsFunc(operand, unicode.IsSpace) {
			return fmt.Errorf("DOMAIN-WILDCARD 应填写域名通配模式")
		}
	case "IP-CIDR", "IP-CIDR6", "SRC-IP-CIDR", "IP-SUFFIX", "SRC-IP-SUFFIX":
		if err := validatePrefix(operand); err != nil {
			return err
		}
		if tp == "IP-CIDR6" {
			p, _ := netip.ParsePrefix(operand)
			if !p.Addr().Is6() {
				return fmt.Errorf("IP-CIDR6 必须填写 IPv6 网段")
			}
		}
	case "SRC-PORT", "DST-PORT", "IN-PORT":
		return validateRanges(operand, 65535, "端口")
	case "DSCP":
		if operand == "*" {
			return nil
		}
		return validateRanges(operand, 63, "DSCP")
	case "UID":
		return validateRanges(operand, 4294967295, "UID")
	case "IP-ASN", "SRC-IP-ASN":
		if n, err := strconv.ParseUint(operand, 10, 32); err != nil || n == 0 {
			return fmt.Errorf("ASN 必须填写 1–4294967295 的数字")
		}
	case "NETWORK":
		if !strings.EqualFold(operand, "TCP") && !strings.EqualFold(operand, "UDP") {
			return fmt.Errorf("NETWORK 只能填写 TCP 或 UDP")
		}
	case "IN-TYPE":
		for _, value := range strings.Split(operand, "/") {
			switch strings.ToUpper(strings.TrimSpace(value)) {
			case "HTTP", "HTTPS", "SOCKS", "SOCKS4", "SOCKS5", "SHADOWSOCKS", "SNELL", "VMESS", "VLESS", "REDIR", "TPROXY", "TROJAN", "TUNNEL", "TUN", "TUIC", "HYSTERIA2", "ANYTLS", "MIERU", "SUDOKU", "TRUSTTUNNEL", "SHADOWQUIC", "INNER":
			default:
				return fmt.Errorf("不支持的 IN-TYPE 类型：%s", value)
			}
		}
	case "IN-NAME", "IN-USER", "REMATCH-NAME":
		for _, part := range strings.Split(operand, "/") {
			if strings.TrimSpace(part) == "" {
				return fmt.Errorf("%s 的名称不能留空", tp)
			}
		}
	case "GEOSITE", "GEOIP", "SRC-GEOIP":
		if strings.ContainsFunc(operand, unicode.IsSpace) || strings.ContainsAny(operand, "<>\\") {
			return fmt.Errorf("%s 分类名称格式错误", tp)
		}
	case "PROCESS-NAME", "PROCESS-PATH", "PROCESS-NAME-WILDCARD", "PROCESS-PATH-WILDCARD":
		// Process paths and names may include spaces, platform-specific separators,
		// wildcard characters, or non-ASCII text. Their existence is runtime-specific.
	default:
		return fmt.Errorf("不支持的规则类型：%s", tp)
	}
	return nil
}

// NormalizePayloadLine preserves the compatibility with pasted routing rules
// while respecting Mihomo's comma-containing regex and logical payloads.
func NormalizePayloadLine(raw string) string {
	line := NormalizeLine(raw)
	tp, _, ok := strings.Cut(line, ",")
	if !ok {
		return line
	}
	tp = strings.ToUpper(strings.TrimSpace(tp))
	parts := strings.Split(line, ",")
	index := 2 // ordinary routing rules keep their policy before any parameters
	switch tp {
	case "DOMAIN-REGEX", "PROCESS-NAME-REGEX", "PROCESS-PATH-REGEX":
		// A regex may end in a literal comma and policy name. Provider grammar
		// has no target, so removing that suffix changes the user's expression.
		return line
	case "AND", "OR", "NOT":
		index = len(parts) - 1
	}
	if index < len(parts) && knownPolicy(strings.TrimSpace(parts[index])) {
		parts = append(parts[:index], parts[index+1:]...)
	}
	if index == 2 {
		for i := 2; i < len(parts); i++ {
			if flag := strings.ToLower(strings.TrimSpace(parts[i])); flag == "no-resolve" || flag == "src" {
				parts[i] = flag
			}
		}
	}
	return strings.TrimSpace(strings.Join(parts, ","))
}

func knownPolicy(value string) bool {
	switch strings.ToUpper(value) {
	case "PROXY", "DIRECT", "REJECT", "🚀 节点选择", "⚡ 自动选择", "🐟 漏网之鱼":
		return true
	}
	return false
}

func validateRanges(value string, max uint64, label string) error {
	parts := strings.Split(value, "/")
	if len(parts) > 28 {
		return fmt.Errorf("%s 最多包含 28 个区间", label)
	}
	for _, part := range parts {
		bounds := strings.Split(part, "-")
		if len(bounds) > 2 || part == "" {
			return fmt.Errorf("%s 格式错误，请使用数字、起止区间或 / 分隔的多个区间", label)
		}
		for _, b := range bounds {
			n, err := strconv.ParseUint(strings.Trim(b, " []"), 10, 64)
			if err != nil || n > max {
				return fmt.Errorf("%s 必须在 0–%d 范围内", label, max)
			}
		}
	}
	return nil
}

func validateLogic(tp, value string, depth int) error {
	if len(value) < 2 || value[0] != '(' || value[len(value)-1] != ')' {
		return fmt.Errorf("%s 逻辑规则必须使用括号包裹，例如 AND,((DOMAIN,example.com),(NETWORK,TCP))", tp)
	}
	inside := strings.TrimSpace(value[1 : len(value)-1])
	count := 0
	for inside != "" {
		if inside[0] != '(' {
			return fmt.Errorf("%s 子规则必须分别用括号包裹", tp)
		}
		// Mihomo logic.format counts literal parentheses, including those in
		// regex payloads. Keep this scan aligned with its parsing rather than
		// applying regexp escape or character-class rules to the outer grammar.
		balance := 0
		end := -1
		for i, r := range inside {
			if r == '(' {
				balance++
			} else if r == ')' {
				balance--
				if balance == 0 {
					end = i
					break
				}
			}
		}
		if end < 0 {
			return fmt.Errorf("%s 逻辑规则括号不匹配", tp)
		}
		if err := validateClassical(strings.TrimSpace(inside[1:end]), depth+1); err != nil {
			return err
		}
		count++
		inside = strings.TrimSpace(inside[end+1:])
		if inside != "" {
			if inside[0] != ',' {
				return fmt.Errorf("%s 子规则之间必须用逗号分隔", tp)
			}
			inside = strings.TrimSpace(inside[1:])
			if inside == "" {
				return fmt.Errorf("%s 末尾存在多余逗号", tp)
			}
		}
	}
	if count == 0 || tp == "NOT" && count != 1 {
		return fmt.Errorf("%s 必须包含%s子规则", tp, map[bool]string{true: "一个", false: "至少一个"}[tp == "NOT"])
	}
	return nil
}
