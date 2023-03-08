package orion

import (
	"context"
	"strings"

	"github.com/carousell/Orion/v2/utils/errors/notifier"
	"github.com/carousell/Orion/v2/utils/log"
)

var (
	//DefaultInitializers are the initializers applied by orion as default
	DefaultInitializers = []Initializer{
		ErrorLoggingInitializer(),
	}
)

//ErrorLoggingInitializer returns a Initializer implementation for error notifier
func ErrorLoggingInitializer() Initializer {
	return &errorLoggingInitializer{}
}

type errorLoggingInitializer struct{}

func (e *errorLoggingInitializer) Init(svr Server) error {
	env := svr.GetOrionConfig().Env
	// environment for error notification
	notifier.SetEnvironemnt(env)

	// rollbar
	rToken := svr.GetOrionConfig().RollbarToken
	if strings.TrimSpace(rToken) != "" {
		notifier.InitRollbar(rToken, env)
		log.Debug(context.Background(), "rollbarToken", rToken, "env", env)
	}

	//sentry
	sToken := svr.GetOrionConfig().SentryDSN
	if strings.TrimSpace(sToken) != "" {
		notifier.InitSentry(sToken)
		log.Debug(context.Background(), "sentryDSN", rToken, "env", env)
	}
	return nil
}

func (e *errorLoggingInitializer) ReInit(svr Server) error {
	return e.Init(svr)
}
