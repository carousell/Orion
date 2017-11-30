package utils

import (
	"context"
	"log"
	"os"

	newrelic "github.com/newrelic/go-agent"
)

const (
	newRelicTransactionId string = "NewRelicTransaction"
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
