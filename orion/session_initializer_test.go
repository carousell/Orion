package orion

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/carousell/Orion/interceptors"
	kafka "github.com/carousell/go-utils/kafka"
	"google.golang.org/grpc"
)

// ── test doubles ─────────────────────────────────────────────────────────────

type stubServer struct {
	orionCfg Config
}

func (s *stubServer) GetOrionConfig() Config                                     { return s.orionCfg }
func (s *stubServer) GetConfig() map[string]interface{}                          { return nil }
func (s *stubServer) AddInitializers(inits ...Initializer)                       {}
func (s *stubServer) Start()                                                     {}
func (s *stubServer) Stop(timeout time.Duration) error                           { return nil }
func (s *stubServer) Wait() error                                                { return nil }
func (s *stubServer) RegisterService(sd *grpc.ServiceDesc, sf interface{}) error { return nil }

// capturingProducer records the last topic passed to Produce.
type capturingProducer struct {
	lastTopic string
}

func (p *capturingProducer) Run() {}
func (p *capturingProducer) Produce(_ context.Context, topic string, _ []byte, _ []byte) error {
	p.lastTopic = topic
	return nil
}

// errorProducer returns a fixed error from Produce.
type errorProducer struct{ err error }

func (p *errorProducer) Run() {}
func (p *errorProducer) Produce(_ context.Context, _ string, _ []byte, _ []byte) error {
	return p.err
}

// blockingProducer blocks Produce until release is closed.
type blockingProducer struct {
	release chan struct{}
}

func (p *blockingProducer) Run() {}
func (p *blockingProducer) Produce(_ context.Context, _ string, _ []byte, _ []byte) error {
	<-p.release
	return nil
}

// ── factory helpers ───────────────────────────────────────────────────────────

func factoryFor(p rawProducer) func([]string, ...kafka.Option) (rawProducer, error) {
	return func(_ []string, _ ...kafka.Option) (rawProducer, error) { return p, nil }
}

func errorFactory(err error) func([]string, ...kafka.Option) (rawProducer, error) {
	return func(_ []string, _ ...kafka.Option) (rawProducer, error) { return nil, err }
}

func blockingFactory(release chan struct{}) func([]string, ...kafka.Option) (rawProducer, error) {
	return func(_ []string, _ ...kafka.Option) (rawProducer, error) {
		<-release
		return &capturingProducer{}, nil
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func resetSessionGlobals(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		interceptors.SetGlobalSessionActivityProducer(nil)
		interceptors.SetGlobalSessionServiceName("")
		interceptors.SetGlobalSessionActivityTopic(interceptors.DefaultSessionActivityTopic)
	})
}

func newTestInitializer(factory func([]string, ...kafka.Option) (rawProducer, error), initTimeout time.Duration) *sessionInitializer {
	return &sessionInitializer{newProducer: factory, initTimeout: initTimeout}
}

// ── sessionProducerAdapter tests ──────────────────────────────────────────────

func TestAdapter_PublishAsync_HappyPath(t *testing.T) {
	resetSessionGlobals(t)
	cp := &capturingProducer{}
	a := &sessionProducerAdapter{producer: cp, defaultTopic: "my-topic", produceTimeout: 50 * time.Millisecond}
	if err := a.PublishAsync("", map[string]string{"k": "v"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cp.lastTopic != "my-topic" {
		t.Errorf("topic = %q, want %q", cp.lastTopic, "my-topic")
	}
}

func TestAdapter_PublishAsync_ExplicitTopicOverridesDefault(t *testing.T) {
	resetSessionGlobals(t)
	cp := &capturingProducer{}
	a := &sessionProducerAdapter{producer: cp, defaultTopic: "default-topic", produceTimeout: 50 * time.Millisecond}
	if err := a.PublishAsync("override-topic", "payload"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cp.lastTopic != "override-topic" {
		t.Errorf("topic = %q, want %q", cp.lastTopic, "override-topic")
	}
}

func TestAdapter_PublishAsync_EmptyTopicFallsBackToDefault(t *testing.T) {
	resetSessionGlobals(t)
	cp := &capturingProducer{}
	a := &sessionProducerAdapter{producer: cp, defaultTopic: "default-topic", produceTimeout: 50 * time.Millisecond}
	if err := a.PublishAsync("", "payload"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cp.lastTopic != "default-topic" {
		t.Errorf("topic = %q, want %q", cp.lastTopic, "default-topic")
	}
}

func TestAdapter_PublishAsync_ProducerErrorSurfaced(t *testing.T) {
	resetSessionGlobals(t)
	sentinel := errors.New("broker refused")
	a := &sessionProducerAdapter{producer: &errorProducer{err: sentinel}, defaultTopic: "t", produceTimeout: 50 * time.Millisecond}
	err := a.PublishAsync("", "event")
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestAdapter_PublishAsync_BadPayloadReturnsMarshalError(t *testing.T) {
	resetSessionGlobals(t)
	a := &sessionProducerAdapter{producer: &capturingProducer{}, defaultTopic: "t", produceTimeout: 50 * time.Millisecond}
	if err := a.PublishAsync("", make(chan int)); err == nil {
		t.Fatal("expected marshal error, got nil")
	}
}

func TestAdapter_PublishAsync_BlockingProducerTimesOut(t *testing.T) {
	resetSessionGlobals(t)
	bp := &blockingProducer{release: make(chan struct{})}
	defer close(bp.release)

	timeout := 80 * time.Millisecond
	a := &sessionProducerAdapter{producer: bp, defaultTopic: "t", produceTimeout: timeout}

	start := time.Now()
	err := a.PublishAsync("", "event")
	elapsed := time.Since(start)

	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error, got %v", err)
	}
	if elapsed < timeout/2 {
		t.Errorf("returned too quickly (%v)", elapsed)
	}
	if elapsed > timeout*3 {
		t.Errorf("took too long (%v)", elapsed)
	}
}

// ── sessionInitializer.Init tests ─────────────────────────────────────────────

func TestSessionInitializer_NoBrokers_IsNoOp(t *testing.T) {
	resetSessionGlobals(t)
	si := &sessionInitializer{}
	svr := &stubServer{orionCfg: Config{
		OrionServerName:       "test-svc",
		SessionTrackingConfig: SessionTrackingConfig{KafkaBrokers: nil},
	}}
	if err := si.Init(svr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if interceptors.GlobalSessionActivityProducer != nil {
		t.Error("expected GlobalSessionActivityProducer to be nil")
	}
}

func TestSessionInitializer_Init_ExplicitTopic(t *testing.T) {
	resetSessionGlobals(t)
	si := newTestInitializer(factoryFor(&capturingProducer{}), 50*time.Millisecond)
	svr := &stubServer{orionCfg: Config{
		OrionServerName: "my-svc",
		SessionTrackingConfig: SessionTrackingConfig{
			KafkaBrokers: []string{"broker:9092"},
			KafkaTopic:   "custom-events",
		},
	}}
	if err := si.Init(svr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if interceptors.GlobalSessionActivityTopic != "custom-events" {
		t.Errorf("topic = %q, want %q", interceptors.GlobalSessionActivityTopic, "custom-events")
	}
	if interceptors.GlobalSessionServiceName != "my-svc" {
		t.Errorf("service = %q, want %q", interceptors.GlobalSessionServiceName, "my-svc")
	}
	if interceptors.GlobalSessionActivityProducer == nil {
		t.Error("expected GlobalSessionActivityProducer to be set")
	}
}

func TestSessionInitializer_Init_DefaultTopicWhenConfigEmpty(t *testing.T) {
	resetSessionGlobals(t)
	si := newTestInitializer(factoryFor(&capturingProducer{}), 50*time.Millisecond)
	svr := &stubServer{orionCfg: Config{
		OrionServerName: "my-svc",
		SessionTrackingConfig: SessionTrackingConfig{
			KafkaBrokers: []string{"broker:9092"},
			KafkaTopic:   "",
		},
	}}
	if err := si.Init(svr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if interceptors.GlobalSessionActivityTopic != interceptors.DefaultSessionActivityTopic {
		t.Errorf("topic = %q, want %q", interceptors.GlobalSessionActivityTopic, interceptors.DefaultSessionActivityTopic)
	}
}

func TestSessionInitializer_Init_EmptyServiceNameBecomesUnknown(t *testing.T) {
	resetSessionGlobals(t)
	si := newTestInitializer(factoryFor(&capturingProducer{}), 50*time.Millisecond)
	svr := &stubServer{orionCfg: Config{
		OrionServerName: "",
		SessionTrackingConfig: SessionTrackingConfig{
			KafkaBrokers: []string{"broker:9092"},
		},
	}}
	if err := si.Init(svr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if interceptors.GlobalSessionServiceName != "unknown-service" {
		t.Errorf("service = %q, want %q", interceptors.GlobalSessionServiceName, "unknown-service")
	}
}

func TestSessionInitializer_Init_ProducerErrorIsNonFatal(t *testing.T) {
	resetSessionGlobals(t)
	si := newTestInitializer(errorFactory(errors.New("connection refused")), 50*time.Millisecond)
	svr := &stubServer{orionCfg: Config{
		OrionServerName: "my-svc",
		SessionTrackingConfig: SessionTrackingConfig{
			KafkaBrokers: []string{"broker:9092"},
		},
	}}
	if err := si.Init(svr); err != nil {
		t.Fatalf("Init must be non-fatal on producer error, got: %v", err)
	}
	if interceptors.GlobalSessionActivityProducer != nil {
		t.Error("expected producer to be nil after error")
	}
}

func TestSessionInitializer_Init_ProducerTimeoutIsNonFatal(t *testing.T) {
	resetSessionGlobals(t)
	release := make(chan struct{})
	defer close(release)

	timeout := 80 * time.Millisecond
	si := newTestInitializer(blockingFactory(release), timeout)
	svr := &stubServer{orionCfg: Config{
		OrionServerName: "my-svc",
		SessionTrackingConfig: SessionTrackingConfig{
			KafkaBrokers: []string{"broker:9092"},
		},
	}}

	start := time.Now()
	if err := si.Init(svr); err != nil {
		t.Fatalf("Init must be non-fatal on timeout, got: %v", err)
	}
	elapsed := time.Since(start)

	if interceptors.GlobalSessionActivityProducer != nil {
		t.Error("expected producer to be nil after timeout")
	}
	if elapsed < timeout/2 {
		t.Errorf("returned too quickly (%v)", elapsed)
	}
	if elapsed > timeout*3 {
		t.Errorf("took too long (%v)", elapsed)
	}
}

// TestSessionInitializer_Init_DefaultFactory covers the nil-newProducer path (production
// default). newProducer == nil causes Init to call the real kafka.NewProducer internally.
// We supply an unreachable broker so the OS returns "connection refused" immediately
// (no goroutine leak, no waiting for Sarama's 30s dial timeout).
// initTimeout: 1ms ensures the test never stalls even if the OS is slow.
func TestSessionInitializer_Init_DefaultFactory(t *testing.T) {
	resetSessionGlobals(t)
	si := &sessionInitializer{initTimeout: 1 * time.Millisecond} // newProducer intentionally nil
	svr := &stubServer{orionCfg: Config{
		OrionServerName: "my-svc",
		SessionTrackingConfig: SessionTrackingConfig{
			KafkaBrokers: []string{"localhost:29999"}, // refused immediately by OS
		},
	}}
	if err := si.Init(svr); err != nil {
		t.Fatalf("Init must be non-fatal when factory is nil and broker is unreachable, got: %v", err)
	}
	if interceptors.GlobalSessionActivityProducer != nil {
		t.Error("expected producer to be nil when broker is unreachable")
	}
}

// ── sessionInitializer.ReInit test ────────────────────────────────────────────

func TestSessionInitializer_ReInit_IsNoOp(t *testing.T) {
	resetSessionGlobals(t)
	si := &sessionInitializer{}
	svr := &stubServer{orionCfg: Config{OrionServerName: "test-svc"}}
	if err := si.ReInit(svr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestAdapter_PublishAsync_DefaultProduceTimeout covers the produceTimeout == 0 path.
// The adapter falls back to kafkaProduceTimeout when produceTimeout is not set.
// A non-blocking producer returns before the 5s timer fires, so the test is instant.
func TestAdapter_PublishAsync_DefaultProduceTimeout(t *testing.T) {
	resetSessionGlobals(t)
	cp := &capturingProducer{}
	a := &sessionProducerAdapter{producer: cp, defaultTopic: "t"} // produceTimeout intentionally 0
	if err := a.PublishAsync("", "event"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ── SessionInitializer constructor ────────────────────────────────────────────

func TestSessionInitializer_Constructor_ReturnsInitializer(t *testing.T) {
	i := SessionInitializer()
	if i == nil {
		t.Fatal("expected non-nil Initializer")
	}
	if _, ok := i.(*sessionInitializer); !ok {
		t.Fatalf("expected *sessionInitializer, got %T", i)
	}
}
