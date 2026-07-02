package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"langs/internal/domain"
)

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02 15:04", value)
	assert.NoError(t, err)
	return parsed
}

func TestIsInQuietHoursAt(t *testing.T) {
	newUser := func(enabled bool, start, end, days string) *model.User {
		return &model.User{
			Timezone: "UTC",
			QuietHours: model.QuietHours{
				Enabled:    enabled,
				StartTime:  start,
				EndTime:    end,
				DaysOfWeek: days,
			},
		}
	}

	// 2026-07-01 is a Wednesday (ISO day 3).
	t.Run("DisabledNeverQuiet", func(t *testing.T) {
		user := newUser(false, "00:00", "23:59", "1,2,3,4,5,6,7")
		assert.False(t, user.IsInQuietHoursAt(mustTime(t, "2026-07-01 12:00")))
	})

	t.Run("OvernightWindowEvening", func(t *testing.T) {
		user := newUser(true, "22:00", "08:00", "1,2,3,4,5,6,7")
		assert.True(t, user.IsInQuietHoursAt(mustTime(t, "2026-07-01 23:30")))
	})

	t.Run("OvernightWindowMorning", func(t *testing.T) {
		user := newUser(true, "22:00", "08:00", "1,2,3,4,5,6,7")
		assert.True(t, user.IsInQuietHoursAt(mustTime(t, "2026-07-01 07:00")))
	})

	t.Run("OvernightWindowDaytimeIsNotQuiet", func(t *testing.T) {
		user := newUser(true, "22:00", "08:00", "1,2,3,4,5,6,7")
		assert.False(t, user.IsInQuietHoursAt(mustTime(t, "2026-07-01 12:00")))
	})

	t.Run("SameDayWindow", func(t *testing.T) {
		user := newUser(true, "14:00", "16:00", "1,2,3,4,5,6,7")
		assert.True(t, user.IsInQuietHoursAt(mustTime(t, "2026-07-01 15:00")))
		assert.False(t, user.IsInQuietHoursAt(mustTime(t, "2026-07-01 17:00")))
		assert.False(t, user.IsInQuietHoursAt(mustTime(t, "2026-07-01 13:59")))
	})

	t.Run("DayNotSelected", func(t *testing.T) {
		// Only Monday (1) selected, but the date is Wednesday.
		user := newUser(true, "14:00", "16:00", "1")
		assert.False(t, user.IsInQuietHoursAt(mustTime(t, "2026-07-01 15:00")))
	})

	t.Run("OvernightMorningBelongsToPreviousDay", func(t *testing.T) {
		// Quiet hours only on Tuesday (2): Tuesday 22:00 - Wednesday 08:00.
		user := newUser(true, "22:00", "08:00", "2")
		assert.True(t, user.IsInQuietHoursAt(mustTime(t, "2026-07-01 07:00")), "Wednesday morning belongs to Tuesday's window")
		assert.False(t, user.IsInQuietHoursAt(mustTime(t, "2026-07-01 23:00")), "Wednesday evening is not selected")
	})

	t.Run("InvalidTimeIsNeverQuiet", func(t *testing.T) {
		user := newUser(true, "invalid", "08:00", "1,2,3,4,5,6,7")
		assert.False(t, user.IsInQuietHoursAt(mustTime(t, "2026-07-01 07:00")))
	})
}

func TestQuietHoursToggleDay(t *testing.T) {
	t.Run("RemoveDay", func(t *testing.T) {
		q := model.QuietHours{DaysOfWeek: "1,2,3"}
		q.ToggleDay(2)
		assert.Equal(t, "1,3", q.DaysOfWeek)
	})

	t.Run("AddDay", func(t *testing.T) {
		q := model.QuietHours{DaysOfWeek: "1,3"}
		q.ToggleDay(2)
		assert.Equal(t, "1,2,3", q.DaysOfWeek)
	})

	t.Run("ToggleLastDay", func(t *testing.T) {
		q := model.QuietHours{DaysOfWeek: "5"}
		q.ToggleDay(5)
		assert.Equal(t, "", q.DaysOfWeek)
		assert.False(t, q.HasDay(5))
	})

	t.Run("IgnoresInvalidDay", func(t *testing.T) {
		q := model.QuietHours{DaysOfWeek: "1,2"}
		q.ToggleDay(9)
		assert.Equal(t, "1,2", q.DaysOfWeek)
	})

	t.Run("NormalizesCorruptedList", func(t *testing.T) {
		q := model.QuietHours{DaysOfWeek: "1,,3, 2"}
		q.ToggleDay(4)
		assert.Equal(t, "1,2,3,4", q.DaysOfWeek)
	})
}

func TestGetTestIntervalHours(t *testing.T) {
	t.Run("ValidInterval", func(t *testing.T) {
		user := &model.User{TestInterval: 5}
		assert.Equal(t, 5, user.GetTestIntervalHours())
	})

	t.Run("ZeroFallsBackToDefault", func(t *testing.T) {
		user := &model.User{TestInterval: 0}
		assert.Equal(t, model.DefaultTestIntervalHours, user.GetTestIntervalHours())
	})

	t.Run("LegacyValueFallsBackToDefault", func(t *testing.T) {
		// Legacy rows stored 60 (minutes); it is out of the 1-8 range.
		user := &model.User{TestInterval: 60}
		assert.Equal(t, model.DefaultTestIntervalHours, user.GetTestIntervalHours())
	})
}

func TestIsTestDue(t *testing.T) {
	now := mustTime(t, "2026-07-01 12:00")

	t.Run("NeverSentIsDue", func(t *testing.T) {
		user := &model.User{TestInterval: 3}
		assert.True(t, user.IsTestDue(now))
	})

	t.Run("NotEnoughTimePassed", func(t *testing.T) {
		user := &model.User{TestInterval: 3, LastTestSentAt: now.Add(-2 * time.Hour)}
		assert.False(t, user.IsTestDue(now))
	})

	t.Run("EnoughTimePassed", func(t *testing.T) {
		user := &model.User{TestInterval: 3, LastTestSentAt: now.Add(-3 * time.Hour)}
		assert.True(t, user.IsTestDue(now))
	})
}

func TestGetCurrentTimeInTimezone(t *testing.T) {
	t.Run("ValidTimezone", func(t *testing.T) {
		user := &model.User{Timezone: "Europe/London"}
		assert.NotZero(t, user.GetCurrentTimeInTimezone())
	})

	t.Run("InvalidTimezoneFallsBackToUTC", func(t *testing.T) {
		user := &model.User{Timezone: "Not/AZone"}
		current := user.GetCurrentTimeInTimezone()
		assert.Equal(t, time.UTC, current.Location())
	})
}
