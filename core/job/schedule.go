package job

import (
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"go.oease.dev/goe/v2/contract"
)

// cronSchedule wraps a cron expression
type cronSchedule struct {
	expr     string
	schedule cron.Schedule
}

func (s *cronSchedule) Next(after time.Time) time.Time {
	return s.schedule.Next(after)
}

func (s *cronSchedule) String() string {
	return fmt.Sprintf("cron: %s", s.expr)
}

func (s *cronSchedule) Cron() string {
	return s.expr
}

// intervalSchedule runs at fixed intervals
type intervalSchedule struct {
	interval time.Duration
}

func (s *intervalSchedule) Next(after time.Time) time.Time {
	return after.Add(s.interval)
}

func (s *intervalSchedule) String() string {
	return fmt.Sprintf("every %s", s.interval)
}

func (s *intervalSchedule) Cron() string {
	return ""
}

// dailySchedule runs once per day at a specific time
type dailySchedule struct {
	hour   int
	minute int
}

func (s *dailySchedule) Next(after time.Time) time.Time {
	next := time.Date(after.Year(), after.Month(), after.Day(), s.hour, s.minute, 0, 0, after.Location())
	if !next.After(after) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

func (s *dailySchedule) String() string {
	return fmt.Sprintf("daily at %02d:%02d", s.hour, s.minute)
}

func (s *dailySchedule) Cron() string {
	return fmt.Sprintf("%d %d * * *", s.minute, s.hour)
}

// weeklySchedule runs once per week on specific days
type weeklySchedule struct {
	weekdays []time.Weekday
	hour     int
	minute   int
}

func (s *weeklySchedule) Next(after time.Time) time.Time {
	// Start from the beginning of the next minute
	next := time.Date(after.Year(), after.Month(), after.Day(), s.hour, s.minute, 0, 0, after.Location())

	// If today's scheduled time has passed, start checking from tomorrow
	if !next.After(after) {
		next = next.AddDate(0, 0, 1)
	}

	// Find the next matching weekday
	for i := 0; i < 8; i++ {
		for _, wd := range s.weekdays {
			if next.Weekday() == wd {
				return next
			}
		}
		next = next.AddDate(0, 0, 1)
	}

	return next
}

func (s *weeklySchedule) String() string {
	days := make([]string, len(s.weekdays))
	for i, wd := range s.weekdays {
		days[i] = wd.String()
	}
	return fmt.Sprintf("weekly on %s at %02d:%02d", strings.Join(days, ", "), s.hour, s.minute)
}

func (s *weeklySchedule) Cron() string {
	days := make([]string, len(s.weekdays))
	for i, wd := range s.weekdays {
		days[i] = fmt.Sprintf("%d", wd)
	}
	return fmt.Sprintf("%d %d * * %s", s.minute, s.hour, strings.Join(days, ","))
}

// monthlySchedule runs once per month on specific days
type monthlySchedule struct {
	days   []int // 1-31, or -1 for last day
	hour   int
	minute int
}

func (s *monthlySchedule) Next(after time.Time) time.Time {
	next := time.Date(after.Year(), after.Month(), 1, s.hour, s.minute, 0, 0, after.Location())

	for i := 0; i < 13; i++ { // Check up to 13 months ahead
		lastDay := time.Date(next.Year(), next.Month()+1, 0, 0, 0, 0, 0, next.Location()).Day()

		for _, day := range s.days {
			targetDay := day
			if day == -1 || day > lastDay {
				targetDay = lastDay
			}

			candidate := time.Date(next.Year(), next.Month(), targetDay, s.hour, s.minute, 0, 0, next.Location())
			if candidate.After(after) {
				return candidate
			}
		}

		// Move to next month
		next = time.Date(next.Year(), next.Month()+1, 1, s.hour, s.minute, 0, 0, next.Location())
	}

	return next
}

func (s *monthlySchedule) String() string {
	days := make([]string, len(s.days))
	for i, d := range s.days {
		if d == -1 {
			days[i] = "last"
		} else {
			days[i] = fmt.Sprintf("%d", d)
		}
	}
	return fmt.Sprintf("monthly on day %s at %02d:%02d", strings.Join(days, ", "), s.hour, s.minute)
}

func (s *monthlySchedule) Cron() string {
	days := make([]string, len(s.days))
	for i, d := range s.days {
		if d == -1 {
			days[i] = "L"
		} else {
			days[i] = fmt.Sprintf("%d", d)
		}
	}
	return fmt.Sprintf("%d %d %s * *", s.minute, s.hour, strings.Join(days, ","))
}

// =============================================================================
// Schedule Builder - Fluent API for creating schedules
// =============================================================================

// ScheduleBuilder provides a fluent API for creating schedules
type ScheduleBuilder struct {
	err      error
	schedule contract.Schedule
}

// Every creates a schedule that runs at fixed intervals
//
// Examples:
//
//	Every(5 * time.Minute)   // Every 5 minutes
//	Every(time.Hour)         // Every hour
//	Every(24 * time.Hour)    // Every day
func Every(interval time.Duration) contract.Schedule {
	if interval <= 0 {
		interval = time.Minute
	}
	return &intervalSchedule{interval: interval}
}

// EveryMinute creates a schedule that runs every minute
func EveryMinute() contract.Schedule {
	return Every(time.Minute)
}

// EveryFiveMinutes creates a schedule that runs every 5 minutes
func EveryFiveMinutes() contract.Schedule {
	return Every(5 * time.Minute)
}

// EveryTenMinutes creates a schedule that runs every 10 minutes
func EveryTenMinutes() contract.Schedule {
	return Every(10 * time.Minute)
}

// EveryFifteenMinutes creates a schedule that runs every 15 minutes
func EveryFifteenMinutes() contract.Schedule {
	return Every(15 * time.Minute)
}

// EveryThirtyMinutes creates a schedule that runs every 30 minutes
func EveryThirtyMinutes() contract.Schedule {
	return Every(30 * time.Minute)
}

// Hourly creates a schedule that runs every hour
func Hourly() contract.Schedule {
	return Every(time.Hour)
}

// HourlyAt creates a schedule that runs every hour at a specific minute
//
// Example:
//
//	HourlyAt(15)  // Every hour at :15
func HourlyAt(minute int) contract.Schedule {
	if minute < 0 || minute > 59 {
		minute = 0
	}
	return Cron(fmt.Sprintf("%d * * * *", minute))
}

// Daily creates a schedule that runs once per day at midnight
func Daily() contract.Schedule {
	return &dailySchedule{hour: 0, minute: 0}
}

// DailyAt creates a schedule that runs once per day at a specific time
//
// Examples:
//
//	DailyAt(9, 0)    // Every day at 9:00 AM
//	DailyAt(17, 30)  // Every day at 5:30 PM
func DailyAt(hour, minute int) contract.Schedule {
	if hour < 0 || hour > 23 {
		hour = 0
	}
	if minute < 0 || minute > 59 {
		minute = 0
	}
	return &dailySchedule{hour: hour, minute: minute}
}

// TwiceDaily creates a schedule that runs twice per day
//
// Example:
//
//	TwiceDaily(8, 20)  // At 8:00 AM and 8:00 PM
func TwiceDaily(hour1, hour2 int) contract.Schedule {
	return Cron(fmt.Sprintf("0 %d,%d * * *", hour1, hour2))
}

// Weekly creates a schedule that runs once per week on Sunday at midnight
func Weekly() contract.Schedule {
	return &weeklySchedule{
		weekdays: []time.Weekday{time.Sunday},
		hour:     0,
		minute:   0,
	}
}

// WeeklyOn creates a schedule that runs on specific days of the week
//
// Examples:
//
//	WeeklyOn(time.Monday, 9, 0)                          // Mondays at 9:00 AM
//	WeeklyOn(time.Monday, time.Wednesday, time.Friday)  // Mon/Wed/Fri at midnight
func WeeklyOn(weekday time.Weekday, hourAndMinute ...int) contract.Schedule {
	hour, minute := 0, 0
	if len(hourAndMinute) >= 1 {
		hour = hourAndMinute[0]
	}
	if len(hourAndMinute) >= 2 {
		minute = hourAndMinute[1]
	}
	return &weeklySchedule{
		weekdays: []time.Weekday{weekday},
		hour:     hour,
		minute:   minute,
	}
}

// Weekdays creates a schedule that runs Monday through Friday
//
// Example:
//
//	Weekdays(9, 0)  // Monday-Friday at 9:00 AM
func Weekdays(hour, minute int) contract.Schedule {
	return &weeklySchedule{
		weekdays: []time.Weekday{
			time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday,
		},
		hour:   hour,
		minute: minute,
	}
}

// Weekends creates a schedule that runs on Saturday and Sunday
//
// Example:
//
//	Weekends(10, 0)  // Saturday and Sunday at 10:00 AM
func Weekends(hour, minute int) contract.Schedule {
	return &weeklySchedule{
		weekdays: []time.Weekday{time.Saturday, time.Sunday},
		hour:     hour,
		minute:   minute,
	}
}

// Monthly creates a schedule that runs on the first day of each month at midnight
func Monthly() contract.Schedule {
	return &monthlySchedule{days: []int{1}, hour: 0, minute: 0}
}

// MonthlyOn creates a schedule that runs on specific days of the month
//
// Examples:
//
//	MonthlyOn(1, 9, 0)     // 1st of each month at 9:00 AM
//	MonthlyOn(15, 12, 0)   // 15th of each month at noon
//	MonthlyOn(-1, 18, 0)   // Last day of each month at 6:00 PM
func MonthlyOn(day, hour, minute int) contract.Schedule {
	return &monthlySchedule{days: []int{day}, hour: hour, minute: minute}
}

// LastDayOfMonth creates a schedule that runs on the last day of each month
//
// Example:
//
//	LastDayOfMonth(17, 0)  // Last day of month at 5:00 PM
func LastDayOfMonth(hour, minute int) contract.Schedule {
	return MonthlyOn(-1, hour, minute)
}

// Quarterly creates a schedule that runs on the first day of each quarter
//
// Example:
//
//	Quarterly(9, 0)  // Jan 1, Apr 1, Jul 1, Oct 1 at 9:00 AM
func Quarterly(hour, minute int) contract.Schedule {
	return Cron(fmt.Sprintf("%d %d 1 1,4,7,10 *", minute, hour))
}

// Yearly creates a schedule that runs on January 1st at midnight
func Yearly() contract.Schedule {
	return Cron("0 0 1 1 *")
}

// YearlyOn creates a schedule that runs on a specific date each year
//
// Example:
//
//	YearlyOn(time.March, 15, 9, 0)  // March 15th at 9:00 AM
func YearlyOn(month time.Month, day, hour, minute int) contract.Schedule {
	return Cron(fmt.Sprintf("%d %d %d %d *", minute, hour, day, month))
}

// Cron creates a schedule from a cron expression
//
// Supports standard 5-field cron expressions:
//
//	┌───────────── minute (0-59)
//	│ ┌───────────── hour (0-23)
//	│ │ ┌───────────── day of month (1-31)
//	│ │ │ ┌───────────── month (1-12)
//	│ │ │ │ ┌───────────── day of week (0-6, Sunday=0)
//	│ │ │ │ │
//	* * * * *
//
// Examples:
//
//	Cron("0 0 * * *")      // Every day at midnight
//	Cron("*/5 * * * *")    // Every 5 minutes
//	Cron("0 9 * * 1-5")    // Weekdays at 9 AM
func Cron(expression string) contract.Schedule {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(expression)
	if err != nil {
		// Return a schedule that never runs on error
		return &neverSchedule{err: err, expr: expression}
	}
	return &cronSchedule{
		expr:     expression,
		schedule: schedule,
	}
}

// neverSchedule is returned when a cron expression is invalid
type neverSchedule struct {
	err  error
	expr string
}

func (s *neverSchedule) Next(after time.Time) time.Time {
	// Return a time far in the future (effectively never)
	return time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
}

func (s *neverSchedule) String() string {
	return fmt.Sprintf("invalid cron: %s (error: %v)", s.expr, s.err)
}

func (s *neverSchedule) Cron() string {
	return s.expr
}

// =============================================================================
// Schedule Combinators
// =============================================================================

// Between wraps a schedule to only run between specific hours
type betweenSchedule struct {
	inner     contract.Schedule
	startHour int
	endHour   int
}

func (s *betweenSchedule) Next(after time.Time) time.Time {
	for i := 0; i < 366; i++ { // Check up to a year
		next := s.inner.Next(after)
		hour := next.Hour()
		if hour >= s.startHour && hour < s.endHour {
			return next
		}
		after = next
	}
	return time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
}

func (s *betweenSchedule) String() string {
	return fmt.Sprintf("%s (between %02d:00 and %02d:00)", s.inner.String(), s.startHour, s.endHour)
}

func (s *betweenSchedule) Cron() string {
	return s.inner.Cron()
}

// Between restricts a schedule to only run between specific hours
//
// Example:
//
//	Between(Every(15 * time.Minute), 9, 17)  // Every 15 min, but only 9 AM - 5 PM
func Between(schedule contract.Schedule, startHour, endHour int) contract.Schedule {
	return &betweenSchedule{
		inner:     schedule,
		startHour: startHour,
		endHour:   endHour,
	}
}

// Skip wraps a schedule to skip certain days
type skipDaysSchedule struct {
	inner    contract.Schedule
	skipDays []time.Weekday
}

func (s *skipDaysSchedule) Next(after time.Time) time.Time {
	for i := 0; i < 366; i++ {
		next := s.inner.Next(after)
		skip := false
		for _, day := range s.skipDays {
			if next.Weekday() == day {
				skip = true
				break
			}
		}
		if !skip {
			return next
		}
		after = next
	}
	return time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
}

func (s *skipDaysSchedule) String() string {
	days := make([]string, len(s.skipDays))
	for i, d := range s.skipDays {
		days[i] = d.String()
	}
	return fmt.Sprintf("%s (skip %s)", s.inner.String(), strings.Join(days, ", "))
}

func (s *skipDaysSchedule) Cron() string {
	return s.inner.Cron()
}

// SkipWeekends restricts a schedule to skip Saturday and Sunday
//
// Example:
//
//	SkipWeekends(Daily())  // Every day except weekends
func SkipWeekends(schedule contract.Schedule) contract.Schedule {
	return &skipDaysSchedule{
		inner:    schedule,
		skipDays: []time.Weekday{time.Saturday, time.Sunday},
	}
}
