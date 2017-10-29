package main

import (
	"github.com/go-kit/kit/log"
	newrelic "github.com/newrelic/go-agent"
	stdopentracing "github.com/opentracing/opentracing-go"
)

// these are local vars used through the life span of this service
var (
	logger      log.Logger
	tracer      stdopentracing.Tracer
	newrelicApp newrelic.Application
)
