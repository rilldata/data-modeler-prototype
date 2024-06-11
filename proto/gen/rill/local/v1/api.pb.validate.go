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

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
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

	_ = adminv1.GithubPermission(0)
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

// Validate checks the field values on DeployValidationRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *DeployValidationRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DeployValidationRequest with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// DeployValidationRequestMultiError, or nil if none found.
func (m *DeployValidationRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *DeployValidationRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return DeployValidationRequestMultiError(errors)
	}

	return nil
}

// DeployValidationRequestMultiError is an error wrapping multiple validation
// errors returned by DeployValidationRequest.ValidateAll() if the designated
// constraints aren't met.
type DeployValidationRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DeployValidationRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DeployValidationRequestMultiError) AllErrors() []error { return m }

// DeployValidationRequestValidationError is the validation error returned by
// DeployValidationRequest.Validate if the designated constraints aren't met.
type DeployValidationRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DeployValidationRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DeployValidationRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DeployValidationRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DeployValidationRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DeployValidationRequestValidationError) ErrorName() string {
	return "DeployValidationRequestValidationError"
}

// Error satisfies the builtin error interface
func (e DeployValidationRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDeployValidationRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DeployValidationRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DeployValidationRequestValidationError{}

// Validate checks the field values on DeployValidationResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *DeployValidationResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DeployValidationResponse with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// DeployValidationResponseMultiError, or nil if none found.
func (m *DeployValidationResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *DeployValidationResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for IsAuthenticated

	// no validation rules for LoginUrl

	// no validation rules for IsGithubConnected

	// no validation rules for GithubGrantAccessUrl

	// no validation rules for GithubUserName

	// no validation rules for GithubUserPermission

	// no validation rules for GithubOrganizationPermissions

	// no validation rules for IsGithubRepo

	// no validation rules for IsGithubRemoteFound

	// no validation rules for IsGithubRepoAccessGranted

	// no validation rules for GithubUrl

	// no validation rules for RillOrgExistsAsGithubUserName

	// no validation rules for LocalProjectName

	if m.HasUncommittedChanges != nil {
		// no validation rules for HasUncommittedChanges
	}

	if len(errors) > 0 {
		return DeployValidationResponseMultiError(errors)
	}

	return nil
}

// DeployValidationResponseMultiError is an error wrapping multiple validation
// errors returned by DeployValidationResponse.ValidateAll() if the designated
// constraints aren't met.
type DeployValidationResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DeployValidationResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DeployValidationResponseMultiError) AllErrors() []error { return m }

// DeployValidationResponseValidationError is the validation error returned by
// DeployValidationResponse.Validate if the designated constraints aren't met.
type DeployValidationResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DeployValidationResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DeployValidationResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DeployValidationResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DeployValidationResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DeployValidationResponseValidationError) ErrorName() string {
	return "DeployValidationResponseValidationError"
}

// Error satisfies the builtin error interface
func (e DeployValidationResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDeployValidationResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DeployValidationResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DeployValidationResponseValidationError{}

// Validate checks the field values on PushToGithubRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *PushToGithubRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on PushToGithubRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// PushToGithubRequestMultiError, or nil if none found.
func (m *PushToGithubRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *PushToGithubRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Account

	// no validation rules for Repo

	if len(errors) > 0 {
		return PushToGithubRequestMultiError(errors)
	}

	return nil
}

// PushToGithubRequestMultiError is an error wrapping multiple validation
// errors returned by PushToGithubRequest.ValidateAll() if the designated
// constraints aren't met.
type PushToGithubRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m PushToGithubRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m PushToGithubRequestMultiError) AllErrors() []error { return m }

// PushToGithubRequestValidationError is the validation error returned by
// PushToGithubRequest.Validate if the designated constraints aren't met.
type PushToGithubRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e PushToGithubRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e PushToGithubRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e PushToGithubRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e PushToGithubRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e PushToGithubRequestValidationError) ErrorName() string {
	return "PushToGithubRequestValidationError"
}

// Error satisfies the builtin error interface
func (e PushToGithubRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sPushToGithubRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = PushToGithubRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = PushToGithubRequestValidationError{}

// Validate checks the field values on PushToGithubResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *PushToGithubResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on PushToGithubResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// PushToGithubResponseMultiError, or nil if none found.
func (m *PushToGithubResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *PushToGithubResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for GithubUrl

	// no validation rules for Account

	// no validation rules for Repo

	if len(errors) > 0 {
		return PushToGithubResponseMultiError(errors)
	}

	return nil
}

// PushToGithubResponseMultiError is an error wrapping multiple validation
// errors returned by PushToGithubResponse.ValidateAll() if the designated
// constraints aren't met.
type PushToGithubResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m PushToGithubResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m PushToGithubResponseMultiError) AllErrors() []error { return m }

// PushToGithubResponseValidationError is the validation error returned by
// PushToGithubResponse.Validate if the designated constraints aren't met.
type PushToGithubResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e PushToGithubResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e PushToGithubResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e PushToGithubResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e PushToGithubResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e PushToGithubResponseValidationError) ErrorName() string {
	return "PushToGithubResponseValidationError"
}

// Error satisfies the builtin error interface
func (e PushToGithubResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sPushToGithubResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = PushToGithubResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = PushToGithubResponseValidationError{}

// Validate checks the field values on DeployRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *DeployRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DeployRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in DeployRequestMultiError, or
// nil if none found.
func (m *DeployRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *DeployRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Org

	// no validation rules for ProjectName

	// no validation rules for Upload

	if len(errors) > 0 {
		return DeployRequestMultiError(errors)
	}

	return nil
}

// DeployRequestMultiError is an error wrapping multiple validation errors
// returned by DeployRequest.ValidateAll() if the designated constraints
// aren't met.
type DeployRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DeployRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DeployRequestMultiError) AllErrors() []error { return m }

// DeployRequestValidationError is the validation error returned by
// DeployRequest.Validate if the designated constraints aren't met.
type DeployRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DeployRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DeployRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DeployRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DeployRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DeployRequestValidationError) ErrorName() string { return "DeployRequestValidationError" }

// Error satisfies the builtin error interface
func (e DeployRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDeployRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DeployRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DeployRequestValidationError{}

// Validate checks the field values on DeployResponse with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *DeployResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DeployResponse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in DeployResponseMultiError,
// or nil if none found.
func (m *DeployResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *DeployResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for DeployId

	// no validation rules for Org

	// no validation rules for Project

	// no validation rules for FrontendUrl

	if len(errors) > 0 {
		return DeployResponseMultiError(errors)
	}

	return nil
}

// DeployResponseMultiError is an error wrapping multiple validation errors
// returned by DeployResponse.ValidateAll() if the designated constraints
// aren't met.
type DeployResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DeployResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DeployResponseMultiError) AllErrors() []error { return m }

// DeployResponseValidationError is the validation error returned by
// DeployResponse.Validate if the designated constraints aren't met.
type DeployResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DeployResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DeployResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DeployResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DeployResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DeployResponseValidationError) ErrorName() string { return "DeployResponseValidationError" }

// Error satisfies the builtin error interface
func (e DeployResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDeployResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DeployResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DeployResponseValidationError{}
