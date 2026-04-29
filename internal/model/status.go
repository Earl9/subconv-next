package model

import "time"

type RuntimeStatus struct {
	StartedAt                time.Time
	Running                  bool
	LastRefreshAt            time.Time
	LastSuccessAt            time.Time
	NextRefreshAt            time.Time
	RefreshInterval          int
	Refreshing               bool
	RefreshStage             string
	YAMLExists               bool
	YAMLUpdatedAt            time.Time
	UpstreamSourceCount      int
	NodeCount                int
	NodeNames                []string
	EnabledSubscriptionCount int
	OutputPath               string
	LastError                string
}
