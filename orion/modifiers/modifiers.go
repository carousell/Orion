package modifiers

import (
	"context"

	"github.com/carousell/Orion/v2/utils/options"
)

// constatnts for specific serializers
const (
	IgnoreError = "IGNORE_ERROR"
)

//DontLogError makes sure, error returned is not reported to external services
func DontLogError(ctx context.Context) {
	options.AddToOptions(ctx, IgnoreError, true)
}

// HasDontLogError check ifs the error should be reported or not
func HasDontLogError(ctx context.Context) bool {
	opt := options.FromContext(ctx)
	_, found := opt.Get(IgnoreError)
	return found
}
