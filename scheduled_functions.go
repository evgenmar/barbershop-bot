package main

import (
	"barbershop-bot/lib/e"
	"barbershop-bot/storage"
	"context"
	"errors"
	"time"
)

// makeSchedules just calls makeSchedule for all barbers specified in barberIDs.
// See makeSchedule for details.
func makeSchedules(rep storage.Storage, barberIDs []int64, days uint16) error {
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

	LatestWorkDate, err := rep.LatestWorkDate(context.TODO(), barberID)
	if err != nil && !errors.Is(err, storage.ErrNoSavedWorkdates) {
		return err
	}
	dayDuration := 24 * time.Hour
	today := time.Now().In(location).Truncate(dayDuration)
	for date := today.Add(time.Duration(days) * dayDuration); date.Compare(today) >= 0; date = date.Add(-dayDuration) {
		if date.Compare(LatestWorkDate) == 1 {
			if date.Weekday() != time.Monday {
				err := rep.SaveWorkday(context.TODO(), storage.Workday{BarberID: barberID, Date: date, StartTime: "10:00", EndTime: "19:00"})
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
