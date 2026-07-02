package model

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

const CustomTranslationScenario = "cts"
const WritingExamScenario = "wes"
const WordSearchScenario = "wss"

// Bounds for how often (in hours) scheduled word tests are sent to a user.
const (
	MinTestIntervalHours     = 1
	MaxTestIntervalHours     = 8
	DefaultTestIntervalHours = 3
)

type User struct {
	TestInterval   uint16     `json:"test_interval,omitempty" gorm:"default:3"`
	TargetRate     uint16     `json:"target_rate,omitempty" gorm:"default:5"`
	ChatId         int64      `json:"chat_id" gorm:"primaryKey"`
	BookId         int64      `json:"book_id"`
	InterfaceLang  string     `json:"interface_lang,omitempty"`
	NativeLang     string     `json:"native_lang,omitempty"`
	TargetLang     string     `json:"target_lang,omitempty"`
	Timezone       string     `json:"timezone,omitempty" gorm:"default:'UTC'"`
	LastTestSentAt time.Time  `json:"last_test_sent_at,omitempty"`
	QuietHours     QuietHours `gorm:"embedded"`
	StateData      StateData  `gorm:"embedded"`
}

type QuietHours struct {
	Enabled    bool   `json:"enabled" gorm:"default:false"`
	StartTime  string `json:"start_time,omitempty" gorm:"default:'22:00'"`
	EndTime    string `json:"end_time,omitempty" gorm:"default:'08:00'"`
	DaysOfWeek string `json:"days_of_week,omitempty" gorm:"default:'1,2,3,4,5,6,7'"`
}

type StateData struct {
	Scenario string `json:"scenario,omitempty"`
	Value    string `json:"value,omitempty"`
}

func (u *User) GetUserLangs() []string {
	return []string{u.NativeLang, u.TargetLang}
}

func (u *User) IsAwaitingInput() bool {
	return u.StateData.Scenario != ""
}

func (u *User) GetOppositeLang(lang string) string {
	if u.NativeLang == lang {
		return u.TargetLang
	}
	return u.NativeLang
}

// GetTestIntervalHours returns the configured test interval clamped to the
// supported 1-8 hour range, falling back to the default for invalid values
// (including legacy rows that stored other units).
func (u *User) GetTestIntervalHours() int {
	interval := int(u.TestInterval)
	if interval < MinTestIntervalHours || interval > MaxTestIntervalHours {
		return DefaultTestIntervalHours
	}
	return interval
}

// IsTestDue reports whether enough time has passed since the last scheduled
// test for this user to receive a new one.
func (u *User) IsTestDue(now time.Time) bool {
	if u.LastTestSentAt.IsZero() {
		return true
	}
	interval := time.Duration(u.GetTestIntervalHours()) * time.Hour
	return now.Sub(u.LastTestSentAt) >= interval
}

func (u *User) IsInQuietHours() bool {
	return u.IsInQuietHoursAt(u.GetCurrentTimeInTimezone())
}

// IsInQuietHoursAt reports whether t falls into the user's quiet-hours window.
// Supports same-day windows (14:00-16:00) and overnight windows (22:00-08:00).
// For overnight windows the after-midnight part belongs to the day the window started.
func (u *User) IsInQuietHoursAt(t time.Time) bool {
	if !u.QuietHours.Enabled {
		return false
	}

	start, ok := parseMinutes(u.QuietHours.StartTime)
	if !ok {
		return false
	}
	end, ok := parseMinutes(u.QuietHours.EndTime)
	if !ok {
		return false
	}

	current := t.Hour()*60 + t.Minute()
	day := isoWeekday(t)

	if start <= end {
		return current >= start && current <= end && u.QuietHours.HasDay(day)
	}

	// Overnight window: evening part belongs to the current day,
	// morning part belongs to the previous day (when the window started).
	if current >= start {
		return u.QuietHours.HasDay(day)
	}
	if current <= end {
		return u.QuietHours.HasDay(previousDay(day))
	}
	return false
}

func (u *User) GetCurrentTimeInTimezone() time.Time {
	loc, err := time.LoadLocation(u.Timezone)
	if err != nil {
		loc = time.UTC
	}
	return time.Now().In(loc)
}

// HasDay reports whether the given ISO day (1=Mon .. 7=Sun) is selected.
func (q *QuietHours) HasDay(day int) bool {
	for _, d := range q.days() {
		if d == day {
			return true
		}
	}
	return false
}

// ToggleDay adds or removes an ISO day (1=Mon .. 7=Sun) from the selection.
func (q *QuietHours) ToggleDay(day int) {
	if day < 1 || day > 7 {
		return
	}

	days := q.days()
	filtered := days[:0]
	removed := false
	for _, d := range days {
		if d == day {
			removed = true
			continue
		}
		filtered = append(filtered, d)
	}
	if !removed {
		filtered = append(filtered, day)
	}

	sort.Ints(filtered)
	q.setDays(filtered)
}

func (q *QuietHours) days() []int {
	var result []int
	for _, part := range strings.Split(q.DaysOfWeek, ",") {
		d, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || d < 1 || d > 7 {
			continue
		}
		result = append(result, d)
	}
	return result
}

func (q *QuietHours) setDays(days []int) {
	parts := make([]string, 0, len(days))
	for _, d := range days {
		parts = append(parts, strconv.Itoa(d))
	}
	q.DaysOfWeek = strings.Join(parts, ",")
}

func parseMinutes(value string) (int, bool) {
	t, err := time.Parse("15:04", strings.TrimSpace(value))
	if err != nil {
		return 0, false
	}
	return t.Hour()*60 + t.Minute(), true
}

// isoWeekday converts time.Weekday (Sunday=0) to ISO (Monday=1 .. Sunday=7).
func isoWeekday(t time.Time) int {
	day := int(t.Weekday())
	if day == 0 {
		return 7
	}
	return day
}

func previousDay(day int) int {
	if day == 1 {
		return 7
	}
	return day - 1
}
