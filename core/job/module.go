package job

import (
	"context"
	"errors"

	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/internal/configvalidator"
)

// Module represents the job module for Fx dependency injection
type Module struct {
	manager *Manager
	logger  contract.Logger
	config  contract.Config

	// connectionSetInCode records that an Option supplied the Redis endpoint,
	// so ValidateConfig does not demand a JOB_REDIS_* environment variable.
	connectionSetInCode bool
}

// NewModule creates a new job module.
//
// Configuration resolves in layers: defaults, then JOB_* environment
// variables, then opts. Anything set through an Option wins over the
// environment. Option errors abort construction before any Redis connection
// is attempted, and every error is reported at once.
func NewModule(config contract.Config, logger contract.Logger, opts ...Option) (*Module, error) {
	// Tag once; the manager and workers share this module-scoped logger.
	logger = logger.With("module", "job")

	// Resolve job configuration: defaults -> environment -> code.
	jobConfig, connectionSetInCode, optErrs := resolveConfig(config, opts)
	if len(optErrs) > 0 {
		return nil, errors.Join(optErrs...)
	}

	// Create manager
	manager, err := NewManager(jobConfig, logger)
	if err != nil {
		return nil, err
	}

	return &Module{
		manager:             manager,
		logger:              logger,
		config:              config,
		connectionSetInCode: connectionSetInCode,
	}, nil
}

// Name returns the module name
func (m *Module) Name() string {
	return "job"
}

// OnStart is called when the module starts
func (m *Module) OnStart(ctx context.Context) error {
	// Validate configuration
	if err := m.ValidateConfig(); err != nil {
		return err
	}

	// Test connection
	if err := m.manager.Health(ctx); err != nil {
		return err
	}

	// Start the job manager
	if err := m.manager.Start(ctx); err != nil {
		return err
	}

	m.logger.Info("Job module started",
		"redis_addr", m.manager.config.RedactedTarget(),
		"concurrency", m.manager.config.Concurrency,
		"default_queue", m.manager.config.DefaultQueue,
		"scheduler_enabled", m.manager.config.SchedulerEnabled,
	)

	return nil
}

// OnStop is called when the module stops
func (m *Module) OnStop(ctx context.Context) error {
	m.logger.Debug("Job module stopping")

	// Stop the job manager
	if err := m.manager.Stop(ctx); err != nil {
		m.logger.Error("Error stopping job manager", "error", err)
		return err
	}

	m.logger.Info("Job module stopped")
	return nil
}

// Provide returns the job manager instance for Fx
func (m *Module) Provide() contract.JobManager {
	return m.manager
}

// ProvideJobManager returns the job manager instance for Fx (alias for Provide)
func (m *Module) ProvideJobManager() contract.JobManager {
	return m.manager
}

// ValidateConfig validates the job module configuration
func (m *Module) ValidateConfig() error {
	v := configvalidator.NewConfigValidator(m.config, "job")

	// Require at least one Redis configuration — from the environment or from
	// a WithRedisHosts option in code.
	hasJobConfig := m.connectionSetInCode ||
		m.config.Has("JOB_REDIS_URL") ||
		m.config.Has("JOB_REDIS_ADDR") ||
		m.config.Has("JOB_REDIS_HOST") ||
		m.config.Has("JOB_REDIS_HOSTS")

	if !hasJobConfig {
		v.RequireWithValidator("JOB_REDIS_ADDR", "Redis server address for job system", configvalidator.ValidateHostPort)
	}

	// Validate JOB_* specific settings if present
	if m.config.Has("JOB_REDIS_ADDR") {
		v.Optional("JOB_REDIS_ADDR", "Redis server address", configvalidator.ValidateHostPort)
	}

	if m.config.Has("JOB_REDIS_DB") {
		v.Optional("JOB_REDIS_DB", "Redis database number", configvalidator.ValidateNonNegativeInt)
	}

	if m.config.Has("JOB_REDIS_POOL_SIZE") {
		v.Optional("JOB_REDIS_POOL_SIZE", "Redis pool size", configvalidator.ValidatePositiveInt)
	}

	if m.config.Has("JOB_CONCURRENCY") {
		v.Optional("JOB_CONCURRENCY", "Number of concurrent workers", configvalidator.ValidatePositiveInt)
	}

	if m.config.Has("JOB_MAX_CONCURRENCY") {
		v.Optional("JOB_MAX_CONCURRENCY", "Maximum concurrent workers", configvalidator.ValidatePositiveInt)
	}

	if m.config.Has("JOB_DEFAULT_MAX_ATTEMPTS") {
		v.Optional("JOB_DEFAULT_MAX_ATTEMPTS", "Default maximum retry attempts", configvalidator.ValidatePositiveInt)
	}

	return v.Validate()
}
