package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"subconv-next/internal/config"
	"subconv-next/internal/model"
)

type customRuleValidationRequest struct {
	Rule                model.CustomRule `json:"rule"`
	PayloadText         *string          `json:"payload_text,omitempty"`
	AdditionalRulesText *string          `json:"additional_rules_text,omitempty"`
}

type customRuleValidationResponse struct {
	OK    bool                        `json:"ok"`
	Error *customRuleValidationDetail `json:"error,omitempty"`
}

type customRuleValidationDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field"`
	Line    int    `json:"line,omitempty"`
}

// Validate only the submitted draft. This path neither fetches remote rules nor
// changes the saved config, node state or current publication.
func (s *Server) handleValidateCustomRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	if _, err := s.requireWorkspace(r); err != nil {
		handleWorkspaceError(w, err)
		return
	}
	var req customRuleValidationRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if req.AdditionalRulesText != nil {
		lines := strings.Split(strings.ReplaceAll(*req.AdditionalRulesText, "\r\n", "\n"), "\n")
		if s.snapshotConfig().Service.PublicConverter && len(lines) > publicMaxRulePayloadItems {
			writeJSON(w, http.StatusBadRequest, customRuleValidationResponse{Error: &customRuleValidationDetail{
				Code: "INVALID_CUSTOM_RULE", Field: "additional_rules", Message: "直写规则不能超过 10000 行",
			}})
			return
		}
		if err := config.ValidateAdditionalRules(lines); err != nil {
			writeCustomRuleValidationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, customRuleValidationResponse{OK: true})
		return
	}
	payload := req.Rule.Payload
	if req.PayloadText != nil {
		payload = strings.Split(strings.ReplaceAll(*req.PayloadText, "\r\n", "\n"), "\n")
	}
	seed := model.DefaultConfig()
	seed.Render.CustomRules = []model.CustomRule{req.Rule}
	seed.Render.CustomRules[0].Payload = payload
	rule := config.Normalize(seed).Render.CustomRules[0]
	if strings.EqualFold(rule.SourceType, "inline") {
		rule.Payload = payload
	}
	if s.snapshotConfig().Service.PublicConverter && len(payload) > publicMaxRulePayloadItems {
		writeJSON(w, http.StatusBadRequest, customRuleValidationResponse{Error: &customRuleValidationDetail{
			Code: "INVALID_CUSTOM_RULE", Field: "payload", Message: "内联规则不能超过 10000 行",
		}})
		return
	}
	if err := config.ValidateCustomRule(rule); err != nil {
		writeCustomRuleValidationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, customRuleValidationResponse{OK: true})
}

func writeCustomRuleValidationError(w http.ResponseWriter, err error) {
	detail := &customRuleValidationDetail{Code: "INVALID_CUSTOM_RULE", Message: "自定义规则格式不正确", Field: "rule"}
	var fieldError *config.CustomRuleValidationError
	if errors.As(err, &fieldError) {
		detail.Field = fieldError.Field
		detail.Line = fieldError.Line
		detail.Message = fieldError.Message
		if detail.Line > 0 {
			detail.Message = fmt.Sprintf("第 %d 行：%s", detail.Line, detail.Message)
		}
	}
	writeJSON(w, http.StatusBadRequest, customRuleValidationResponse{OK: false, Error: detail})
}
