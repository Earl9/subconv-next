package model

type AuditReport struct {
	RawCount      int            `json:"raw_count"`
	FinalCount    int            `json:"final_count"`
	ExcludedCount int            `json:"excluded_count"`
	ExcludedNodes []ExcludedNode `json:"excluded_nodes,omitempty"`
	Warnings      []AuditWarning `json:"warnings,omitempty"`
}

type ExcludedNode struct {
	ID     string     `json:"id,omitempty"`
	Name   string     `json:"name,omitempty"`
	Source SourceInfo `json:"source,omitempty"`
	Reason string     `json:"reason,omitempty"`
}

type AuditWarning struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}
