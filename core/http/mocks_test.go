package http

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// MockConfig implements contract.Config for testing
type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) Get(key string) any {
	args := m.Called(key)
	return args.Get(0)
}

func (m *MockConfig) GetString(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *MockConfig) GetInt(key string) int {
	args := m.Called(key)
	return args.Int(0)
}

func (m *MockConfig) GetInt64(key string) int64 {
	args := m.Called(key)
	return args.Get(0).(int64)
}

func (m *MockConfig) GetFloat64(key string) float64 {
	args := m.Called(key)
	return args.Get(0).(float64)
}

func (m *MockConfig) GetBool(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockConfig) GetDuration(key string) time.Duration {
	args := m.Called(key)
	return args.Get(0).(time.Duration)
}

func (m *MockConfig) GetStringSlice(key string) []string {
	args := m.Called(key)
	return args.Get(0).([]string)
}

func (m *MockConfig) GetStringMap(key string) map[string]any {
	args := m.Called(key)
	return args.Get(0).(map[string]any)
}

func (m *MockConfig) Set(key string, value any) {
	m.Called(key, value)
}

func (m *MockConfig) Has(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockConfig) All() map[string]any {
	args := m.Called()
	return args.Get(0).(map[string]any)
}

func (m *MockConfig) Reload() error {
	args := m.Called()
	return args.Error(0)
}

// MockLogger implements contract.Logger for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Info(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Warn(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Error(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Fatal(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Debugf(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Infof(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Warnf(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Errorf(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Fatalf(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Debugw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Infow(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Warnw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Errorw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Fatalw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) With(keysAndValues ...any) contract.Logger {
	// Module tagging calls With during construction; return self so the same
	// mock receives subsequent calls.
	return m
}

func (m *MockLogger) WithContext(ctx context.Context) contract.Logger {
	args := m.Called(ctx)
	return args.Get(0).(contract.Logger)
}

func (m *MockLogger) WithError(err error) contract.Logger {
	args := m.Called(err)
	return args.Get(0).(contract.Logger)
}

func (m *MockLogger) GetLogger() *zap.SugaredLogger {
	args := m.Called()
	return args.Get(0).(*zap.SugaredLogger)
}

// MockApplication for testing
type MockApplication struct {
	mock.Mock
}

func (m *MockApplication) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockApplication) Version() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockApplication) Environment() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockApplication) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

func (m *MockApplication) Container() *fx.App {
	args := m.Called()
	return args.Get(0).(*fx.App)
}

func (m *MockApplication) IsRunning() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockApplication) Register(options ...fx.Option) error {
	args := m.Called(options)
	return args.Error(0)
}

func (m *MockApplication) AddModule(module contract.Module) error {
	args := m.Called(module)
	return args.Error(0)
}

func (m *MockApplication) AddProvider(provider contract.Provider) error {
	args := m.Called(provider)
	return args.Error(0)
}

func (m *MockApplication) AddInvoker(invoker contract.Invoker) error {
	args := m.Called(invoker)
	return args.Error(0)
}

func (m *MockApplication) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockApplication) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockApplication) Run() {
	m.Called()
}
