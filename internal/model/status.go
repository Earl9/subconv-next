package model

import "time"

type RuntimeStatus struct {
	StartedAt                time.Time
	Running                  bool
	LastRefreshAt            time.Time
	LastSuccessAt            time.Time
	NodeCount                int
	EnabledSubscriptionCount int
	OutputPath               string
	LastError                string
}
