package services

import (
	"fmt"
	"time"
)

// UserTimeHelper helps determine the user's local time
type UserTimeHelper struct {
	timezoneService *TimezoneLocationService
}

// NewUserTimeHelper creates a new user time helper
func NewUserTimeHelper() *UserTimeHelper {
	return &UserTimeHelper{
		timezoneService: NewTimezoneLocationService(),
	}
}

// GetUserLocalTime returns the user's local time based on their timezone
func (uth *UserTimeHelper) GetUserLocalTime(deviceTimezone string) time.Time {
	// Load the timezone location
	loc, err := time.LoadLocation(deviceTimezone)
	if err != nil {
		// Fallback to server time if timezone is invalid
		fmt.Printf("[USER_TIME] Failed to load timezone %s: %v, using server time\n", deviceTimezone, err)
		return time.Now()
	}

	// Return current time in user's timezone
	userTime := time.Now().In(loc)
	fmt.Printf("[USER_TIME] User timezone: %s, Local time: %s\n", deviceTimezone, userTime.Format("15:04:05"))
	return userTime
}

// GetUserLocalHour returns the user's current hour (0-23)
func (uth *UserTimeHelper) GetUserLocalHour(deviceTimezone string) int {
	userTime := uth.GetUserLocalTime(deviceTimezone)
	return userTime.Hour()
}

// GetNatureSoundForUserTime selects appropriate nature sound based on user's local time
func (uth *UserTimeHelper) GetNatureSoundForUserTime(deviceTimezone string) string {
	hour := uth.GetUserLocalHour(deviceTimezone)

	// Select nature sound based on user's local hour
	switch {
	case hour >= 5 && hour < 9:
		// Early morning (5am-9am)
		return "morning_birds"
	case hour >= 9 && hour < 12:
		// Late morning (9am-12pm)
		return "forest"
	case hour >= 12 && hour < 17:
		// Afternoon (12pm-5pm)
		return "meadow"
	case hour >= 17 && hour < 20:
		// Evening (5pm-8pm)
		return "gentle_rain"
	case hour >= 20 && hour < 22:
		// Late evening (8pm-10pm)
		return "stream"
	default:
		// Night (10pm-5am)
		return "night"
	}
}

// GetTimeOfDayGreeting returns a greeting based on user's local time
func (uth *UserTimeHelper) GetTimeOfDayGreeting(deviceTimezone string) string {
	hour := uth.GetUserLocalHour(deviceTimezone)

	switch {
	case hour >= 5 && hour < 12:
		return "Good morning"
	case hour >= 12 && hour < 17:
		return "Good afternoon"
	case hour >= 17 && hour < 21:
		return "Good evening"
	default:
		return "Hello"
	}
}

// IsUserDaytime returns true if it's daytime for the user (6am-8pm)
func (uth *UserTimeHelper) IsUserDaytime(deviceTimezone string) bool {
	hour := uth.GetUserLocalHour(deviceTimezone)
	return hour >= 6 && hour < 20
}

// GetUserTimeContext provides context about the user's time of day
func (uth *UserTimeHelper) GetUserTimeContext(deviceTimezone string) map[string]interface{} {
	userTime := uth.GetUserLocalTime(deviceTimezone)
	hour := userTime.Hour()

	return map[string]interface{}{
		"timezone":     deviceTimezone,
		"local_time":   userTime.Format("15:04:05"),
		"hour":         hour,
		"greeting":     uth.GetTimeOfDayGreeting(deviceTimezone),
		"is_daytime":   uth.IsUserDaytime(deviceTimezone),
		"nature_sound": uth.GetNatureSoundForUserTime(deviceTimezone),
		"time_period":  getTimePeriod(hour),
	}
}

// getTimePeriod returns a descriptive time period
func getTimePeriod(hour int) string {
	switch {
	case hour >= 5 && hour < 7:
		return "early_morning"
	case hour >= 7 && hour < 9:
		return "morning"
	case hour >= 9 && hour < 12:
		return "late_morning"
	case hour >= 12 && hour < 14:
		return "midday"
	case hour >= 14 && hour < 17:
		return "afternoon"
	case hour >= 17 && hour < 19:
		return "early_evening"
	case hour >= 19 && hour < 21:
		return "evening"
	case hour >= 21 && hour < 23:
		return "late_evening"
	default:
		return "night"
	}
}
