package idle

// Tests for shouldExecuteCustom — verifies that the custom schedule type correctly
// evaluates StartTime/EndTime windows, DaysOfWeek filters, and timezone handling.

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// newCustomSchedule builds a minimal Schedule of type custom for testing.
func newCustomSchedule(startTime, endTime, timezone string, days ...DayOfWeek) *Schedule {
	return &Schedule{
		ID:         "test-custom",
		Name:       "test",
		Type:       ScheduleTypeCustom,
		Enabled:    true,
		StartTime:  startTime,
		EndTime:    endTime,
		Timezone:   timezone,
		DaysOfWeek: days,
	}
}

// newTestScheduler returns a minimal Scheduler sufficient for shouldExecuteCustom tests.
func newTestScheduler() *Scheduler {
	return &Scheduler{
		active: make(map[string]*ScheduleExecution),
	}
}

// TestShouldExecuteCustom_InWindow verifies that the schedule fires when the current
// time falls within the configured [StartTime, EndTime) window.
func TestShouldExecuteCustom_InWindow(t *testing.T) {
	s := newTestScheduler()
	schedule := newCustomSchedule("09:00", "17:00", "")

	// 12:00 UTC — inside the 09:00–17:00 window
	now := time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC) // Monday
	assert.True(t, s.shouldExecuteCustom(schedule, now))
}

// TestShouldExecuteCustom_OutOfWindow verifies that the schedule does not fire when the
// current time is outside the configured window.
func TestShouldExecuteCustom_OutOfWindow(t *testing.T) {
	s := newTestScheduler()
	schedule := newCustomSchedule("09:00", "17:00", "")

	// 20:00 UTC — outside the window
	now := time.Date(2026, 3, 30, 20, 0, 0, 0, time.UTC)
	assert.False(t, s.shouldExecuteCustom(schedule, now))
}

// TestShouldExecuteCustom_WrongDay verifies that the schedule does not fire when the
// current day is not in DaysOfWeek.
func TestShouldExecuteCustom_WrongDay(t *testing.T) {
	s := newTestScheduler()
	// Only on Monday; 2026-03-31 is a Tuesday
	schedule := newCustomSchedule("09:00", "17:00", "", Monday)

	now := time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC) // Tuesday
	assert.False(t, s.shouldExecuteCustom(schedule, now))
}

// TestShouldExecuteCustom_CorrectDay verifies that the schedule fires when the current day
// matches one of the configured DaysOfWeek and the time is in-window.
func TestShouldExecuteCustom_CorrectDay(t *testing.T) {
	s := newTestScheduler()
	// Monday and Wednesday; 2026-03-30 is a Monday
	schedule := newCustomSchedule("09:00", "17:00", "", Monday, Wednesday)

	now := time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC) // Monday, 12:00
	assert.True(t, s.shouldExecuteCustom(schedule, now))
}

// TestShouldExecuteCustom_NoConstraints verifies that a custom schedule with no time
// window or day restriction always returns true.
func TestShouldExecuteCustom_NoConstraints(t *testing.T) {
	s := newTestScheduler()
	schedule := newCustomSchedule("", "", "")

	now := time.Date(2026, 3, 30, 3, 0, 0, 0, time.UTC) // 3 AM Sunday
	assert.True(t, s.shouldExecuteCustom(schedule, now))
}

// TestShouldExecuteCustom_Timezone verifies that the time window is evaluated in the
// configured timezone, not UTC.
func TestShouldExecuteCustom_Timezone(t *testing.T) {
	s := newTestScheduler()
	// Window 09:00–17:00 in US/Eastern (UTC-4 during DST)
	schedule := newCustomSchedule("09:00", "17:00", "America/New_York")

	// 14:00 UTC = 10:00 Eastern → inside window
	inWindow := time.Date(2026, 3, 30, 14, 0, 0, 0, time.UTC)
	assert.True(t, s.shouldExecuteCustom(schedule, inWindow))

	// 22:00 UTC = 18:00 Eastern → outside window
	outOfWindow := time.Date(2026, 3, 30, 22, 0, 0, 0, time.UTC)
	assert.False(t, s.shouldExecuteCustom(schedule, outOfWindow))
}
