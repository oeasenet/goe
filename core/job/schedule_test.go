package job

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEvery(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		expected string
	}{
		{"minute", time.Minute, "every 1m0s"},
		{"5 minutes", 5 * time.Minute, "every 5m0s"},
		{"hour", time.Hour, "every 1h0m0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Every(tt.interval)
			assert.Equal(t, tt.expected, s.String())
			assert.Empty(t, s.Cron())

			// Test Next
			now := time.Now()
			next := s.Next(now)
			assert.True(t, next.After(now))
			assert.Equal(t, now.Add(tt.interval), next)
		})
	}
}

func TestEveryHelpers(t *testing.T) {
	tests := []struct {
		name     string
		schedule func() testSchedule
		interval time.Duration
	}{
		{"EveryMinute", func() testSchedule { return EveryMinute().(testSchedule) }, time.Minute},
		{"EveryFiveMinutes", func() testSchedule { return EveryFiveMinutes().(testSchedule) }, 5 * time.Minute},
		{"EveryTenMinutes", func() testSchedule { return EveryTenMinutes().(testSchedule) }, 10 * time.Minute},
		{"EveryFifteenMinutes", func() testSchedule { return EveryFifteenMinutes().(testSchedule) }, 15 * time.Minute},
		{"EveryThirtyMinutes", func() testSchedule { return EveryThirtyMinutes().(testSchedule) }, 30 * time.Minute},
		{"Hourly", func() testSchedule { return Hourly().(testSchedule) }, time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.schedule()
			now := time.Now()
			next := s.Next(now)
			assert.Equal(t, now.Add(tt.interval), next)
		})
	}
}

func TestDailyAt(t *testing.T) {
	// Test at 9:30 AM
	s := DailyAt(9, 30)
	assert.Equal(t, "daily at 09:30", s.String())
	assert.Equal(t, "30 9 * * *", s.Cron())

	// Test Next - if it's before 9:30, should return today at 9:30
	morning := time.Date(2024, 1, 15, 7, 0, 0, 0, time.UTC)
	next := s.Next(morning)
	expected := time.Date(2024, 1, 15, 9, 30, 0, 0, time.UTC)
	assert.Equal(t, expected, next)

	// Test Next - if it's after 9:30, should return tomorrow at 9:30
	afternoon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	next = s.Next(afternoon)
	expected = time.Date(2024, 1, 16, 9, 30, 0, 0, time.UTC)
	assert.Equal(t, expected, next)
}

func TestDaily(t *testing.T) {
	s := Daily()
	assert.Equal(t, "daily at 00:00", s.String())
	assert.Equal(t, "0 0 * * *", s.Cron())
}

func TestWeeklyOn(t *testing.T) {
	s := WeeklyOn(time.Monday, 9, 0)
	assert.Contains(t, s.String(), "Monday")
	assert.Contains(t, s.String(), "09:00")

	// Find the next Monday at 9:00
	// Start from a Wednesday
	wednesday := time.Date(2024, 1, 17, 10, 0, 0, 0, time.UTC) // Jan 17, 2024 is Wednesday
	next := s.Next(wednesday)
	assert.Equal(t, time.Monday, next.Weekday())
	assert.Equal(t, 9, next.Hour())
	assert.Equal(t, 0, next.Minute())
}

func TestWeekdays(t *testing.T) {
	s := Weekdays(9, 0)
	assert.Contains(t, s.String(), "Monday")
	assert.Contains(t, s.String(), "Friday")

	// Start from Saturday, should get Monday
	saturday := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC) // Jan 20, 2024 is Saturday
	next := s.Next(saturday)
	assert.Equal(t, time.Monday, next.Weekday())
}

func TestWeekends(t *testing.T) {
	s := Weekends(10, 0)
	assert.Contains(t, s.String(), "Saturday")
	assert.Contains(t, s.String(), "Sunday")

	// Start from Monday, should get Saturday
	monday := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC) // Jan 15, 2024 is Monday
	next := s.Next(monday)
	assert.Equal(t, time.Saturday, next.Weekday())
}

func TestMonthlyOn(t *testing.T) {
	// On the 15th at 9:00
	s := MonthlyOn(15, 9, 0)
	assert.Contains(t, s.String(), "15")
	assert.Contains(t, s.String(), "09:00")

	// Before the 15th, should get this month
	jan10 := time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)
	next := s.Next(jan10)
	assert.Equal(t, 15, next.Day())
	assert.Equal(t, time.January, next.Month())

	// After the 15th, should get next month
	jan20 := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC)
	next = s.Next(jan20)
	assert.Equal(t, 15, next.Day())
	assert.Equal(t, time.February, next.Month())
}

func TestLastDayOfMonth(t *testing.T) {
	s := LastDayOfMonth(17, 0)
	assert.Contains(t, s.String(), "last")

	// January has 31 days
	jan10 := time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)
	next := s.Next(jan10)
	assert.Equal(t, 31, next.Day())
	assert.Equal(t, time.January, next.Month())

	// February 2024 has 29 days (leap year)
	feb1 := time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC)
	next = s.Next(feb1)
	assert.Equal(t, 29, next.Day())
	assert.Equal(t, time.February, next.Month())
}

func TestCron(t *testing.T) {
	tests := []struct {
		name string
		expr string
		desc string
	}{
		{"every minute", "* * * * *", "cron: * * * * *"},
		{"every hour", "0 * * * *", "cron: 0 * * * *"},
		{"every day at midnight", "0 0 * * *", "cron: 0 0 * * *"},
		{"weekdays at 9am", "0 9 * * 1-5", "cron: 0 9 * * 1-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Cron(tt.expr)
			assert.Equal(t, tt.desc, s.String())
			assert.Equal(t, tt.expr, s.Cron())

			// Should return a valid next time
			now := time.Now()
			next := s.Next(now)
			assert.True(t, next.After(now))
		})
	}
}

func TestCronInvalid(t *testing.T) {
	s := Cron("invalid cron")
	assert.Contains(t, s.String(), "invalid")

	// Should return a time far in the future (effectively never)
	now := time.Now()
	next := s.Next(now)
	assert.True(t, next.Year() > 2100)
}

func TestHourlyAt(t *testing.T) {
	s := HourlyAt(15)
	// Should run at :15 of every hour
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	next := s.Next(now)
	assert.Equal(t, 15, next.Minute())
}

func TestTwiceDaily(t *testing.T) {
	s := TwiceDaily(8, 20)
	assert.Equal(t, "0 8,20 * * *", s.Cron())
}

func TestQuarterly(t *testing.T) {
	s := Quarterly(9, 0)
	// Should include months 1, 4, 7, 10
	assert.Contains(t, s.Cron(), "1,4,7,10")
}

func TestYearly(t *testing.T) {
	s := Yearly()
	assert.Equal(t, "0 0 1 1 *", s.Cron())
}

func TestYearlyOn(t *testing.T) {
	s := YearlyOn(time.March, 15, 9, 0)
	// March 15th at 9:00 AM
	assert.Contains(t, s.Cron(), "15 3")
}

func TestBetween(t *testing.T) {
	// Every 15 minutes, but only between 9 AM and 5 PM
	base := Every(15 * time.Minute)
	s := Between(base, 9, 17)

	assert.Contains(t, s.String(), "between 09:00 and 17:00")

	// At 8:00 AM, should skip to 9:00
	morning := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	next := s.Next(morning)
	assert.True(t, next.Hour() >= 9 && next.Hour() < 17)

	// At 6:00 PM, should skip to next day at 9:00
	evening := time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC)
	next = s.Next(evening)
	assert.True(t, next.Hour() >= 9 && next.Hour() < 17)
	assert.Equal(t, 16, next.Day()) // Next day
}

func TestSkipWeekends(t *testing.T) {
	base := Daily()
	s := SkipWeekends(base)

	assert.Contains(t, s.String(), "skip")
	assert.Contains(t, s.String(), "Saturday")

	// Friday, should get Monday
	friday := time.Date(2024, 1, 19, 10, 0, 0, 0, time.UTC) // Jan 19, 2024 is Friday
	next := s.Next(friday)
	assert.Equal(t, time.Monday, next.Weekday())

	// Wednesday, should get Thursday
	wednesday := time.Date(2024, 1, 17, 10, 0, 0, 0, time.UTC)
	next = s.Next(wednesday)
	assert.Equal(t, time.Thursday, next.Weekday())
}

// testSchedule interface for testing
type testSchedule interface {
	Next(time.Time) time.Time
	String() string
	Cron() string
}
