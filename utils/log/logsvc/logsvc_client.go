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
	closeChan  chan bool
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

func (l *LogsvcClient) applyConfig(config *logsvc_proto.LogConfig) {
	switch config.Level {
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

func (l *LogsvcClient) processStream() error {
	l.closeChan = make(chan bool)

	stream, err := l.client.ConfigStream(context.Background())
	if err != nil {
		return err
	}

	errChan := make(chan error, 3)
	go func() {
		for {
			in, err := stream.Recv()
			if err != nil {
				errChan <- err
				return
			}
			switch in.GetType() {
			case logsvc_proto.ConfigDownstream_HEALTH_CHECK:
				break
			case logsvc_proto.ConfigDownstream_CONFIG:
				l.applyConfig(in.Config)
			}
		}
	}()

	err = l.sendRegisterRequest(stream)
	if err != nil {
		stream.CloseSend()
		return err
	}

	select {
	case err = <-errChan:
		return err
	case <-l.closeChan:
		return stream.CloseSend()
	}
}

func (l *LogsvcClient) runStreamReceiver() error {
	for {
		l.processStream()
		time.Sleep(5 * time.Second)
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
	select {
	case l.closeChan <- true:
	case <-time.After(5 * time.Second):
	}
	return nil
}
