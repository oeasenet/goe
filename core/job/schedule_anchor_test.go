//go:build integration

package job

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/internal/testutil"
)

// TestSchedule_DoesNotFireOnBoot guards a regression that shipped to production:
// last-run times were held in a process-local map, so an unseen schedule had a
// zero last-run. Schedule.Next(zeroTime) returns a moment in year 1, which is
// always in the past, so every schedule fired immediately on the first tick
// regardless of its period — a daily 03:00 job ran on every deploy.
func TestSchedule_DoesNotFireOnBoot(t *testing.T) {
	manager := newTestManager(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var hourly, daily, monthly atomic.Int32

	register := func(name string, s contract.Schedule, counter *atomic.Int32) {
		require.NoError(t, manager.RegisterSchedule(&contract.ScheduledJob{
			Name:     name,
			Schedule: s,
			Overlap:  true, // isolate the anchor behaviour from uniqueness suppression
			Handler: contract.JobHandlerFunc(func(context.Context, contract.Job) error {
				counter.Add(1)
				return nil
			}),
		}))
	}
	register("anchor-hourly", Every(time.Hour), &hourly)
	register("anchor-daily", DailyAt(3, 0), &daily)
	register("anchor-monthly", MonthlyOn(15, 12, 0), &monthly)

	require.NoError(t, manager.Start(ctx))
	defer manager.Stop(ctx)

	// Several scheduler ticks (interval is 100ms in tests) must pass without any
	// of these long-period schedules firing.
	time.Sleep(3 * time.Second)

	assert.Zero(t, hourly.Load(), "Every(1h) must not fire at boot")
	assert.Zero(t, daily.Load(), "DailyAt(03:00) must not fire at boot")
	assert.Zero(t, monthly.Load(), "MonthlyOn(15th) must not fire at boot")
}

// TestSchedule_ShortIntervalStillFires is the counterpart: anchoring must not
// stop schedules from running once their period genuinely elapses.
func TestSchedule_ShortIntervalStillFires(t *testing.T) {
	manager := newTestManager(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var fired atomic.Int32

	require.NoError(t, manager.RegisterSchedule(&contract.ScheduledJob{
		Name:     "anchor-fast",
		Schedule: Every(time.Second),
		Overlap:  true,
		Handler: contract.JobHandlerFunc(func(context.Context, contract.Job) error {
			fired.Add(1)
			return nil
		}),
	}))

	require.NoError(t, manager.Start(ctx))
	defer manager.Stop(ctx)

	// More than one execution proves the schedule keeps recurring after the
	// initial anchor, not just that it fired once.
	testutil.AssertEventually(t, func() bool { return fired.Load() > 1 },
		25*time.Second, "1s schedule should fire repeatedly after anchoring")
	assert.Greater(t, fired.Load(), int32(1))
}

// TestSchedule_LastRunSurvivesRestart verifies the marker is shared rather than
// per-process: a second Manager against the same Redis must inherit the anchor
// instead of treating the schedule as unseen and firing immediately.
func TestSchedule_LastRunSurvivesRestart(t *testing.T) {
	first := newTestManager(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var firstFired, secondFired atomic.Int32

	mk := func(m *Manager, counter *atomic.Int32) {
		require.NoError(t, m.RegisterSchedule(&contract.ScheduledJob{
			Name:     "anchor-shared",
			Schedule: Every(time.Hour),
			Overlap:  true,
			Handler: contract.JobHandlerFunc(func(context.Context, contract.Job) error {
				counter.Add(1)
				return nil
			}),
		}))
	}

	mk(first, &firstFired)
	require.NoError(t, first.Start(ctx))
	time.Sleep(1 * time.Second) // let the first instance anchor the schedule
	require.NoError(t, first.Stop(ctx))

	// Second instance, same Redis. newTestManager flushes the DB, so build this
	// one directly to preserve the marker the first instance wrote.
	second, err := NewManager(getTestConfig(), &testLogger{t: t})
	require.NoError(t, err)
	mk(second, &secondFired)
	require.NoError(t, second.Start(ctx))
	defer second.Stop(ctx)

	time.Sleep(2 * time.Second)

	assert.Zero(t, firstFired.Load(), "hourly schedule must not fire at boot")
	assert.Zero(t, secondFired.Load(),
		"a restarted instance must inherit the shared anchor, not re-fire")
}
