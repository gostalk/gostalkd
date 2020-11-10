package gtime

import (
	"time"
)

// TimeWrapper is a wrapper for stdlib struct time.Time.
// It's used for overwriting some functions of time.Time, for example: String.
type TimeWrapper struct {
	time.Time
}

// String overwrites the String function of time.Time.
func (t TimeWrapper) String() string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
