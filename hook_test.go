package goe_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"go.oease.dev/goe/v2"
)

func TestLifecycleHooks_Options(t *testing.T) {
	var startExecuted atomic.Bool
	var stopExecuted atomic.Bool

	// Create application with hooks in options
	app := goe.New(goe.Options{
		OnStart: []func(context.Context) error{
			func(ctx context.Context) error {
				startExecuted.Store(true)
				return nil
			},
		},
		OnStop: []func(context.Context) error{
			func(ctx context.Context) error {
				stopExecuted.Store(true)
				return nil
			},
		},
	})

	// Start application
	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		t.Fatalf("failed to start app: %v", err)
	}

	// Verify OnStart was executed
	if !startExecuted.Load() {
		t.Error("OnStart hook was not executed")
	}

	// Stop application
	if err := app.Stop(ctx); err != nil {
		t.Fatalf("failed to stop app: %v", err)
	}

	// Verify OnStop was executed
	if !stopExecuted.Load() {
		t.Error("OnStop hook was not executed")
	}
}

func TestLifecycleHooks_GlobalFunctionsPanic(t *testing.T) {
	// Create application first
	_ = goe.New(goe.Options{})

	t.Run("OnStart panics after app creation", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when calling OnStart after app creation")
			} else {
				expectedMsg := "OnStart cannot be called after goe.New(). Use Options.OnStart instead"
				if r != expectedMsg {
					t.Errorf("expected panic message %q, got %q", expectedMsg, r)
				}
			}
		}()

		goe.OnStart(func(ctx context.Context) error {
			return nil
		})
	})

	t.Run("OnStop panics after app creation", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when calling OnStop after app creation")
			} else {
				expectedMsg := "OnStop cannot be called after goe.New(). Use Options.OnStop instead"
				if r != expectedMsg {
					t.Errorf("expected panic message %q, got %q", expectedMsg, r)
				}
			}
		}()

		goe.OnStop(func(ctx context.Context) error {
			return nil
		})
	})
}

func TestLifecycleHooks_WithHooks(t *testing.T) {
	var startCount atomic.Int32
	var stopCount atomic.Int32

	// Create application with WithHooks helper
	app := goe.New(
		goe.WithHooks(
			goe.WithOnStart(func(ctx context.Context) error {
				startCount.Add(1)
				return nil
			}),
			goe.WithOnStart(func(ctx context.Context) error {
				startCount.Add(1)
				return nil
			}),
			goe.WithOnStop(func(ctx context.Context) error {
				stopCount.Add(1)
				return nil
			}),
		),
		goe.Options{
			WithHTTP: false,
		},
	)

	// Start application
	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		t.Fatalf("failed to start app: %v", err)
	}

	// Verify both OnStart hooks were executed
	if startCount.Load() != 2 {
		t.Errorf("expected 2 OnStart hooks to execute, got %d", startCount.Load())
	}

	// Stop application
	if err := app.Stop(ctx); err != nil {
		t.Fatalf("failed to stop app: %v", err)
	}

	// Verify OnStop hook was executed
	if stopCount.Load() != 1 {
		t.Errorf("expected 1 OnStop hook to execute, got %d", stopCount.Load())
	}
}

func TestLifecycleHooks_ExecutionOrder(t *testing.T) {
	var executionOrder []string
	orderChan := make(chan string, 10)

	// Create application with multiple hooks
	app := goe.New(goe.Options{
		OnStart: []func(context.Context) error{
			func(ctx context.Context) error {
				orderChan <- "start-1"
				return nil
			},
			func(ctx context.Context) error {
				orderChan <- "start-2"
				return nil
			},
			func(ctx context.Context) error {
				orderChan <- "start-3"
				return nil
			},
		},
		OnStop: []func(context.Context) error{
			func(ctx context.Context) error {
				orderChan <- "stop-1"
				return nil
			},
			func(ctx context.Context) error {
				orderChan <- "stop-2"
				return nil
			},
			func(ctx context.Context) error {
				orderChan <- "stop-3"
				return nil
			},
		},
	})

	// Start application
	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		t.Fatalf("failed to start app: %v", err)
	}

	// Stop application
	if err := app.Stop(ctx); err != nil {
		t.Fatalf("failed to stop app: %v", err)
	}

	// Collect execution order
	close(orderChan)
	for item := range orderChan {
		executionOrder = append(executionOrder, item)
	}

	// Verify OnStart hooks executed in order
	if len(executionOrder) < 3 || executionOrder[0] != "start-1" || executionOrder[1] != "start-2" || executionOrder[2] != "start-3" {
		t.Errorf("OnStart hooks did not execute in order: %v", executionOrder)
	}

	// Verify OnStop hooks executed in reverse order
	if len(executionOrder) < 6 || executionOrder[3] != "stop-3" || executionOrder[4] != "stop-2" || executionOrder[5] != "stop-1" {
		t.Errorf("OnStop hooks did not execute in reverse order: %v", executionOrder)
	}
}

func TestLifecycleHooks_ErrorHandling(t *testing.T) {
	t.Run("OnStart error aborts startup", func(t *testing.T) {
		var firstHookExecuted atomic.Bool
		var secondHookExecuted atomic.Bool

		app := goe.New(goe.Options{
			OnStart: []func(context.Context) error{
				func(ctx context.Context) error {
					firstHookExecuted.Store(true)
					return errors.New("startup failed")
				},
				func(ctx context.Context) error {
					secondHookExecuted.Store(true)
					return nil
				},
			},
		})

		// Start should fail
		ctx := context.Background()
		err := app.Start(ctx)
		if err == nil {
			t.Fatal("expected error from failing OnStart hook")
		}

		// First hook should have executed
		if !firstHookExecuted.Load() {
			t.Error("first hook should have executed")
		}

		// Second hook should not have executed
		if secondHookExecuted.Load() {
			t.Error("second hook should not have executed after first hook failed")
		}
	})

	t.Run("OnStop errors don't prevent shutdown", func(t *testing.T) {
		var firstHookExecuted atomic.Bool
		var secondHookExecuted atomic.Bool

		app := goe.New(goe.Options{
			OnStop: []func(context.Context) error{
				func(ctx context.Context) error {
					firstHookExecuted.Store(true)
					return errors.New("cleanup failed")
				},
				func(ctx context.Context) error {
					secondHookExecuted.Store(true)
					return nil
				},
			},
		})

		// Start application
		ctx := context.Background()
		if err := app.Start(ctx); err != nil {
			t.Fatalf("failed to start app: %v", err)
		}

		// Stop application - should not fail despite error
		if err := app.Stop(ctx); err != nil {
			t.Fatalf("stop should not fail due to hook error: %v", err)
		}

		// Both hooks should have executed
		if !firstHookExecuted.Load() {
			t.Error("first stop hook should have executed")
		}
		if !secondHookExecuted.Load() {
			t.Error("second stop hook should have executed despite first hook error")
		}
	})
}

func TestLifecycleHooks_ServiceAccess(t *testing.T) {
	var configAccessible atomic.Bool
	var loggerAccessible atomic.Bool

	// Create application with services
	app := goe.New(goe.Options{
		WithHTTP: true,
		OnStart: []func(context.Context) error{
			func(ctx context.Context) error {
				// Access config
				if config := goe.Config(); config != nil {
					configAccessible.Store(true)
				}

				// Access logger
				if logger := goe.Log(); logger != nil {
					loggerAccessible.Store(true)
					logger.Info("Hook accessing services")
				}

				return nil
			},
		},
	})

	// Start application
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Start(ctx); err != nil {
		t.Fatalf("failed to start app: %v", err)
	}

	// Verify services were accessible
	if !configAccessible.Load() {
		t.Error("config was not accessible in hook")
	}
	if !loggerAccessible.Load() {
		t.Error("logger was not accessible in hook")
	}

	// Stop application
	if err := app.Stop(context.Background()); err != nil {
		t.Fatalf("failed to stop app: %v", err)
	}
}

func TestLifecycleHooks_ContextRespect(t *testing.T) {
	var hookStarted atomic.Bool
	var hookCompleted atomic.Bool

	app := goe.New(goe.Options{
		OnStart: []func(context.Context) error{
			func(ctx context.Context) error {
				hookStarted.Store(true)

				// Simulate long-running operation
				select {
				case <-time.After(100 * time.Millisecond):
					hookCompleted.Store(true)
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			},
		},
	})

	// Start with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := app.Start(ctx)

	// Should timeout
	if err == nil {
		t.Fatal("expected context timeout error")
	}

	// Hook should have started but not completed
	if !hookStarted.Load() {
		t.Error("hook should have started")
	}
	if hookCompleted.Load() {
		t.Error("hook should not have completed due to timeout")
	}
}
