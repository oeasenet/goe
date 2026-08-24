// Package log contains the permanent logging module. It bridges contract.Logger
// with zap and ensures consistent structured logging across optional modules.
//
// # Configuration
//
//	LOG_LEVEL=info                  # debug | info | warn | error (default info)
//	LOG_FORMAT=text                 # text (default) | json; json emits machine-readable lines
//	LOG_OUTPUT=console              # console (default) and/or file
//	LOG_CALLER=true                 # annotate lines with file:line
//	LOG_STACKTRACE=false            # attach stacktraces at error level
//	LOG_MODULE_LEVELS=job:debug     # per-module level overrides
//
// # JSON identity fields
//
// With LOG_FORMAT=json every line carries base identity fields resolved once at
// startup: service from APP_NAME, env from OEASE_ENV (falling back to GOE_ENV)
// and version from APP_VERSION. Unset values are omitted, and text output stays
// clean. Code-first configuration wins over the environment:
//
//	goe.New(goe.Options{
//	    Log: []log.Option{
//	        log.WithVersion(buildinfo.Version),        // ldflags-stamped build version
//	        log.WithBaseFields("region", "us-east-1"), // extra identity fields
//	    },
//	})
//
// # Request correlation
//
// Logger.WithContext(ctx) enriches a logger from a context.Context: request_id
// from the fiber requestid middleware (stored into the request context when
// PassLocalsToContext is on, GOE's default) and trace_id/span_id from an active
// OpenTelemetry span. Handlers can use http.WithReqCtx(c) directly; WithContext
// covers service-layer code that only holds a context.
package log
