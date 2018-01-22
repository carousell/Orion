package options

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	TestKey := "TestKey"
	TestValue := "TestValue"
	ctx := context.Background()
	// add
	ctx = AddToOptions(ctx, TestKey, TestValue)
	options := FromContext(ctx)
	// fetch
	value, found := options.Get(TestKey)
	assert.True(t, found, "key should be found")
	assert.Equal(t, TestValue, value, "values should be equal")

	//delete
	options.Del(TestKey)
	//fetch
	options2 := FromContext(ctx)
	value, found = options2.Get(TestKey)
	assert.False(t, found, "key should NOT be found")
	assert.NotEqual(t, TestValue, value, "values should NOT be equal")
}
