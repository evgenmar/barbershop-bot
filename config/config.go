package config

import "time"

const (
	SqliteStoragePath = "data/sqlite/storage.db"

	//scheduledWeeks is the number of weeks for which the barbershop schedule is compiled.
	ScheduledWeeks uint8 = 26

	DbQueryTimoutWrite time.Duration = 3 * time.Second
	DbQueryTimoutRead  time.Duration = 2 * time.Second
)

// location is the time zone where the barbershop is located.
var Location *time.Location

func init() {
	Location = time.FixedZone("MSK", 3*60*60)
}
