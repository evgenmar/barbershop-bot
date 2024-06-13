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

func (d Duration) String() string {
	hours := int(d / 60)
	minutes := int(d) - hours*60
	return fmt.Sprintf("%02d:%02d", hours, minutes)
}

func Today() time.Time {
	return time.Date(
		time.Now().In(cfg.Location).Year(),
		time.Now().In(cfg.Location).Month(),
		time.Now().In(cfg.Location).Day(),
		0, 0, 0, 0, cfg.Location,
	)
}