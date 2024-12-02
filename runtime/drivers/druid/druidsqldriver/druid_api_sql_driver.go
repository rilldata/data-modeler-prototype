package druidsqldriver

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/pkg/urlutil"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/druid/retrier"
)

var (
	// non-retryable HTTP errors
	// retryable Druid errors
	errCoordinatorDown = regexp.MustCompile("A leader node could not be found for") // HTTP 500
	errBrokerDown      = regexp.MustCompile("There are no available brokers")       // HTTP 500
	errNoObject        = regexp.MustCompile("Object '.*' not found")                // HTTP 400
)

const (
	SkipEmptyBucketsContextKey           = "skipEmptyBuckets"
	sqlQueryIDContextKey                 = "sqlQueryId"
	enableTimeBoundaryPlanningContextKey = "enableTimeBoundaryPlanning"
)

type druidSQLDriver struct{}

var _ driver.Driver = &druidSQLDriver{}

func (a *druidSQLDriver) Open(dsn string) (driver.Conn, error) {
	client := http.Client{}

	return &sqlConnection{
		client: &client,
		dsn:    dsn,
	}, nil
}

func init() {
	sql.Register("druid", &druidSQLDriver{})
}

type sqlConnection struct {
	client *http.Client
	dsn    string
}

var _ driver.QueryerContext = &sqlConnection{}

func (c *sqlConnection) Prepare(query string) (driver.Stmt, error) {
	return &stmt{
		query: query,
		conn:  c,
	}, nil
}

func (c *sqlConnection) Close() error {
	return nil
}

func (c *sqlConnection) Begin() (driver.Tx, error) {
	return nil, fmt.Errorf("unsupported")
}

func (c *sqlConnection) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	// total sum is 126 seconds (sum(2*2^x) from 0 to 5 inclusive)
	re := retrier.NewRetrier(6, 2*time.Second, &coordinatorHTTPCheck{
		c: c,
	})
	return re.RunCtx(ctx, func(ctx context.Context) (driver.Rows, retrier.Action, error) {
		dr := newDruidRequest(query, args)
		b, err := json.Marshal(dr)
		if err != nil {
			return nil, retrier.Fail, err
		}

		bodyReader := bytes.NewReader(b)

		context.AfterFunc(ctx, func() {
			tctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			r, err := http.NewRequestWithContext(tctx, http.MethodDelete, urlutil.MustJoinURL(c.dsn, dr.Context[sqlQueryIDContextKey].(string)), http.NoBody)
			if err != nil {
				return
			}

			resp, err := c.client.Do(r)
			if err != nil {
				return
			}
			resp.Body.Close()
		})

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.dsn, bodyReader)
		if err != nil {
			if strings.Contains(err.Error(), c.dsn) { // avoid returning the actual DSN with the password which will be logged
				return nil, retrier.Fail, fmt.Errorf("%s", strings.ReplaceAll(err.Error(), c.dsn, "<masked>"))
			}
			return nil, retrier.Fail, err
		}

		req.Header.Add("Content-Type", "application/json")
		resp, err := c.client.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), c.dsn) { // avoid returning the actual DSN with the password which will be logged
				return nil, retrier.Fail, fmt.Errorf("%s", strings.ReplaceAll(err.Error(), c.dsn, "<masked>"))
			}
			return nil, retrier.Fail, err
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			return nil, retrier.Retry, fmt.Errorf("Too many requests")
		}

		// Druid sends well-formed response for 200, 400 and 500 status codes, for others use this
		// ref - https://druid.apache.org/docs/latest/api-reference/sql-api/#responses
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError {
			resp.Body.Close()
			return nil, retrier.Fail, fmt.Errorf("unexpected status code: %d, status: %s", resp.StatusCode, resp.Status)
		}

		dec := json.NewDecoder(resp.Body)

		var obj any
		err = dec.Decode(&obj)
		if err != nil {
			resp.Body.Close()
			return nil, retrier.Fail, err
		}
		switch v := obj.(type) {
		case map[string]any:
			resp.Body.Close()
			a := retrier.Fail
			if em, ok := v["errorMessage"].(string); ok {
				if errCoordinatorDown.MatchString(em) || errBrokerDown.MatchString(em) {
					a = retrier.Retry
				} else if errNoObject.MatchString(em) {
					// if a table doesn't exist then it can be a restarting Coordinator
					// note: there's still can be a restarting historical node that cannot be identifed by error messages
					a = retrier.AdditionalCheck
				}
			}
			// example:
			// 500: Unable to parse the SQL
			return nil, a, fmt.Errorf("%v", obj)
		case []any:
			columns := toStringArray(v)
			err = dec.Decode(&obj)
			if err != nil {
				resp.Body.Close()
				return nil, retrier.Fail, err
			}

			types := toStringArray(obj.([]any))

			transformers := make([]func(any) (any, error), len(columns))
			for i, c := range types {
				transformers[i] = createTransformer(c)
			}

			druidRows := &druidRows{
				closer:        resp.Body,
				dec:           dec,
				columns:       columns,
				types:         types,
				transformers:  transformers,
				currentValues: make([]any, len(columns)),
			}
			return druidRows, retrier.Succeed, nil
		default:
			resp.Body.Close()
			return nil, retrier.Fail, fmt.Errorf("unexpected response: %v", obj)
		}
	})
}

func createTransformer(columnType string) func(any) (any, error) {
	switch columnType {
	case "TINYINT":
		return func(v any) (any, error) {
			if f, ok := v.(float64); ok {
				return int8(f), nil
			}
			return v, nil
		}
	case "SMALLINT":
		return func(v any) (any, error) {
			if f, ok := v.(float64); ok {
				return int16(f), nil
			}
			return v, nil
		}
	case "INTEGER":
		return func(v any) (any, error) {
			if f, ok := v.(float64); ok {
				return int32(f), nil
			}
			return v, nil
		}
	case "BIGINT":
		return func(v any) (any, error) {
			if f, ok := v.(float64); ok {
				return int64(f), nil
			}
			return v, nil
		}
	case "FLOAT":
		return func(v any) (any, error) {
			switch v := v.(type) {
			case float64:
				return float32(v), nil
			case string:
				return strconv.ParseFloat(v, 32)
			default:
				return v, nil
			}
		}
	case "DOUBLE", "REAL", "DECIMAL":
		return func(v any) (any, error) {
			if s, ok := v.(string); ok {
				return strconv.ParseFloat(s, 64)
			}
			return v, nil
		}
	case "TIMESTAMP":
		return func(v any) (any, error) {
			if s, ok := v.(string); ok {
				return time.Parse(time.RFC3339, s)
			}
			return v, nil
		}
	case "ARRAY":
		return func(v any) (any, error) {
			if s, ok := v.(string); ok {
				var l []any
				err := json.Unmarshal([]byte(s), &l)
				return l, err
			}
			return v, nil
		}
	case "OTHER":
		return func(v any) (any, error) {
			if s, ok := v.(string); ok {
				var l map[string]any
				err := json.Unmarshal([]byte(s), &l)
				return l, err
			}
			return v, nil
		}
	default:
		return identityTransformer
	}
}

type druidRows struct {
	closer        io.ReadCloser
	dec           *json.Decoder
	columns       []string
	types         []string
	transformers  []func(any) (any, error)
	currentValues []any
}

var (
	_ driver.Rows                           = &druidRows{}
	_ driver.RowsColumnTypeDatabaseTypeName = &druidRows{}
)

func (dr *druidRows) Columns() []string {
	return dr.columns
}

func (dr *druidRows) Close() error {
	return dr.closer.Close()
}

func (dr *druidRows) Next(dest []driver.Value) error {
	err := dr.dec.Decode(&dr.currentValues)
	if err != nil {
		return err
	}

	for i, v := range dr.currentValues {
		v, err := dr.transformers[i](v)
		if err != nil {
			return err
		}

		dest[i] = v
	}

	return nil
}

func (dr *druidRows) ColumnTypeScanType(index int) reflect.Type {
	switch dr.types[index] {
	case "BOOLEAN":
		return reflect.TypeOf(true)
	case "TINYINT":
		return reflect.TypeOf(int8(0))
	case "SMALLINT":
		return reflect.TypeOf(int16(0))
	case "INTEGER":
		return reflect.TypeOf(int32(0))
	case "BIGINT":
		return reflect.TypeOf(int64(0))
	case "FLOAT":
		return reflect.TypeOf(float32(0))
	case "DOUBLE":
		return reflect.TypeOf(float64(0))
	case "REAL":
		return reflect.TypeOf(float64(0))
	case "DECIMAL":
		return reflect.TypeOf(float64(0))
	case "CHAR":
		return reflect.TypeOf("")
	case "VARCHAR":
		return reflect.TypeOf("")
	case "TIMESTAMP":
		return reflect.TypeOf(time.Time{})
	case "DATE":
		return reflect.TypeOf(time.Time{})
	case "OTHER":
		return reflect.TypeOf("")
	}
	return nil
}

func (dr *druidRows) ColumnTypeDatabaseTypeName(index int) string {
	return dr.types[index]
}

type stmt struct {
	conn  *sqlConnection
	query string
}

func (s *stmt) Close() error {
	return nil
}

func (s *stmt) NumInput() int {
	return 0
}

type coordinatorHTTPCheck struct {
	c *sqlConnection
}

var _ retrier.AdditionalTest = &coordinatorHTTPCheck{}

// isHardFailure is called when the previous error doesn't say explicitly if the issue with a datasource or the coordinator.
// If Coordinator is down for a transient reason it's not a hard failure.
// For example, the previous request can return `no such table 'A'`, then isHardFailure checks
// a) if the coordinator is OK -> hard-failure - the table 'A' definitely doesn't exist
// b) if the coordinator has a transient error -> not a hard-failure - the table 'A' can exist
// c) if the coordinator returns not a transient error (ie access-denied) -> hard-failure - we shouldn't wait until the configuration is changed by someone
func (chc *coordinatorHTTPCheck) IsHardFailure(ctx context.Context) (bool, error) {
	dr := newDruidRequest("SELECT * FROM sys.segments LIMIT 1", nil)
	b, err := json.Marshal(dr)
	if err != nil {
		return false, err
	}

	bodyReader := bytes.NewReader(b)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, chc.c.dsn, bodyReader)
	if err != nil {
		return false, err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := chc.c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		return false, nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return true, fmt.Errorf("Unauthorized request")
	}

	// Druid sends well-formed response for 200, 400 and 500 status codes, for others use this
	// ref - https://druid.apache.org/docs/latest/api-reference/sql-api/#responses
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError {
		resp.Body.Close()
		return true, fmt.Errorf("unexpected status code: %d, status: %s", resp.StatusCode, resp.Status)
	}

	dec := json.NewDecoder(resp.Body)

	var obj any
	err = dec.Decode(&obj)
	if err != nil {
		return false, err
	}
	switch v := obj.(type) {
	case map[string]any:
		if em, ok := v["errorMessage"].(string); ok && errCoordinatorDown.MatchString(em) {
			return false, nil
		}
		return true, nil
	case []any:
		return true, nil
	default:
		return true, fmt.Errorf("unexpected response: %v", obj)
	}
}

type DruidParameter struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type DruidRequest struct {
	Query          string           `json:"query"`
	Header         bool             `json:"header"`
	SQLTypesHeader bool             `json:"sqlTypesHeader"`
	ResultFormat   string           `json:"resultFormat"`
	Parameters     []DruidParameter `json:"parameters"`
	Context        map[string]any   `json:"context"`
}

func newDruidRequest(query string, args []driver.NamedValue) *DruidRequest {
	queryCtx := make(map[string]any)
	queryCtx[sqlQueryIDContextKey] = uuid.New().String()
	queryCtx[enableTimeBoundaryPlanningContextKey] = true

	var parameters []DruidParameter
	for _, arg := range args {
		if strings.HasPrefix(arg.Name, drivers.DialectDruid.ContextKeyArgPrefix()) {
			queryCtx[strings.TrimPrefix(arg.Name, drivers.DialectDruid.ContextKeyArgPrefix())] = arg.Value
			continue
		}

		parameters = append(parameters, DruidParameter{
			Type:  toType(arg.Value),
			Value: arg.Value,
		})
	}

	return &DruidRequest{
		Query:          query,
		Header:         true,
		SQLTypesHeader: true,
		ResultFormat:   "arrayLines",
		Parameters:     parameters,
		Context:        queryCtx,
	}
}

func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, fmt.Errorf("unsupported")
}

func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	return nil, fmt.Errorf("unsupported")
}

func identityTransformer(v any) (any, error) {
	return v, nil
}

func toStringArray(values []any) []string {
	s := make([]string, len(values))
	for i, v := range values {
		vv, _ := v.(string)
		s[i] = vv
	}
	return s
}

func toType(v any) string {
	switch v.(type) {
	case int:
		return "INTEGER"
	case float64:
		return "DOUBLE"
	case bool:
		return "BOOLEAN"
	default:
		return "VARCHAR"
	}
}
