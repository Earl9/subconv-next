package config

import (
	"errors"
	"strings"

	"subconv-next/internal/rulevalidation"
)

// ValidateAdditionalRules checks complete rules, including their target policy.
// Empty additional rules are valid because the renderer supplies the defaults.
func ValidateAdditionalRules(rules []string) error {
	var lines []string
	for _, item := range rules {
		lines = append(lines, strings.Split(strings.ReplaceAll(item, "\r\n", "\n"), "\n")...)
	}
	if err := rulevalidation.ValidateRules(lines); err != nil {
		var detail *rulevalidation.Error
		if errors.As(err, &detail) {
			return &CustomRuleValidationError{Field: "additional_rules", Line: detail.Line, Message: detail.Message}
		}
		return &CustomRuleValidationError{Field: "additional_rules", Message: "直写规则格式不正确"}
	}
	return nil
}
