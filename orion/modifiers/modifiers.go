package modifiers

import (
	"context"

	"github.com/carousell/Orion/utils/options"
)

const (
	Request_HTTP = "OrionRequestHTTP"
	Request_gRPC = "OrionRequestGRPC"
	serializeOut = "SerializeOut"
	JSON         = "JSON"
	JSONPB       = "JSONPB"
	ProtoBuf     = "PROTO"
	IgnoreError  = "IGNORE_ERROR"
)

// SerializeOutJSON forces the output to be json.Marshal for http request
func SerializeOutJSON(ctx context.Context) {
	options.AddToOptions(ctx, serializeOut, JSON)
}

// SerializeOutJSONPB forces the output to be jsonpb.Marshal for http request
func SerializeOutJSONPB(ctx context.Context) {
	options.AddToOptions(ctx, serializeOut, JSONPB)
}

// SerializeOutProtoBuf forces the output to be protobuf binary for http request
func SerializeOutProtBuf(ctx context.Context) {
	options.AddToOptions(ctx, serializeOut, ProtoBuf)
}

// GetSerializationType gets the serialization type for the given response
func GetSerilizationType(ctx context.Context) (string, bool) {
	opt := options.FromContext(ctx)
	val, found := opt.Get(serializeOut)
	if !found {
		return "", false
	}
	return val.(string), true
}

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
