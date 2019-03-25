package errors

import (
	"io"
	"testing"
)

func TestWrap(t *testing.T) {
	var tests = []struct {
		name     string
		err      error
		message  string
		expected string
	}{
		{
			"original error is wrapped",
			io.EOF,
			"read error",
			"read error: EOF",
		},
		{
			"wrapping a wrapped error results in an error wrapped twice",
			Wrap(io.EOF, "read error"),
			"client error",
			"client error: read error: EOF",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := Wrap(tt.err, tt.message)
			if err.Error() != tt.expected {
				t.Errorf("(%+v, %+v): expected %+v, got %+v", tt.err, tt.message, tt.expected, err)
			}

		})
	}
}
