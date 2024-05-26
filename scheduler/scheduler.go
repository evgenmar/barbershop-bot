package scheduler

import (
	"barbershop-bot/config"
	"barbershop-bot/lib/e"
	"barbershop-bot/storage"
	"context"
	"errors"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

// CronWithSettings creates *cron.Cron and set it with several functions to be run on a schedule.
// Triggered events:
//   - making schedules for barbers - every Monday at 03:00 AM
func CronWithSettings(rep storage.Storage, barberIDs []int64) *cron.Cron {
	crn := cron.New(cron.WithLocation(config.Location))
	crn.AddFunc("0 3 * * 1",
		func() {
			err := MakeSchedules(rep, barberIDs, config.ScheduledDays)
			if err != nil {
				log.Print(err)
			}
		})
	return crn
}

// MakeSchedules just calls makeSchedule for all barbers specified in barberIDs.
// See makeSchedule for details.
func MakeSchedules(rep storage.Storage, barberIDs []int64, days uint16) error {
	for _, barberID := range barberIDs {
		err := makeSchedule(rep, barberID, days)
		if err != nil {
			return e.Wrap("can't make schedules", err)
		}
	}
	return nil
}

// makeSchedule make a working schedule for barber with barberID and saves it.
//
// The schedule is created for a period starting from the day the function is called for the next days days.
// If a saved schedule already exists in the storage for some (or all) days from the specified period
// then it remains valid. The function creates a schedule only for those days from the specified period
// for which there was no schedule.
//
// Mondays are accepted as non-working days. On other days the working time is from 10:00 to 19:00.
func makeSchedule(rep storage.Storage, barberID int64, days uint16) (err error) {
	defer func() { err = e.WrapIfErr("can't make schedule", err) }()

	latestWD, err := rep.GetLatestWorkDate(context.TODO(), barberID)
	if err != nil && !errors.Is(err, storage.ErrNoSavedWorkdates) {
		return err
	}
	latestWorkDate, err := time.ParseInLocation(time.DateOnly, latestWD, config.Location)
	if err != nil {
		return err
	}
	dayDuration := 24 * time.Hour
	today := time.Now().In(config.Location).Truncate(dayDuration)
	for date := today.Add(time.Duration(days) * dayDuration); date.Compare(today) >= 0; date = date.Add(-dayDuration) {
		if date.Compare(latestWorkDate) == 1 {
			if date.Weekday() != time.Monday {
				workday := storage.Workday{
					BarberID:  barberID,
					Date:      date.Format(time.DateOnly),
					StartTime: "10:00",
					EndTime:   "19:00",
				}
				err := rep.CreateWorkday(context.TODO(), workday)
				if err != nil {
					return err
				}
			}
		} else {
			return nil
		}
	}
	return nil
}
