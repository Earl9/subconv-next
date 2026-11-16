package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"subconv-next/internal/model"
	"subconv-next/internal/rulevalidation"
)

// CustomRuleValidationError identifies the input to fix without including the
// contents of a rule or a URL, which may contain private query parameters.
type CustomRuleValidationError struct {
	Field   string
	Line    int
	Message string
}

func (err *CustomRuleValidationError) Error() string {
	if err.Line > 0 {
		return fmt.Sprintf("%s 第 %d 行：%s", err.Field, err.Line, err.Message)
	}
	return err.Field + ": " + err.Message
}

// ValidateCustomRule validates a normalized custom rule. Payload errors retain
// their original input line even when pasted concatenated rules are repaired.
func ValidateCustomRule(rule model.CustomRule) error {
	invalid := func(field, message string) error {
		return &CustomRuleValidationError{Field: field, Message: message}
	}
	if strings.TrimSpace(rule.Key) == "" {
		return invalid("key", "规则 Key 不能为空")
	}
	if !isValidCustomRuleKey(rule.Key) {
		return invalid("key", "规则 Key 只能包含小写字母、数字、下划线和短横线")
	}
	if strings.TrimSpace(rule.Label) == "" {
		return invalid("label", "显示名称不能为空")
	}
	if !validRulePolicyName(rule.Label) {
		return invalid("label", "显示名称不能包含逗号、换行或控制字符")
	}
	if !validRulePolicyName(rule.Icon) {
		return invalid("emoji", "图标不能包含逗号、换行或控制字符")
	}
	if !isAllowedCustomRuleTargetMode(rule.TargetMode) {
		return invalid("target_mode", "请选择有效的目标策略")
	}
	if !isAllowedCustomRuleSourceType(rule.SourceType) {
		return invalid("source_type", "请选择有效的规则来源类型")
	}
	if !isAllowedProviderBehavior(rule.Behavior) {
		return invalid("behavior", "规则行为必须为 domain、ipcidr 或 classical")
	}
	if !isAllowedProviderFormat(rule.Format) {
		return invalid("format", "规则格式必须为 yaml、text 或 mrs")
	}
	if !isAllowedCustomRuleInsertPosition(rule.InsertPosition) {
		return invalid("insert_position", "请选择有效的规则匹配位置")
	}
	if strings.EqualFold(rule.TargetMode, "existing_group") && strings.TrimSpace(rule.TargetGroup) == "" {
		return invalid("target_group", "请选择或填写目标组名称")
	}
	if !validRulePolicyName(rule.TargetGroup) {
		return invalid("target_group", "目标组名称不能包含逗号、换行或控制字符")
	}
	if strings.EqualFold(rule.SourceType, "group_only") {
		return nil
	}
	if strings.EqualFold(rule.Format, "mrs") && strings.EqualFold(rule.Behavior, "classical") {
		return invalid("format", "mrs 格式不能搭配 classical 行为")
	}
	switch strings.ToLower(strings.TrimSpace(rule.SourceType)) {
	case "http":
		if !validCustomRuleURL(rule.URL) {
			return invalid("url", "请填写完整的 http:// 或 https:// 规则 URL，不能包含空白、登录凭据或片段")
		}
		if rule.Interval <= 0 {
			return invalid("interval", "更新间隔必须是大于 0 的整数秒数")
		}
	case "file":
		if strings.TrimSpace(rule.Path) == "" {
			return invalid("path", "本地文件路径不能为空")
		}
		if strings.IndexFunc(rule.Path, unicode.IsControl) >= 0 {
			return invalid("path", "文件路径不能包含换行或控制字符")
		}
	case "inline":
		if strings.EqualFold(rule.Format, "mrs") {
			return invalid("format", "内联规则请使用 text 或 yaml，mrs 仅支持远程或文件来源")
		}
		var payload []string
		var originalLines []int
		lineNumber := 0
		for _, item := range rule.Payload {
			for _, line := range strings.Split(strings.ReplaceAll(item, "\r\n", "\n"), "\n") {
				lineNumber++
				for _, segment := range splitConcatenatedCustomRulePayloadLine(line) {
					segment = prepareCustomRulePayloadLine(segment)
					if segment != "" {
						payload = append(payload, segment)
						originalLines = append(originalLines, lineNumber)
					}
				}
			}
		}
		if len(payload) == 0 {
			return invalid("payload", "内联规则内容至少需要一条有效规则")
		}
		behavior := rulevalidation.EffectiveBehavior(rule.Behavior, payload)
		if err := rulevalidation.ValidatePayload(behavior, payload); err != nil {
			var detail *rulevalidation.Error
			if errors.As(err, &detail) {
				line := detail.Line
				if line > 0 && line <= len(originalLines) {
					line = originalLines[line-1]
				}
				return &CustomRuleValidationError{Field: "payload", Line: line, Message: detail.Message}
			}
			return invalid("payload", "规则内容格式不正确")
		}
	}
	return nil
}

func validRulePolicyName(value string) bool {
	return !strings.Contains(value, ",") && strings.IndexFunc(value, unicode.IsControl) < 0
}

func validCustomRuleURL(value string) bool {
	if value == "" || strings.IndexFunc(value, unicode.IsSpace) >= 0 || strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return false
	}
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || parsed.User != nil || parsed.Fragment != "" {
		return false
	}
	if strings.Contains(parsed.Hostname(), ":") && net.ParseIP(parsed.Hostname()) == nil {
		return false
	}
	if port := parsed.Port(); port != "" {
		number, err := strconv.Atoi(port)
		if err != nil || number < 1 || number > 65535 {
			return false
		}
	}
	return true
}
