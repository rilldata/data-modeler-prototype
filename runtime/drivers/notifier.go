package drivers

import (
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// Notifier sends notifications.
type Notifier interface {
	SendAlertStatus(s *AlertStatus) error
	SendScheduledReport(s *ScheduledReport) error
}

type AlertStatus struct {
	// TODO: Remove ToEmail, ToName once email notifier is created
	ToEmail        string
	ToName         string
	DisplayName    string
	ExecutionTime  time.Time
	Status         runtimev1.AssertionStatus
	IsRecover      bool
	FailRow        map[string]any
	ExecutionError string
	OpenLink       string
	EditLink       string
}

type ScheduledReport struct {
	DisplayName     string
	ReportTime      time.Time
	DownloadFormat  string
	OpenLink        string
	DownloadLink    string
	UnsubscribeLink string
}
