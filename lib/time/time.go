package time

import (
	"encoding/binary"
	"fmt"
	"time"

	cfg "github.com/evgenmar/barbershop-bot/lib/config"
)

// A Duration represents the elapsed time between two instants as an int16 minute count.
// Duration is designed to measure time intervals within one day.
type Duration int16

// Month represents the number of the month starting from January 2001
type Month int16

const (
	Minute Duration = 1
	Hour            = 60 * Minute
)

func (d Duration) Bytes() []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, uint16(d))
	return bytes
}

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

func (d Duration) RoundUpToMultipleOf30() Duration {
	if d%30 == 0 {
		return d
	}
	return (d/30 + 1) * 30
}

func (d Duration) ShortString() string {
	if d == 0 {
		return ""
	}
	hours := int(d / 60)
	minutes := int(d) - hours*60
	return fmt.Sprintf("%02d:%02d", hours, minutes)
}

func CurrentDayTime() Duration {
	now := time.Now().In(cfg.Location)
	return Hour*Duration(now.Hour()) + Minute*Duration(now.Minute())
}

func MonthFromNow(deltaMonth byte) Month {
	return ParseMonth(time.Now().In(cfg.Location)) + Month(deltaMonth)
}

func ParseMonth(t time.Time) Month {
	return (Month(t.Year())-2001)*12 + Month(t.Month())
}

func ShowDate(date time.Time) string {
	return date.Format("02.01.2006")
}

func Today() time.Time {
	year, month, day := time.Now().In(cfg.Location).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, cfg.Location)
}
