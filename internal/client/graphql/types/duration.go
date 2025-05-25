// Package types provides custom types for GraphQL serialization and deserialization
package types

import (
	"bytes"
	"errors"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

// MarshalDuration serializes a time.Duration to a byte slice in ISO8601 format.
func MarshalDuration(t *time.Duration) ([]byte, error) {
	if t == nil {
		return nil, errors.New("duration pointer is nil")
	}

	var buf bytes.Buffer

	graphql.MarshalDuration(*t).MarshalGQL(&buf)

	return buf.Bytes(), nil
}

// UnmarshalDuration deserializes a byte slice in ISO8601 format to a time.Duration.
func UnmarshalDuration(b []byte, t *time.Duration) error {
	unquotedStr, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}

	d, err := graphql.UnmarshalDuration(unquotedStr)
	if err != nil {
		return err
	}

	*t = d

	return nil
}
