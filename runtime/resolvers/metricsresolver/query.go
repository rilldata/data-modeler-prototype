package metricsresolver

import (
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

type Query struct {
	MetricsView         string      `mapstructure:"metrics_view"`
	Dimensions          []Dimension `mapstructure:"dimensions"`
	Measures            []Measure   `mapstructure:"measures"`
	PivotOn             []string    `mapstructure:"pivot_on"`
	Sort                []Sort      `mapstructure:"sort"`
	TimeRange           *TimeRange  `mapstructure:"time_range"`
	ComparisonTimeRange *TimeRange  `mapstructure:"comparison_time_range"`
	Where               *Expression `mapstructure:"where"`
	Having              *Expression `mapstructure:"having"`
	Limit               *int        `mapstructure:"limit"`
	Offset              *int        `mapstructure:"offset"`
	TimeZone            *string     `mapstructure:"time_zone"`
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
	Count           bool    `mapstructure:"count"`
	CountDistinct   *string `mapstructure:"count_distinct"`
	ComparisonValue *string `mapstructure:"comparison_value"`
	ComparisonDelta *string `mapstructure:"comparison_delta"`
	ComparisonRatio *string `mapstructure:"comparison_ratio"`
}

type Sort struct {
	Name string `mapstructure:"name"`
	Desc bool   `mapstructure:"desc"`
}

type TimeRange struct {
	Start        string    `mapstructure:"start"`
	End          string    `mapstructure:"end"`
	IsoDuration  string    `mapstructure:"iso_duration"`
	IsoOffset    string    `mapstructure:"iso_offset"`
	RoundToGrain TimeGrain `mapstructure:"round_to_grain"`
}

type Expression struct {
	Name      string     `mapstructure:"name"`
	Condition *Condition `mapstructure:"condition"`
	Value     any        `mapstructure:"value"`
	Subquery  *Subquery  `mapstructure:"subquery"`
}

type Condition struct {
	Op    Operator      `mapstructure:"op"`
	Exprs []*Expression `mapstructure:"exprs"`
}

type Subquery struct {
	Dimensions []*Dimension `mapstructure:"dimensions"`
	Measures   []*Measure   `mapstructure:"measures"`
	Sort       []*Sort      `mapstructure:"sort"`
	Where      *Expression  `mapstructure:"where"`
	Having     *Expression  `mapstructure:"having"`
	Limit      *int         `mapstructure:"limit"`
	Offset     *int         `mapstructure:"offset"`
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
	OperatorOr          Operator = "or"
	OperatorAnd         Operator = "and"
	OperatorIn          Operator = "in"
	OperatorNin         Operator = "nin"
	OperatorLike        Operator = "like"
	OperatorNlike       Operator = "nlike"
)

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
	case TimeGrainMillisecond, TimeGrainSecond, TimeGrainMinute, TimeGrainHour, TimeGrainDay, TimeGrainWeek, TimeGrainMonth, TimeGrainQuarter, TimeGrainYear:
		return true
	}
	return false
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
