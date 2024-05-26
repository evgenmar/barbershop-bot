package config

import "time"

const (
	SqliteStoragePath = "data/sqlite/storage.db"

	//scheduledDays is the number of days for which the barbershop schedule is compiled.
	ScheduledDays uint16 = 183
)

// location is the time zone where the barbershop is located.
var Location *time.Location

func init() {
	Location = time.FixedZone("MSK", 3*60*60)
	_ = Location
	_ = SqliteStoragePath
	_ = ScheduledDays
}
