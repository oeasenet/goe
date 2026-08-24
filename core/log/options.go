package log

import "fmt"

// Option configures the log module from Go code instead of environment
// variables. Options are applied after the environment, so anything set here
// wins over the matching APP_* value while the environment still supplies
// everything left unset.
//
// An option that fails records its error and applies nothing; ValidateConfig
// reports the failure so startup aborts before the application serves.
type Option func(*settings) error

// settings holds the code-first configuration collected from Options.
type settings struct {
	// version overrides the APP_VERSION-derived version base field.
	version string
	// baseFields holds extra key/value pairs stamped alongside the identity
	// fields on every JSON log line. Keys are validated to be strings.
	baseFields []any
}

// WithVersion sets the version base field stamped on every JSON log line,
// winning over APP_VERSION. Use it to carry an ldflags-stamped build version
// without environment plumbing:
//
//	goe.New(goe.Options{
//	    Log: []log.Option{log.WithVersion(buildinfo.Version)},
//	})
//
// An empty value is ignored, so passing a never-stamped ldflags variable keeps
// the APP_VERSION fallback.
func WithVersion(version string) Option {
	return func(s *settings) error {
		s.version = version
		return nil
	}
}

// WithBaseFields adds key/value pairs to the base fields stamped on every JSON
// log line, alongside service/env/version. Keys must be strings and arguments
// must come in pairs; anything else fails startup via ValidateConfig.
func WithBaseFields(keysAndValues ...any) Option {
	return func(s *settings) error {
		if len(keysAndValues)%2 != 0 {
			return fmt.Errorf("log.WithBaseFields: odd number of arguments (%d), want key/value pairs", len(keysAndValues))
		}
		for i := 0; i < len(keysAndValues); i += 2 {
			if _, ok := keysAndValues[i].(string); !ok {
				return fmt.Errorf("log.WithBaseFields: key at index %d is %T, want string", i, keysAndValues[i])
			}
		}
		s.baseFields = append(s.baseFields, keysAndValues...)
		return nil
	}
}
