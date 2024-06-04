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
func CronWithSettings(rep storage.Storage) *cron.Cron {
	crn := cron.New(cron.WithLocation(config.Location))
	crn.AddFunc("0 3 * * 1",
		func() {
			err := MakeSchedules(rep)
			if err != nil {
				log.Print(err)
			}
		})
	return crn
}

// MakeSchedules just calls makeSchedule for all barbers specified in barberIDs.
// See makeSchedule for details.
func MakeSchedules(rep storage.Storage) (err error) {
	defer func() { err = e.WrapIfErr("can't make schedules", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutRead)
	barberIDs, err := rep.FindAllBarberIDs(ctx)
	cancel()
	if err != nil {
		return err
	}
	for _, barberID := range barberIDs {
		err := MakeSchedule(rep, barberID)
		if err != nil {
			return err
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
func MakeSchedule(rep storage.Storage, barberID int64) (err error) {
	defer func() { err = e.WrapIfErr("can't make schedule", err) }()
	latestWorkDate, err := getLatestWorkDate(rep, barberID)
	if err != nil {
		return err
	}
	var workdays []storage.Workday
	dayDuration := 24 * time.Hour
	today := today()
	for date := today.Add(time.Duration(config.ScheduledWeeks) * dayDuration * 7); date.Compare(today) >= 0 && date.After(latestWorkDate); date = date.Add(-dayDuration) {
		if date.Weekday() != time.Monday {
			workdays = append(workdays, storage.Workday{
				BarberID:  barberID,
				Date:      date.Format(time.DateOnly),
				StartTime: "10:00",
				EndTime:   "19:00",
			})
		}
	}
	if len(workdays) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutWrite)
		err = rep.CreateWorkdays(ctx, workdays...)
		cancel()
		if err != nil {
			return err
		}
	}
	return nil
}

// getLatestWorkDate returns the latest working date existing in storage
func getLatestWorkDate(rep storage.Storage, barberID int64) (latestWorkDate time.Time, err error) {
	defer func() { err = e.WrapIfErr("can't get latest work date", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutRead)
	latestWD, err := rep.GetLatestWorkDate(ctx, barberID)
	cancel()
	if err != nil && !errors.Is(err, storage.ErrNoSavedWorkdates) {
		return time.Time{}, err
	}
	latestWorkDate, err = time.ParseInLocation(time.DateOnly, latestWD, config.Location)
	if err != nil {
		return time.Time{}, err
	}
	return latestWorkDate, nil
}

// latestScheduledDate returns the latest date for which the schedule should be made
func today() time.Time {
	return time.Date(
		time.Now().In(config.Location).Year(),
		time.Now().In(config.Location).Month(),
		time.Now().In(config.Location).Day(),
		0, 0, 0, 0, config.Location,
	)
}
