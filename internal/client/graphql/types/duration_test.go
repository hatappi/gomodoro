package types_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/hatappi/gomodoro/internal/client/graphql/types"
)

func TestMarshalDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		{
			name:     "zero duration",
			input:    0,
			expected: `"PT0S"`,
		},
		{
			name:     "one second",
			input:    time.Second,
			expected: `"PT1S"`,
		},
		{
			name:     "one minute",
			input:    time.Minute,
			expected: `"PT1M"`,
		},
		{
			name:     "one hour",
			input:    time.Hour,
			expected: `"PT1H"`,
		},
		{
			name:     "complex duration",
			input:    2*time.Hour + 30*time.Minute + 45*time.Second,
			expected: `"PT2H30M45S"`,
		},
		{
			name:     "milliseconds",
			input:    500 * time.Millisecond,
			expected: `"PT0.5S"`,
		},
		{
			name:     "negative duration",
			input:    -5 * time.Second,
			expected: `"-PT5S"`,
		},
		{
			name:     "microseconds",
			input:    123 * time.Microsecond,
			expected: `"PT0.000123S"`,
		},
		{
			name:     "nanoseconds",
			input:    456 * time.Nanosecond,
			expected: `"PT0.000000456S"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := types.MarshalDuration(&tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if d := cmp.Diff(tt.expected, string(result)); d != "" {
				t.Fatalf("unexpected contents. %s", d)
			}
		})
	}
}

func TestMarshalDuration_NilInput(t *testing.T) {
	t.Parallel()

	result, err := types.MarshalDuration(nil)
	if err == nil {
		t.Fatalf("expected error when calling MarshalDuration with nil input, but got none")
	}
	if result != nil {
		t.Fatalf("expected nil result when error occurs, got: %v", result)
	}
}

func TestUnmarshalDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{
			name:     "zero duration",
			input:    `"PT0S"`,
			expected: 0,
		},
		{
			name:     "one second",
			input:    `"PT1S"`,
			expected: time.Second,
		},
		{
			name:     "one minute",
			input:    `"PT1M"`,
			expected: time.Minute,
		},
		{
			name:     "one hour",
			input:    `"PT1H"`,
			expected: time.Hour,
		},
		{
			name:     "complex duration",
			input:    `"PT2H30M45S"`,
			expected: 2*time.Hour + 30*time.Minute + 45*time.Second,
		},
		{
			name:     "milliseconds",
			input:    `"PT0.5S"`,
			expected: 500 * time.Millisecond,
		},
		{
			name:     "negative duration",
			input:    `"-PT5S"`,
			expected: -5 * time.Second,
		},
		{
			name:     "microseconds",
			input:    `"PT0.000123S"`,
			expected: 123 * time.Microsecond,
		},
		{
			name:     "nanoseconds",
			input:    `"PT0.000000456S"`,
			expected: 456 * time.Nanosecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var result time.Duration
			err := types.UnmarshalDuration([]byte(tt.input), &result)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if d := cmp.Diff(tt.expected, result); d != "" {
				t.Fatalf("unexpected contents. %s", d)
			}
		})
	}
}

func TestUnmarshalDuration_InvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid JSON",
			input: `invalid`,
		},
		{
			name:  "unquoted string",
			input: `PT1S`,
		},
		{
			name:  "invalid ISO8601 format",
			input: `"invalid_format"`,
		},
		{
			name:  "empty string",
			input: `""`,
		},
		{
			name:  "malformed quotes",
			input: `"PT1S`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var result time.Duration
			err := types.UnmarshalDuration([]byte(tt.input), &result)
			if err == nil {
				t.Fatalf("expected error for input %s, but got none", tt.input)
			}
		})
	}
}
