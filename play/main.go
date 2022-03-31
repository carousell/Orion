package main

import (
	"context"
	"github.com/carousell/Orion/utils/log"
	"github.com/carousell/Orion/utils/spanutils"
)

func main() {

	ctx := context.Background()
	span, ctx := spanutils.NewDatastoreSpan(ctx, "Span", "www.google.com")

	// call login user.
	span.SetTag("test_tag", "test_tag_value")
	log.Info(ctx, "test", "value")
	defer span.Finish()

}
