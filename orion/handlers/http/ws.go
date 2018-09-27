package http

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/carousell/Orion/utils/errors"
	"github.com/carousell/Orion/utils/errors/notifier"
	"github.com/carousell/Orion/utils/log"
	"github.com/carousell/Orion/utils/log/loggers"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc/metadata"
)

func (h *httpHandler) getWSHandler(serviceName, methodName string) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		h.wsHandler(resp, req, serviceName, methodName)
	}
}

func (h *httpHandler) wsHandler(resp http.ResponseWriter, req *http.Request, service, method string) {
	var err error
	var ctx context.Context
	defer func(t time.Time) {
		notifier.Notify(err, ctx, req.URL.String())
		log.Info(ctx, "path", req.URL.String(), "duration", time.Since(t), "err", err)
	}(time.Now())
	info, ok := h.mapping.Get(service, method)
	if ok {
		//setup context
		ctx = prepareContext(req, info)
		ctx = processOptions(ctx, req, info)
		ctx = loggers.AddToLogContext(ctx, "transport", "ws")
		req = req.WithContext(ctx)

		notifier.SetTraceId(ctx)
		log.Info(ctx, "path", req.URL.String(), "msg", "new websocket connection")

		// httpHandler allows handling entire http request
		if info.httpHandler != nil {
			req = req.WithContext(ctx)
			if info.httpHandler(resp, req) {
				// short circuit if handler has handled request
				return
			}
		}

		if info.stream == nil {
			log.Error(ctx, "ws", "no stream registered", "url", req.URL.String())
			err = errors.New("No stream registered")
		}

		var con *websocket.Conn
		con, err = DefaultWSUpgrader(resp, req, http.Header{})
		if err != nil {
			log.Error(ctx, "wsUpgrade", "failed", "err", err, "url", req.URL.String())
			return
		}
		defer con.Close()

		//create a cancelable context from request context
		ctx = req.Context()
		streamCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		stream := streamServer{
			ctx: streamCtx,
			con: con,
			han: h,
		}
		// handle the stream
		err = info.stream(info.svc.svc, &stream)
		return
	}
	writeResp(resp, http.StatusNotFound, []byte("Not Found: "+req.URL.String()))
}

type streamServer struct {
	ctx context.Context
	con *websocket.Conn
	han *httpHandler
}

func (s *streamServer) SetHeader(metadata.MD) error {
	return nil
}

func (s *streamServer) SendHeader(metadata.MD) error {
	return nil
}

func (s *streamServer) SetTrailer(metadata.MD) {
}

func (s *streamServer) Context() context.Context {
	return s.ctx
}

func (s *streamServer) SendMsg(m interface{}) error {
	if protoMsg, ok := m.(proto.Message); ok {
		data, contentType, err := s.han.serialize(s.Context(), protoMsg)
		if err != nil {
			return err
		}
		switch contentType {
		case ContentTypeProto:
			return s.con.WriteMessage(websocket.BinaryMessage, data)
		default:
			return s.con.WriteMessage(websocket.TextMessage, data)
		}
	}
	return s.con.WriteJSON(m)
}

func (s *streamServer) RecvMsg(m interface{}) error {
	msgType, data, err := s.con.ReadMessage()
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived, websocket.CloseGoingAway) {
			return io.EOF
		}
		return err
	}
	switch msgType {
	case websocket.TextMessage:
		fallthrough
	case websocket.BinaryMessage:
		return deserialize(s.Context(), data, m)
	case websocket.CloseMessage:
		return io.EOF
	}
	return s.RecvMsg(m)
}
