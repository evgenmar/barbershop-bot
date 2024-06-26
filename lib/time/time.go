package time

import (
	cfg "barbershop-bot/lib/config"
	"fmt"
	"time"
)

// A Duration represents the elapsed time between two instants as an int16 minute count.
// Duration is designed to measure time intervals within one day.
type Duration int16

const (
	Minute Duration = 1
	Hour            = 60 * Minute
)

func (d Duration) LongString() string {
	if d == 0 {
		return ""
	}
	hours := int(d / 60)
	minutes := int(d) - hours*60
	if hours == 0 {
		return fmt.Sprintf("%02d мин", minutes)
	}
	if minutes == 0 {
		return fmt.Sprintf("%01d ч", hours)
	}
	return fmt.Sprintf("%01d ч %02d мин", hours, minutes)
}

func ParseDuration(str string) (Duration, error) {
	var hours, minutes int16
	_, err := fmt.Sscanf(str, "%d:%d", &hours, &minutes)
	if err != nil {
		return 0, err
	}
	return Duration(hours*60 + minutes), nil
}

func (d Duration) ShortString() string {
	if d == 0 {
		return ""
	}
	hours := int(d / 60)
	minutes := int(d) - hours*60
	return fmt.Sprintf("%02d:%02d", hours, minutes)
}

func Today() time.Time {
	now := time.Now().In(cfg.Location)
	year, month, day := now.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, cfg.Location)
}
