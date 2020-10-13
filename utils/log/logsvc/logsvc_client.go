package logsvc

import (
	"context"
	"errors"
	"fmt"
	"github.com/carousell/Orion/utils/log/loggers"
	logsvc_proto "github.com/carousell/Orion/utils/log/logsvc/pb"
	guuid "github.com/google/uuid"
	"google.golang.org/grpc"
	"time"
)

type LogsvcClient struct {
	client     logsvc_proto.LogSvcClient
	logger     loggers.BaseLogger
	config     *LogSvcConfig
	cancelFunc context.CancelFunc
}

type LogSvcConfig struct {
	LogSvcAddr  string
	SvcName     string
	InstanceId  string
	Environment string
}

func InitLogSvcClient(config *LogSvcConfig, logger loggers.BaseLogger) (*LogsvcClient, error) {
	if config == nil {
		return nil, errors.New("invalid config passed while initializing log-svc ")
	}

	conn, err := grpc.Dial(config.LogSvcAddr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("error establishing connection to log-svc %w", err)
	}

	client := logsvc_proto.NewLogSvcClient(conn)

	logsvcClient := &LogsvcClient{
		logger: logger,
		config: config,
		client: client,
	}

	go logsvcClient.runStreamReceiver()

	return logsvcClient, nil
}

func (l *LogsvcClient) SetLogger(logger loggers.BaseLogger) {
	l.logger = logger
}

func (l *LogsvcClient) runStreamReceiver() error {
	defer func() {
		l.cancelFunc = nil
	}()

	for {
		ctx, cancelFunc := context.WithCancel(context.Background())
		l.cancelFunc = cancelFunc

		stream, err := l.client.ConfigStream(context.Background())
		if err != nil {
			time.Sleep(30 * time.Second)
			continue
		}

		var done chan struct{}
		go func() {
			defer close(done)

			for {
				in, err := stream.Recv()
				if err != nil {
					return
				}
				switch in.GetType() {
				case logsvc_proto.ConfigDownstream_HEALTH_CHECK:
					break
				case logsvc_proto.ConfigDownstream_CONFIG:
					switch in.Config.Level {
					case logsvc_proto.LogLevel_PANIC:
						fallthrough
					case logsvc_proto.LogLevel_ERROR:
						l.logger.SetLevel(loggers.ErrorLevel)
					case logsvc_proto.LogLevel_WARNING:
						l.logger.SetLevel(loggers.WarnLevel)
					case logsvc_proto.LogLevel_INFO:
						l.logger.SetLevel(loggers.InfoLevel)
					case logsvc_proto.LogLevel_DEBUG:
						l.logger.SetLevel(loggers.DebugLevel)
					}
				}
			}
		}()

		err = l.sendRegisterRequest(stream)
		if err != nil {
			stream.CloseSend()
			time.Sleep(30 * time.Second)
			continue
		}

		select {
		case <-done:
			continue
		case <-ctx.Done():
			stream.CloseSend()
		}
	}
}

func (l *LogsvcClient) sendRegisterRequest(stream logsvc_proto.LogSvc_ConfigStreamClient) error {
	uuid := guuid.New()
	registerReq := logsvc_proto.ConfigUpstream{
		Type: logsvc_proto.ConfigUpstream_REGISTER,
		RegisterRequest: &logsvc_proto.LoggerMeta{
			ServiceName: l.config.SvcName,
			Environment: l.config.Environment,
			Region:      "",
			InstanceId:  uuid.String(),
		},
	}

	return stream.Send(&registerReq)
}

func (l *LogsvcClient) Close() error {
	if l.cancelFunc != nil {
		l.cancelFunc()
		return nil
	}
	return errors.New("no running stream receiver")
}
