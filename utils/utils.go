package utils

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	newrelic "github.com/newrelic/go-agent"
)

const (
	newRelicTransactionId string = "NewRelicTransaction"
)

var (
	// NewRelicApp is the reference for newrelic application
	NewRelicApp newrelic.Application
)

func GetHostname() string {
	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}
	log.Println("HOST", host)
	return host
}

func GetNewRelicTransactionFromContext(ctx context.Context) newrelic.Transaction {
	t := ctx.Value(newRelicTransactionId)
	if t != nil {
		txn, ok := t.(newrelic.Transaction)
		if ok {
			return txn
		}
	}
	return nil
}

func StoreNewRelicTransactionToContext(ctx context.Context, t newrelic.Transaction) context.Context {
	return context.WithValue(ctx, newRelicTransactionId, t)
}

func StartNRTransaction(path string, ctx context.Context, req *http.Request, w http.ResponseWriter) context.Context {
	if NewRelicApp != nil {
		// check if transaction has been initialized
		t := GetNewRelicTransactionFromContext(ctx)
		if t == nil {
			if req == nil {
				if !strings.HasPrefix(path, "/") {
					path = "/" + path
				}
				req, _ = http.NewRequest("", path, nil)
			}
			t := NewRelicApp.StartTransaction(path, w, req)
			ctx = StoreNewRelicTransactionToContext(ctx, t)
			return ctx
		}
	}
	return ctx
}

func FinishNRTransaction(ctx context.Context, err error) {
	t := GetNewRelicTransactionFromContext(ctx)
	if t != nil {
		t.NoticeError(err)
		t.End()
	}
}
