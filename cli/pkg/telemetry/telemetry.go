package telemetry

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
)

const (
	intakeURL  = "https://intake.rilldata.io/events/data-modeler-metrics"
	intakeUser = "data-modeler"
	intakeKey  = "lkh8T90ozWJP/KxWnQ81PexRzpdghPdzuB0ly2/86TeUU8q/bKiVug=="
	appName    = "rill-developer"
)

var ErrRillIntake = errors.New("failed to fire telemetry")

// Action represents a type of telemetry event.
// Error actions are not needed. Will be inferred from missing events.
type Action string

const (
	ActionInstallSuccess         Action = "install-success"
	ActionDeployStart            Action = "deploy-start"
	ActionDeploySuccess          Action = "deploy-success"
	ActionGithubConnectedStart   Action = "ghconnected-start"
	ActionGithubConnectedSuccess Action = "ghconnected-success"
	ActionDataAccessStart        Action = "dataaccess-start"
	ActionDataAccessSuccess      Action = "dataaccess-success"
	ActionLoginStart             Action = "login-start"
	ActionLoginSuccess           Action = "login-success"
	ActionAppStart               Action = "app-start"
)

type Telemetry struct {
	Enabled   bool
	InstallID string
	Version   cmdutil.Version
	UserID    string
	events    [][]byte
}

func New(ver cmdutil.Version) *Telemetry {
	installID, enabled, err := dotrill.AnalyticsInfo()
	if err != nil {
		// if there is any error just disable the telemetry.
		// this is simpler than null checking everywhere telemetry methods are called
		enabled = false
	}

	return &Telemetry{
		Enabled:   enabled,
		InstallID: installID,
		Version:   ver,
		UserID:    "",
		events:    make([][]byte, 0),
	}
}

func (t *Telemetry) WithUserID(userID string) {
	t.UserID = userID
}

func (t *Telemetry) Emit(action Action) {
	t.emitBehaviourEvent(string(action), "cli", "terminal", "terminal", nil)
}

func (t *Telemetry) EmitStartEvent(sourceDrivers []string, olapDriver string) {
	payload := map[string]any{"connectors": sourceDrivers, "olap_connector": olapDriver}
	t.emitBehaviourEvent(string(ActionAppStart), "cli", "terminal", "terminal", payload)
}

func (t *Telemetry) Flush(ctx context.Context) error {
	if len(t.events) == 0 {
		return nil
	}

	body := make([]byte, 0)
	for _, event := range t.events {
		body = append(body, event...)
		body = append(body, '\n')
	}

	t.events = make([][]byte, 0)
	return t.emit(ctx, body)
}

type behaviourEventFields struct {
	AppName       string         `json:"app_name"`
	InstallID     string         `json:"install_id"`
	BuildID       string         `json:"build_id"`
	Version       string         `json:"version"`
	UserID        string         `json:"user_id"`
	IsDev         bool           `json:"is_dev"`
	Mode          string         `json:"mode"`
	Action        string         `json:"action"`
	Medium        string         `json:"medium"`
	Space         string         `json:"space"`
	ScreenName    string         `json:"screen_name"`
	EventDatetime int64          `json:"event_datetime"`
	EventType     string         `json:"event_type"`
	Payload       map[string]any `json:"payload"`
}

func (t *Telemetry) emitBehaviourEvent(action, medium, space, screenName string, payload map[string]any) {
	if t == nil || !t.Enabled {
		return
	}

	fields := behaviourEventFields{
		AppName:       appName,
		InstallID:     t.InstallID,
		BuildID:       t.Version.Commit,
		Version:       t.Version.Number,
		UserID:        t.UserID,
		IsDev:         t.Version.IsDev(),
		Mode:          "edit",
		Action:        action,
		Medium:        medium,
		Space:         space,
		ScreenName:    screenName,
		EventDatetime: time.Now().Unix() * 1000,
		EventType:     "behavioral",
		Payload:       payload,
	}
	event, err := json.Marshal(&fields)
	if err != nil {
		return
	}

	t.events = append(t.events, event)
}

func (t *Telemetry) emit(ctx context.Context, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, intakeURL, bytes.NewReader(body))
	if err != nil {
		return ErrRillIntake
	}
	req.Header = http.Header{
		"Authorization": []string{fmt.Sprintf(
			"Basic %s",
			base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprintf("%s:%s", intakeUser, intakeKey)),
			),
		)},
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ErrRillIntake
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ErrRillIntake
	}

	return nil
}
