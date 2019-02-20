package utils

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/carousell/Orion/utils/log"
	newrelic "github.com/newrelic/go-agent"
	"go.elastic.co/apm"
)

const (
	newRelicTransactionID = "NewRelicTransaction"
)

var (
	// NewRelicApp is the reference for newrelic application
	NewRelicApp newrelic.Application
)

//GetHostname fetches the hostname of the system
func GetHostname() string {
	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}
	log.Info(context.Background(), "HOST", host)
	return host
}

//GetNewRelicTransactionFromContext fetches the new relic transaction that is stored in the context
func GetNewRelicTransactionFromContext(ctx context.Context) newrelic.Transaction {
	t := ctx.Value(newRelicTransactionID)
	if t != nil {
		txn, ok := t.(newrelic.Transaction)
		if ok {
			return txn
		}
	}
	return nil
}

//StoreNewRelicTransactionToContext stores a new relic transaction object to context
func StoreNewRelicTransactionToContext(ctx context.Context, t newrelic.Transaction) context.Context {
	return context.WithValue(ctx, newRelicTransactionID, t)
}

//StartNRTransaction starts a new newrelic transaction
func StartNRTransaction(path string, ctx context.Context, req *http.Request, w http.ResponseWriter) context.Context {
	if req == nil {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		req, _ = http.NewRequest("", path, nil)
	}
	if NewRelicApp != nil {
		// check if transaction has been initialized
		t := GetNewRelicTransactionFromContext(ctx)
		if t == nil {
			t := NewRelicApp.StartTransaction(path, w, req)
			ctx = StoreNewRelicTransactionToContext(ctx, t)
		}
	}
	// check if transaction has been initialized
	tx := apm.TransactionFromContext(ctx)
	if tx == nil {
		tx := apm.DefaultTracer.StartTransaction(path, "request")
		ctx = apm.ContextWithTransaction(ctx, tx)
	}
	return ctx
}

//FinishNRTransaction finishes an existing transaction
func FinishNRTransaction(ctx context.Context, err error) {
	t := GetNewRelicTransactionFromContext(ctx)
	if t != nil {
		t.NoticeError(err)
		t.End()
	}
	tx := apm.TransactionFromContext(ctx)
	if tx != nil {
		apm.CaptureError(ctx, err)
		tx.End()
	}
}

//IgnoreNRTransaction ignores this NR trasaction and prevents it from being reported
func IgnoreNRTransaction(ctx context.Context) error {
	t := GetNewRelicTransactionFromContext(ctx)
	if t != nil {
		return t.Ignore()
	}
	return nil
}
