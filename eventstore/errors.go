package eventstore

import (
	"fmt"
)

// ConnectionError defines an error when connecting to an outside service
type ConnectionError struct {
	Source string
	Err    error
}

// NewConnectionError returns an instance of ConnectionError
func NewConnectionError(source string, err error) *ConnectionError {
	return &ConnectionError{
		Source: source,
		Err:    err,
	}
}

// Error returns a string form of the error and implements the error interface
func (c *ConnectionError) Error() string {
	return fmt.Sprintf("failed to connect to %s. %s", c.Source, c.Err)
}

// MarshalError defines an error when marshaling from one type to another
type MarshalError struct {
	From string
	To   string
	Err  error
}

// NewMarshalError creates an instance of MarshalError
func NewMarshalError(from, to string, err error) *MarshalError {
	return &MarshalError{
		From: from,
		To:   to,
		Err:  err,
	}
}

// Error returns a string form of the error and implements the error interface
func (m *MarshalError) Error() string {
	return fmt.Sprintf("failed to marshal %s to %s. %s", m.From, m.To, m.Err)
}

// PutError defines an error when putting data into a database
type PutError struct {
	Source       string
	InsertedType string
	Err          error
}

// NewPutError creates a new PutError
func NewPutError(source, insertedType string, err error) *PutError {
	return &PutError{
		Source:       source,
		InsertedType: insertedType,
		Err:          err,
	}
}

// Error returns a string form of the error and implements the error interface
func (p *PutError) Error() string {
	return fmt.Sprintf("failed to put %s into %s. %s", p.InsertedType, p.Source, p.Err)
}

// InitializationError defines an error when initialzing a client
type InitializationError struct {
	Source string
	Err    error
}

// NewInitializationError returns a new InitializationError
func NewInitializationError(source string, err error) *InitializationError {
	return &InitializationError{
		Source: source,
		Err:    err,
	}
}

// Error returns a string form of the error and implements the error interface
func (i *InitializationError) Error() string {
	return fmt.Sprintf("failed to init %s. %s", i.Source, i.Err)
}
