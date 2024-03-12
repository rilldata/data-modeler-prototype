// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: rill/admin/v1/telemetry.proto

package adminv1

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
)

// Validate checks the field values on RecordEventsRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *RecordEventsRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on RecordEventsRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// RecordEventsRequestMultiError, or nil if none found.
func (m *RecordEventsRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *RecordEventsRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	for idx, item := range m.GetEvents() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, RecordEventsRequestValidationError{
						field:  fmt.Sprintf("Events[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, RecordEventsRequestValidationError{
						field:  fmt.Sprintf("Events[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return RecordEventsRequestValidationError{
					field:  fmt.Sprintf("Events[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return RecordEventsRequestMultiError(errors)
	}

	return nil
}

// RecordEventsRequestMultiError is an error wrapping multiple validation
// errors returned by RecordEventsRequest.ValidateAll() if the designated
// constraints aren't met.
type RecordEventsRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m RecordEventsRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m RecordEventsRequestMultiError) AllErrors() []error { return m }

// RecordEventsRequestValidationError is the validation error returned by
// RecordEventsRequest.Validate if the designated constraints aren't met.
type RecordEventsRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e RecordEventsRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e RecordEventsRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e RecordEventsRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e RecordEventsRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e RecordEventsRequestValidationError) ErrorName() string {
	return "RecordEventsRequestValidationError"
}

// Error satisfies the builtin error interface
func (e RecordEventsRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sRecordEventsRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = RecordEventsRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = RecordEventsRequestValidationError{}

// Validate checks the field values on RecordEventsResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *RecordEventsResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on RecordEventsResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// RecordEventsResponseMultiError, or nil if none found.
func (m *RecordEventsResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *RecordEventsResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return RecordEventsResponseMultiError(errors)
	}

	return nil
}

// RecordEventsResponseMultiError is an error wrapping multiple validation
// errors returned by RecordEventsResponse.ValidateAll() if the designated
// constraints aren't met.
type RecordEventsResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m RecordEventsResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m RecordEventsResponseMultiError) AllErrors() []error { return m }

// RecordEventsResponseValidationError is the validation error returned by
// RecordEventsResponse.Validate if the designated constraints aren't met.
type RecordEventsResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e RecordEventsResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e RecordEventsResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e RecordEventsResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e RecordEventsResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e RecordEventsResponseValidationError) ErrorName() string {
	return "RecordEventsResponseValidationError"
}

// Error satisfies the builtin error interface
func (e RecordEventsResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sRecordEventsResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = RecordEventsResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = RecordEventsResponseValidationError{}
