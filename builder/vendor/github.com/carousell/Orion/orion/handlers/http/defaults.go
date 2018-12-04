package http

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/carousell/Orion/orion/modifiers"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
)

// DefaultEncoder encodes a HTTP request if none are registered. This encoder
// populates the proto message with URL route variables or fields from a JSON
// body if either are available.
func DefaultEncoder(req *http.Request, r interface{}) error {
	// check and map url params to request
	params := mux.Vars(req)
	if len(params) > 0 {
		mapstructure.Decode(params, r)
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	return deserialize(req.Context(), data, r)
}

func deserialize(ctx context.Context, data []byte, r interface{}) error {
	serType := ContentTypeFromHeaders(ctx)
	switch serType {
	case modifiers.ProtoBuf:
		if protoReq, ok := r.(proto.Message); ok {
			return proto.Unmarshal(data, protoReq)
		}
		fallthrough
	case modifiers.JSONPB:
		if protoReq, ok := r.(proto.Message); ok {
			return jsonpb.UnmarshalString(string(data), protoReq)
		}
		fallthrough
	default:
		return json.Unmarshal(data, r)
	}
}

// DefaultWSUpgrader upgrades a websocket if none are registered.
func DefaultWSUpgrader(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*websocket.Conn, error) {
	up := websocket.Upgrader{
		HandshakeTimeout: time.Second * 2,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
	}
	return up.Upgrade(w, r, responseHeader)
}
