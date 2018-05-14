package retry

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMaxRetry(t *testing.T) {
	r := NewRetry(WithMaxRetry(5))
	tests := []struct {
		value  int
		result bool
	}{
		{1, true},
		{2, true},
		{3, true},
		{4, true},
		{5, false},
		{9, false},
	}
	req, err := http.NewRequest("GET", "/some-path", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range tests {
		assert.Equal(
			t,
			test.result,
			r.ShouldRetry(test.value, req, nil, nil),
			"ShouldRetry vaules should match",
		)
	}
}

func TestRetryMethods(t *testing.T) {
	r := NewRetry(WithRetryMethods(http.MethodGet, http.MethodHead))
	tests := []struct {
		method string
		result bool
	}{
		{http.MethodGet, true},
		{http.MethodGet, true},
		{http.MethodPost, false},
		{http.MethodPut, false},
	}
	// build a fake request
	req, err := http.NewRequest("GET", "/some-path", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range tests {
		req.Method = test.method
		assert.Equal(
			t,
			test.result,
			r.ShouldRetry(1, req, nil, nil),
			"WithRetryMethods should work",
		)
	}
}

func TestRetryAllMethods(t *testing.T) {
	r := NewRetry(
		WithRetryMethods(http.MethodGet, http.MethodHead),
		WithRetryAllMethods(true),
	)
	tests := []struct {
		method string
		result bool
	}{
		{http.MethodGet, true},
		{http.MethodGet, true},
		{http.MethodPost, true},
		{http.MethodPut, true},
	}
	// build a fake request
	req, err := http.NewRequest("GET", "/some-path", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range tests {
		req.Method = test.method
		assert.Equal(
			t,
			test.result,
			r.ShouldRetry(1, req, nil, nil),
			"RetryAllMethods should be set",
		)
	}
}

type stra struct {
	called   bool
	duration time.Duration
}

func (s *stra) WaitDuration(retryCount int, maxRetry int, req *http.Request, resp *http.Response, err error) time.Duration {
	s.called = true
	return s.duration
}

func TestWithStrategy(t *testing.T) {
	s := &stra{}
	s.called = false
	r := NewRetry(WithStrategy(s))
	tests := []struct {
		result time.Duration
	}{
		{time.Millisecond},
		{time.Millisecond * 100},
		{time.Second},
		{time.Second * 30},
	}
	for _, test := range tests {
		s.called = false
		s.duration = test.result
		assert.Equal(
			t,
			test.result,
			r.WaitDuration(1, nil, nil, nil),
			"WaitDuration vaules should match",
		)
		assert.True(
			t,
			s.called,
			"WaitDuration should be called",
		)
	}
}

func TestServerError(t *testing.T) {
	tests := []struct {
		code   int
		result bool
	}{
		{http.StatusOK, false},
		{http.StatusInternalServerError, true},
		{http.StatusTooManyRequests, false},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
	}
	r := NewRetry()
	// build a fake request
	req, err := http.NewRequest("GET", "/some-path", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp := &http.Response{}
	for _, test := range tests {
		resp.StatusCode = test.code
		assert.Equal(
			t,
			test.result,
			r.ShouldRetry(1, req, resp, nil),
			"Should not retry on server error for code %d",
			test.code,
		)
	}
}
