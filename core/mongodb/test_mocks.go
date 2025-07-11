package mongodb

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.oease.dev/goe/v2/contract"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TestMockConfig is a mock implementation of the Config interface
type TestMockConfig struct {
	mock.Mock
}

func (m *TestMockConfig) Get(key string) any {
	args := m.Called(key)
	return args.Get(0)
}

func (m *TestMockConfig) GetString(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *TestMockConfig) GetInt(key string) int {
	args := m.Called(key)
	return args.Int(0)
}

func (m *TestMockConfig) GetInt64(key string) int64 {
	args := m.Called(key)
	return args.Get(0).(int64)
}

func (m *TestMockConfig) GetFloat64(key string) float64 {
	args := m.Called(key)
	return args.Get(0).(float64)
}

func (m *TestMockConfig) GetBool(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *TestMockConfig) GetDuration(key string) time.Duration {
	args := m.Called(key)
	return args.Get(0).(time.Duration)
}

func (m *TestMockConfig) GetStringSlice(key string) []string {
	args := m.Called(key)
	return args.Get(0).([]string)
}

func (m *TestMockConfig) Has(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

// TestMockLogger is a mock implementation of the Logger interface
type TestMockLogger struct {
	mock.Mock
}

func (m *TestMockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Fatal(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Panic(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Panicf(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *TestMockLogger) Debugf(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *TestMockLogger) Infof(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *TestMockLogger) Warnf(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *TestMockLogger) Errorf(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *TestMockLogger) Fatalf(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *TestMockLogger) Debugw(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Infow(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Warnw(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Errorw(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) With(keysAndValues ...interface{}) contract.Logger {
	args := m.Called(keysAndValues)
	return args.Get(0).(contract.Logger)
}

func (m *TestMockLogger) WithContext(ctx context.Context) contract.Logger {
	args := m.Called(ctx)
	return args.Get(0).(contract.Logger)
}

func (m *TestMockLogger) WithError(err error) contract.Logger {
	args := m.Called(err)
	return args.Get(0).(contract.Logger)
}

func (m *TestMockLogger) GetLogger() *zap.SugaredLogger {
	args := m.Called()
	return args.Get(0).(*zap.SugaredLogger)
}

// TestMockMongoDB implements the contract.MongoDB interface for testing
type TestMockMongoDB struct {
	mock.Mock
}

func (m *TestMockMongoDB) Instance() *mongo.Database {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*mongo.Database)
}

func (m *TestMockMongoDB) Connection(name string) (*mongo.Database, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.Database), args.Error(1)
}

// Removed SetMonitor, IsNoDocumentsError, and Ctx as they are no longer part of the contract

// TestMockMetricsManager implements the contract.MetricsManager interface for testing
type TestMockMetricsManager struct {
	mock.Mock
}

func (m *TestMockMetricsManager) Counter(name string, opts ...contract.MetricOption) contract.Counter {
	args := m.Called(name, opts)
	return args.Get(0).(contract.Counter)
}

func (m *TestMockMetricsManager) Histogram(name string, opts ...contract.MetricOption) contract.Histogram {
	args := m.Called(name, opts)
	return args.Get(0).(contract.Histogram)
}

func (m *TestMockMetricsManager) Gauge(name string, opts ...contract.MetricOption) contract.Gauge {
	args := m.Called(name, opts)
	return args.Get(0).(contract.Gauge)
}

func (m *TestMockMetricsManager) UpDownCounter(name string, opts ...contract.MetricOption) contract.UpDownCounter {
	args := m.Called(name, opts)
	return args.Get(0).(contract.UpDownCounter)
}

func (m *TestMockMetricsManager) GetMeter() metric.Meter {
	args := m.Called()
	return args.Get(0).(metric.Meter)
}

// TestMockTracingManager implements the contract.TracingManager interface for testing
type TestMockTracingManager struct {
	mock.Mock
}

func (m *TestMockTracingManager) StartSpan(ctx context.Context, name string, opts ...contract.SpanOption) (context.Context, contract.Span) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(context.Context), args.Get(1).(contract.Span)
}

func (m *TestMockTracingManager) GetTracer() trace.Tracer {
	args := m.Called()
	return args.Get(0).(trace.Tracer)
}

// TestMockSpan implements the contract.Span interface for testing
type TestMockSpan struct {
	mock.Mock
}

func (m *TestMockSpan) SetAttributes(attrs ...attribute.KeyValue) {
	m.Called(attrs)
}

func (m *TestMockSpan) SetStatus(code codes.Code, description string) {
	m.Called(code, description)
}

func (m *TestMockSpan) RecordError(err error, opts ...trace.EventOption) {
	m.Called(err, opts)
}

func (m *TestMockSpan) AddEvent(name string, opts ...trace.EventOption) {
	m.Called(name, opts)
}

func (m *TestMockSpan) End(opts ...trace.SpanEndOption) {
	m.Called(opts)
}

func (m *TestMockSpan) GetSpan() trace.Span {
	args := m.Called()
	return args.Get(0).(trace.Span)
}

// TestMockCounter implements the contract.Counter interface for testing
type TestMockCounter struct {
	mock.Mock
}

func (m *TestMockCounter) Add(ctx context.Context, value int64, attrs ...attribute.KeyValue) {
	m.Called(ctx, value, attrs)
}

func (m *TestMockCounter) Inc(ctx context.Context, attrs ...attribute.KeyValue) {
	m.Called(ctx, attrs)
}

// TestMockHistogram implements the contract.Histogram interface for testing
type TestMockHistogram struct {
	mock.Mock
}

func (m *TestMockHistogram) Record(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	m.Called(ctx, value, attrs)
}
