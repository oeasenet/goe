package job

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
)

func newRegistryOnlyManager() *Manager {
	// No Redis client: these cases must be rejected before any I/O.
	return &Manager{
		config:    DefaultConfig(),
		logger:    &nopLogger{},
		schedules: make(map[string]*contract.ScheduledJob),
	}
}

func noopHandler() contract.JobHandler {
	return contract.JobHandlerFunc(func(context.Context, contract.Job) error { return nil })
}

// TestRegisterSchedule_RejectsInvalidCron guards a silent-failure bug: Cron()
// cannot return an error, so a parse failure produced a schedule whose next run
// is in the year 9999. Registration succeeded and the job simply never fired.
func TestRegisterSchedule_RejectsInvalidCron(t *testing.T) {
	m := newRegistryOnlyManager()

	err := m.RegisterSchedule(&contract.ScheduledJob{
		Name:     "typo",
		Schedule: Cron("this is not a cron expression"),
		Handler:  noopHandler(),
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidSchedule)
	assert.Contains(t, err.Error(), "typo", "error should name the schedule")
	assert.Empty(t, m.schedules, "an invalid schedule must not be registered")
}

// TestRegisterSchedule_RejectsInvalidCronThroughCombinators ensures wrapping a
// broken expression does not hide it.
func TestRegisterSchedule_RejectsInvalidCronThroughCombinators(t *testing.T) {
	for name, schedule := range map[string]contract.Schedule{
		"Between":      Between(Cron("nope"), 9, 17),
		"SkipWeekends": SkipWeekends(Cron("nope")),
		"nested":       Between(SkipWeekends(Cron("nope")), 9, 17),
	} {
		t.Run(name, func(t *testing.T) {
			m := newRegistryOnlyManager()
			err := m.RegisterSchedule(&contract.ScheduledJob{
				Name: "wrapped", Schedule: schedule, Handler: noopHandler(),
			})
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidSchedule)
		})
	}
}

// TestRegisterSchedule_RejectsNilFields guards against a panic in the scheduler
// goroutine, which has no recover and would take the process down.
func TestRegisterSchedule_RejectsNilFields(t *testing.T) {
	cases := map[string]*contract.ScheduledJob{
		"nil schedule": {Name: "a", Handler: noopHandler()},
		"nil handler":  {Name: "a", Schedule: Every(time.Minute)},
		"empty name":   {Schedule: Every(time.Minute), Handler: noopHandler()},
	}
	for name, job := range cases {
		t.Run(name, func(t *testing.T) {
			m := newRegistryOnlyManager()
			err := m.RegisterSchedule(job)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidSchedule)
			assert.Empty(t, m.schedules)
		})
	}

	t.Run("nil job", func(t *testing.T) {
		m := newRegistryOnlyManager()
		assert.ErrorIs(t, m.RegisterSchedule(nil), ErrInvalidSchedule)
	})
}

// TestRegisterSchedule_AcceptsValidSchedules confirms the validation does not
// reject legitimate schedules, including every builder helper.
func TestRegisterSchedule_AcceptsValidSchedules(t *testing.T) {
	schedules := map[string]contract.Schedule{
		"Every":        Every(time.Minute),
		"Cron":         Cron("*/5 * * * *"),
		"Daily":        Daily(),
		"DailyAt":      DailyAt(3, 30),
		"HourlyAt":     HourlyAt(15),
		"TwiceDaily":   TwiceDaily(8, 20),
		"Weekly":       Weekly(),
		"WeeklyOn":     WeeklyOn(time.Monday, 9, 0),
		"Weekdays":     Weekdays(9, 0),
		"Weekends":     Weekends(10, 0),
		"Monthly":      Monthly(),
		"MonthlyOn":    MonthlyOn(15, 12, 0),
		"LastDay":      LastDayOfMonth(17, 0),
		"Quarterly":    Quarterly(9, 0),
		"Yearly":       Yearly(),
		"YearlyOn":     YearlyOn(time.March, 15, 9, 0),
		"Between":      Between(Every(15*time.Minute), 9, 17),
		"SkipWeekends": SkipWeekends(Daily()),
	}
	for name, schedule := range schedules {
		t.Run(name, func(t *testing.T) {
			m := newRegistryOnlyManager()
			err := m.RegisterSchedule(&contract.ScheduledJob{
				Name: name, Schedule: schedule, Handler: noopHandler(),
			})
			require.NoError(t, err, "%s should be a valid schedule", name)
			assert.Len(t, m.schedules, 1)

			// A valid schedule must also produce a real next occurrence rather
			// than the year-9999 sentinel used for broken ones.
			next := schedule.Next(time.Now())
			assert.Less(t, next.Year(), 9999, "%s produced a never-fires schedule", name)
		})
	}
}

func TestScheduleErr(t *testing.T) {
	assert.NoError(t, scheduleErr(Every(time.Minute)))
	assert.NoError(t, scheduleErr(Cron("0 0 * * *")))

	// The underlying cron parse error is preserved, so callers see why.
	err := scheduleErr(Cron("bad"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected exactly 5 fields")
}
