package strategy

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultStrategy(t *testing.T) {
	duration := time.Millisecond * 50
	s := DefaultStrategy(duration)
	for _, val := range []int{1, 2, 3, 4, 8, 10, 14} {
		assert.Equal(
			t,
			duration,
			s.WaitDuration(val, 20, nil, nil, nil),
			"Duration returned should always be fixed",
		)
	}
}

func TestExonentialStrategy(t *testing.T) {
	duration := time.Millisecond * 50
	s := ExponentialStrategy(duration)
	values := []struct {
		attempt          int
		expectedDuration time.Duration
	}{
		{0, duration},
		{1, duration},
		{2, duration * 3},
		{3, duration * 7},
		{4, duration * 15},
		{5, duration * 31},
	}
	for _, val := range values {
		assert.Equal(
			t,
			val.expectedDuration,
			s.WaitDuration(val.attempt, 20, nil, nil, nil),
			"Duration returned should grow exponentially",
		)
	}
}
