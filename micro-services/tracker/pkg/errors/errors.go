package errors

import (
    "fmt"
)

type ErrorType string

const (
    ErrorTypeValidation   ErrorType = "validation"
    ErrorTypeConnection   ErrorType = "connection"
    ErrorTypeSubscription ErrorType = "subscription"
    ErrorTypePublish      ErrorType = "publish"
    ErrorTypeInternal     ErrorType = "internal"
)

type TrackerError struct {
    Type    ErrorType
    Message string
    Err     error
}

func (e *TrackerError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
    }
    return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func NewValidationError(message string, err error) error {
    return &TrackerError{
        Type:    ErrorTypeValidation,
        Message: message,
        Err:     err,
    }
}

func NewConnectionError(message string, err error) error {
    return &TrackerError{
        Type:    ErrorTypeConnection,
        Message: message,
        Err:     err,
    }
}

func NewSubscriptionError(message string, err error) error {
    return &TrackerError{
        Type:    ErrorTypeSubscription,
        Message: message,
        Err:     err,
    }
}

func NewPublishError(message string, err error) error {
    return &TrackerError{
        Type:    ErrorTypePublish,
        Message: message,
        Err:     err,
    }
}
