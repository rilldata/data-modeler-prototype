// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: rill/local/v1/api.proto

package localv1

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

// Validate checks the field values on PingRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *PingRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on PingRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in PingRequestMultiError, or
// nil if none found.
func (m *PingRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *PingRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return PingRequestMultiError(errors)
	}

	return nil
}

// PingRequestMultiError is an error wrapping multiple validation errors
// returned by PingRequest.ValidateAll() if the designated constraints aren't met.
type PingRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m PingRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m PingRequestMultiError) AllErrors() []error { return m }

// PingRequestValidationError is the validation error returned by
// PingRequest.Validate if the designated constraints aren't met.
type PingRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e PingRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e PingRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e PingRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e PingRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e PingRequestValidationError) ErrorName() string { return "PingRequestValidationError" }

// Error satisfies the builtin error interface
func (e PingRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sPingRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = PingRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = PingRequestValidationError{}

// Validate checks the field values on PingResponse with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *PingResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on PingResponse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in PingResponseMultiError, or
// nil if none found.
func (m *PingResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *PingResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if all {
		switch v := interface{}(m.GetTime()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, PingResponseValidationError{
					field:  "Time",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, PingResponseValidationError{
					field:  "Time",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetTime()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return PingResponseValidationError{
				field:  "Time",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return PingResponseMultiError(errors)
	}

	return nil
}

// PingResponseMultiError is an error wrapping multiple validation errors
// returned by PingResponse.ValidateAll() if the designated constraints aren't met.
type PingResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m PingResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m PingResponseMultiError) AllErrors() []error { return m }

// PingResponseValidationError is the validation error returned by
// PingResponse.Validate if the designated constraints aren't met.
type PingResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e PingResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e PingResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e PingResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e PingResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e PingResponseValidationError) ErrorName() string { return "PingResponseValidationError" }

// Error satisfies the builtin error interface
func (e PingResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sPingResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = PingResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = PingResponseValidationError{}

// Validate checks the field values on GetMetadataRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *GetMetadataRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetMetadataRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// GetMetadataRequestMultiError, or nil if none found.
func (m *GetMetadataRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *GetMetadataRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return GetMetadataRequestMultiError(errors)
	}

	return nil
}

// GetMetadataRequestMultiError is an error wrapping multiple validation errors
// returned by GetMetadataRequest.ValidateAll() if the designated constraints
// aren't met.
type GetMetadataRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetMetadataRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetMetadataRequestMultiError) AllErrors() []error { return m }

// GetMetadataRequestValidationError is the validation error returned by
// GetMetadataRequest.Validate if the designated constraints aren't met.
type GetMetadataRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetMetadataRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetMetadataRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetMetadataRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetMetadataRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetMetadataRequestValidationError) ErrorName() string {
	return "GetMetadataRequestValidationError"
}

// Error satisfies the builtin error interface
func (e GetMetadataRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetMetadataRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetMetadataRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetMetadataRequestValidationError{}

// Validate checks the field values on GetMetadataResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *GetMetadataResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetMetadataResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// GetMetadataResponseMultiError, or nil if none found.
func (m *GetMetadataResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *GetMetadataResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for InstanceId

	// no validation rules for ProjectPath

	// no validation rules for InstallId

	// no validation rules for UserId

	// no validation rules for Version

	// no validation rules for BuildCommit

	// no validation rules for BuildTime

	// no validation rules for IsDev

	// no validation rules for AnalyticsEnabled

	// no validation rules for Readonly

	// no validation rules for GrpcPort

	if len(errors) > 0 {
		return GetMetadataResponseMultiError(errors)
	}

	return nil
}

// GetMetadataResponseMultiError is an error wrapping multiple validation
// errors returned by GetMetadataResponse.ValidateAll() if the designated
// constraints aren't met.
type GetMetadataResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetMetadataResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetMetadataResponseMultiError) AllErrors() []error { return m }

// GetMetadataResponseValidationError is the validation error returned by
// GetMetadataResponse.Validate if the designated constraints aren't met.
type GetMetadataResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetMetadataResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetMetadataResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetMetadataResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetMetadataResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetMetadataResponseValidationError) ErrorName() string {
	return "GetMetadataResponseValidationError"
}

// Error satisfies the builtin error interface
func (e GetMetadataResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetMetadataResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetMetadataResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetMetadataResponseValidationError{}

// Validate checks the field values on GetVersionRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *GetVersionRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetVersionRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// GetVersionRequestMultiError, or nil if none found.
func (m *GetVersionRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *GetVersionRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return GetVersionRequestMultiError(errors)
	}

	return nil
}

// GetVersionRequestMultiError is an error wrapping multiple validation errors
// returned by GetVersionRequest.ValidateAll() if the designated constraints
// aren't met.
type GetVersionRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetVersionRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetVersionRequestMultiError) AllErrors() []error { return m }

// GetVersionRequestValidationError is the validation error returned by
// GetVersionRequest.Validate if the designated constraints aren't met.
type GetVersionRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetVersionRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetVersionRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetVersionRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetVersionRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetVersionRequestValidationError) ErrorName() string {
	return "GetVersionRequestValidationError"
}

// Error satisfies the builtin error interface
func (e GetVersionRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetVersionRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetVersionRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetVersionRequestValidationError{}

// Validate checks the field values on GetVersionResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *GetVersionResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetVersionResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// GetVersionResponseMultiError, or nil if none found.
func (m *GetVersionResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *GetVersionResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Current

	// no validation rules for Latest

	if len(errors) > 0 {
		return GetVersionResponseMultiError(errors)
	}

	return nil
}

// GetVersionResponseMultiError is an error wrapping multiple validation errors
// returned by GetVersionResponse.ValidateAll() if the designated constraints
// aren't met.
type GetVersionResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetVersionResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetVersionResponseMultiError) AllErrors() []error { return m }

// GetVersionResponseValidationError is the validation error returned by
// GetVersionResponse.Validate if the designated constraints aren't met.
type GetVersionResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetVersionResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetVersionResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetVersionResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetVersionResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetVersionResponseValidationError) ErrorName() string {
	return "GetVersionResponseValidationError"
}

// Error satisfies the builtin error interface
func (e GetVersionResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetVersionResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetVersionResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetVersionResponseValidationError{}
