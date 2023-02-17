// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: rill/ui/v1/dashboard.proto

package uiv1

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"

	runtimev1 "github.com/rilldata/rill/rill/runtime/v1"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort

	_ = runtimev1.TimeGrain(0)
)

// Validate checks the field values on DashboardState with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *DashboardState) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DashboardState with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in DashboardStateMultiError,
// or nil if none found.
func (m *DashboardState) ValidateAll() error {
	return m.validate(true)
}

func (m *DashboardState) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.TimeStart != nil {

		if all {
			switch v := interface{}(m.GetTimeStart()).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, DashboardStateValidationError{
						field:  "TimeStart",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, DashboardStateValidationError{
						field:  "TimeStart",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(m.GetTimeStart()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return DashboardStateValidationError{
					field:  "TimeStart",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if m.TimeEnd != nil {

		if all {
			switch v := interface{}(m.GetTimeEnd()).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, DashboardStateValidationError{
						field:  "TimeEnd",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, DashboardStateValidationError{
						field:  "TimeEnd",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(m.GetTimeEnd()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return DashboardStateValidationError{
					field:  "TimeEnd",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if m.TimeGranularity != nil {
		// no validation rules for TimeGranularity
	}

	if m.Filters != nil {

		if all {
			switch v := interface{}(m.GetFilters()).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, DashboardStateValidationError{
						field:  "Filters",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, DashboardStateValidationError{
						field:  "Filters",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(m.GetFilters()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return DashboardStateValidationError{
					field:  "Filters",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return DashboardStateMultiError(errors)
	}

	return nil
}

// DashboardStateMultiError is an error wrapping multiple validation errors
// returned by DashboardState.ValidateAll() if the designated constraints
// aren't met.
type DashboardStateMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DashboardStateMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DashboardStateMultiError) AllErrors() []error { return m }

// DashboardStateValidationError is the validation error returned by
// DashboardState.Validate if the designated constraints aren't met.
type DashboardStateValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DashboardStateValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DashboardStateValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DashboardStateValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DashboardStateValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DashboardStateValidationError) ErrorName() string { return "DashboardStateValidationError" }

// Error satisfies the builtin error interface
func (e DashboardStateValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDashboardState.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DashboardStateValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DashboardStateValidationError{}
