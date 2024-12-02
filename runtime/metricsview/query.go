package metricsview

import (
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

type Query struct {
	MetricsView         string      `mapstructure:"metrics_view"`
	Dimensions          []Dimension `mapstructure:"dimensions"`
	Measures            []Measure   `mapstructure:"measures"`
	PivotOn             []string    `mapstructure:"pivot_on"`
	Spine               *Spine      `mapstructure:"spine"`
	Sort                []Sort      `mapstructure:"sort"`
	TimeRange           *TimeRange  `mapstructure:"time_range"`
	ComparisonTimeRange *TimeRange  `mapstructure:"comparison_time_range"`
	Where               *Expression `mapstructure:"where"`
	Having              *Expression `mapstructure:"having"`
	Limit               *int64      `mapstructure:"limit"`
	Offset              *int64      `mapstructure:"offset"`
	TimeZone            string      `mapstructure:"time_zone"`
	UseDisplayNames     bool        `mapstructure:"use_display_names"`
}

type Dimension struct {
	Name    string            `mapstructure:"name"`
	Compute *DimensionCompute `mapstructure:"compute"`
}

type DimensionCompute struct {
	TimeFloor *DimensionComputeTimeFloor `mapstructure:"time_floor"`
}

type DimensionComputeTimeFloor struct {
	Dimension string    `mapstructure:"dimension"`
	Grain     TimeGrain `mapstructure:"grain"`
}

type Measure struct {
	Name    string          `mapstructure:"name"`
	Compute *MeasureCompute `mapstructure:"compute"`
}

type MeasureCompute struct {
	Count           bool                           `mapstructure:"count"`
	CountDistinct   *MeasureComputeCountDistinct   `mapstructure:"count_distinct"`
	ComparisonValue *MeasureComputeComparisonValue `mapstructure:"comparison_value"`
	ComparisonDelta *MeasureComputeComparisonDelta `mapstructure:"comparison_delta"`
	ComparisonRatio *MeasureComputeComparisonRatio `mapstructure:"comparison_ratio"`
	PercentOfTotal  *MeasureComputePercentOfTotal  `mapstructure:"percent_of_total"`
	URI             *MeasureComputeURI             `mapstructure:"uri"`
}

func (m *MeasureCompute) Validate() error {
	n := 0
	if m.Count {
		n++
	}
	if m.CountDistinct != nil {
		n++
	}
	if m.ComparisonValue != nil {
		n++
	}
	if m.ComparisonDelta != nil {
		n++
	}
	if m.ComparisonRatio != nil {
		n++
	}
	if m.PercentOfTotal != nil {
		n++
	}
	if m.URI != nil {
		n++
	}
	if n == 0 {
		return fmt.Errorf(`must specify a compute operation`)
	}
	if n > 1 {
		return fmt.Errorf("must specify only one compute operation")
	}
	return nil
}

type MeasureComputeCountDistinct struct {
	Dimension string `mapstructure:"dimension"`
}

type MeasureComputeComparisonValue struct {
	Measure string `mapstructure:"measure"`
}

type MeasureComputeComparisonDelta struct {
	Measure string `mapstructure:"measure"`
}

type MeasureComputeComparisonRatio struct {
	Measure string `mapstructure:"measure"`
}

type MeasureComputePercentOfTotal struct {
	Measure string   `mapstructure:"measure"`
	Total   *float64 `mapstructure:"total"`
}

type MeasureComputeURI struct {
	Dimension string `mapstructure:"dimension"`
}

type Spine struct {
	Where     *WhereSpine `mapstructure:"where"`
	TimeRange *TimeSpine  `mapstructure:"time"`
}

type WhereSpine struct {
	Expression *Expression `mapstructure:"expr"`
}

type TimeSpine struct {
	Start time.Time `mapstructure:"start"`
	End   time.Time `mapstructure:"end"`
	Grain TimeGrain `mapstructure:"grain"`
	Alias string    `mapstructure:"alias"`
}

type Sort struct {
	Name string `mapstructure:"name"`
	Desc bool   `mapstructure:"desc"`
}

type TimeRange struct {
	Start        time.Time `mapstructure:"start"`
	End          time.Time `mapstructure:"end"`
	IsoDuration  string    `mapstructure:"iso_duration"`
	IsoOffset    string    `mapstructure:"iso_offset"`
	RoundToGrain TimeGrain `mapstructure:"round_to_grain"`
}

func (tr *TimeRange) IsZero() bool {
	return tr.Start.IsZero() && tr.End.IsZero() && tr.IsoDuration == "" && tr.IsoOffset == "" && tr.RoundToGrain == TimeGrainUnspecified
}

type Expression struct {
	Name      string     `mapstructure:"name"`
	Value     any        `mapstructure:"val"`
	Condition *Condition `mapstructure:"cond"`
	Subquery  *Subquery  `mapstructure:"subquery"`
}

type Condition struct {
	Operator    Operator      `mapstructure:"op"`
	Expressions []*Expression `mapstructure:"exprs"`
}

type Subquery struct {
	Dimension Dimension   `mapstructure:"dimension"`
	Measures  []Measure   `mapstructure:"measures"`
	Where     *Expression `mapstructure:"where"`
	Having    *Expression `mapstructure:"having"`
}

type Operator string

const (
	OperatorUnspecified Operator = ""
	OperatorEq          Operator = "eq"
	OperatorNeq         Operator = "neq"
	OperatorLt          Operator = "lt"
	OperatorLte         Operator = "lte"
	OperatorGt          Operator = "gt"
	OperatorGte         Operator = "gte"
	OperatorIn          Operator = "in"
	OperatorNin         Operator = "nin"
	OperatorIlike       Operator = "ilike"
	OperatorNilike      Operator = "nilike"
	OperatorOr          Operator = "or"
	OperatorAnd         Operator = "and"
)

func (o Operator) Valid() bool {
	switch o {
	case OperatorEq, OperatorNeq, OperatorLt, OperatorLte, OperatorGt, OperatorGte, OperatorIn, OperatorNin, OperatorIlike, OperatorNilike, OperatorOr, OperatorAnd:
		return true
	}
	return false
}

type TimeGrain string

const (
	TimeGrainUnspecified TimeGrain = ""
	TimeGrainMillisecond TimeGrain = "millisecond"
	TimeGrainSecond      TimeGrain = "second"
	TimeGrainMinute      TimeGrain = "minute"
	TimeGrainHour        TimeGrain = "hour"
	TimeGrainDay         TimeGrain = "day"
	TimeGrainWeek        TimeGrain = "week"
	TimeGrainMonth       TimeGrain = "month"
	TimeGrainQuarter     TimeGrain = "quarter"
	TimeGrainYear        TimeGrain = "year"
)

func (t TimeGrain) Valid() bool {
	switch t {
	case TimeGrainUnspecified, TimeGrainMillisecond, TimeGrainSecond, TimeGrainMinute, TimeGrainHour, TimeGrainDay, TimeGrainWeek, TimeGrainMonth, TimeGrainQuarter, TimeGrainYear:
		return true
	}
	return false
}

func (t TimeGrain) ToTimeutil() timeutil.TimeGrain {
	switch t {
	case TimeGrainUnspecified:
		return timeutil.TimeGrainUnspecified
	case TimeGrainMillisecond:
		return timeutil.TimeGrainMillisecond
	case TimeGrainSecond:
		return timeutil.TimeGrainSecond
	case TimeGrainMinute:
		return timeutil.TimeGrainMinute
	case TimeGrainHour:
		return timeutil.TimeGrainHour
	case TimeGrainDay:
		return timeutil.TimeGrainDay
	case TimeGrainWeek:
		return timeutil.TimeGrainWeek
	case TimeGrainMonth:
		return timeutil.TimeGrainMonth
	case TimeGrainQuarter:
		return timeutil.TimeGrainQuarter
	case TimeGrainYear:
		return timeutil.TimeGrainYear
	default:
		panic(fmt.Errorf("invalid time grain %q", t))
	}
}

func (t TimeGrain) ToProto() runtimev1.TimeGrain {
	switch t {
	case TimeGrainUnspecified:
		return runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED
	case TimeGrainMillisecond:
		return runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND
	case TimeGrainSecond:
		return runtimev1.TimeGrain_TIME_GRAIN_SECOND
	case TimeGrainMinute:
		return runtimev1.TimeGrain_TIME_GRAIN_MINUTE
	case TimeGrainHour:
		return runtimev1.TimeGrain_TIME_GRAIN_HOUR
	case TimeGrainDay:
		return runtimev1.TimeGrain_TIME_GRAIN_DAY
	case TimeGrainWeek:
		return runtimev1.TimeGrain_TIME_GRAIN_WEEK
	case TimeGrainMonth:
		return runtimev1.TimeGrain_TIME_GRAIN_MONTH
	case TimeGrainQuarter:
		return runtimev1.TimeGrain_TIME_GRAIN_QUARTER
	case TimeGrainYear:
		return runtimev1.TimeGrain_TIME_GRAIN_YEAR
	default:
		panic(fmt.Errorf("invalid time grain %q", t))
	}
}

func TimeGrainFromProto(t runtimev1.TimeGrain) TimeGrain {
	switch t {
	case runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED:
		return TimeGrainUnspecified
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return TimeGrainMillisecond
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return TimeGrainSecond
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return TimeGrainMinute
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return TimeGrainHour
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		return TimeGrainDay
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		return TimeGrainWeek
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		return TimeGrainMonth
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		return TimeGrainQuarter
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return TimeGrainYear
	default:
		panic(fmt.Errorf("invalid time grain %q", t))
	}
}
