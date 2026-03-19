package mongodb

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.oease.dev/goe/v2/contract"
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

func (m *TestMockLogger) Info(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Debug(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Warn(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Error(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Fatal(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Panic(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Panicf(template string, args ...any) {
	m.Called(template, args)
}

func (m *TestMockLogger) Debugf(template string, args ...any) {
	m.Called(template, args)
}

func (m *TestMockLogger) Infof(template string, args ...any) {
	m.Called(template, args)
}

func (m *TestMockLogger) Warnf(template string, args ...any) {
	m.Called(template, args)
}

func (m *TestMockLogger) Errorf(template string, args ...any) {
	m.Called(template, args)
}

func (m *TestMockLogger) Fatalf(template string, args ...any) {
	m.Called(template, args)
}

func (m *TestMockLogger) Debugw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Infow(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Warnw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Errorw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) Fatalw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *TestMockLogger) With(keysAndValues ...any) contract.Logger {
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
