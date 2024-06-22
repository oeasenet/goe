package mongodb

import (
	"context"
	"go.oease.dev/omgo"
	"time"
)

type Mongodb struct {
	initialized bool
	client      *omgo.Client
	dbName      string
	logger      Logger
}

// NewMongoDB returns a new instance of Mongodb connected to the specified database.
// It takes a connection URI, database name, and an optional logger as parameters.
// If no logger is provided, it uses a default logger.
// It returns a pointer to Mongodb and an error if there is any.
func NewMongoDB(connectionUri string, databaseName string, logger ...Logger) (*Mongodb, error) {
	ctx := context.Background()
	client, err := omgo.NewClient(ctx, &omgo.Config{
		Uri:      connectionUri,
		Database: databaseName,
	})
	if err != nil {
		return nil, err
	}
	m := &Mongodb{
		client: client,
		dbName: databaseName,
	}
	if len(logger) > 0 && logger[0] != nil {
		m.logger = logger[0]
	} else {
		// Default logger
		m.logger = newDefaultLogger()
	}
	m.initialized = true
	return m, nil
}

// newCtxWithTimeout returns a new context with the specified timeout.
// It takes a timeout duration as a parameter.
// If an error occurs while creating the context with the timeout, it logs the error and returns a new context without a timeout.
// It returns the created context.
func (m *Mongodb) newCtxWithTimeout(timeout time.Duration) context.Context {
	ctx, err := context.WithTimeout(context.Background(), timeout)
	if err != nil {
		m.logger.Error(err)
		// Return a new context without timeout
		return context.Background()
	}
	return ctx
}

// newCtx returns a new context with a default timeout of 10 seconds. It internally calls newCtxWithTimeout with the specified duration.
// It doesn't take any parameters.
// It returns the created context.
func (m *Mongodb) newCtx() context.Context {
	return m.newCtxWithTimeout(10 * time.Second)
}

// Close closes the connection to the database.
func (m *Mongodb) Close() error {
	if err := m.client.Close(m.newCtx()); err != nil {
		return err
	}
	return nil
}

// Ping checks the connection to the database.
func (m *Mongodb) Ping() error {
	if err := m.client.Ping(10); err != nil {
		return err
	}
	return nil
}
