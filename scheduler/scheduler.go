package scheduler

import (
	ent "barbershop-bot/entities"
	cfg "barbershop-bot/lib/config"
	"barbershop-bot/lib/e"
	tm "barbershop-bot/lib/time"
	rep "barbershop-bot/repository"
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

var Cron *cron.Cron

func InitCron() {
	Cron = CronWithSettings()
}

// CronWithSettings creates *cron.Cron and set it with several functions to be run on a schedule.
// Triggered events:
//   - making schedules for barbers - every Monday at 03:00 AM
func CronWithSettings() *cron.Cron {
	crn := cron.New(cron.WithLocation(cfg.Location))
	crn.AddFunc("0 3 * * 1",
		func() {
			if err := MakeSchedules(); err != nil {
				log.Print(err)
			}
		})
	return crn
}

// MakeSchedules just calls makeSchedule for all barbers specified in barberIDs.
// See makeSchedule for details.
func MakeSchedules() (err error) {
	barberIDs := cfg.Barbers.IDs()
	for _, barberID := range barberIDs {
		if err := MakeSchedule(barberID); err != nil {
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
func MakeSchedule(barberID int64) (err error) {
	defer func() { err = e.WrapIfErr("can't make schedule", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.TimoutRead)
	latestWorkDate, err := rep.Rep.GetLatestWorkDate(ctx, barberID)
	cancel()
	if err != nil {
		return err
	}
	ctx, cancel = context.WithTimeout(context.Background(), cfg.TimoutRead)
	barber, err := rep.Rep.GetBarberByID(ctx, barberID)
	cancel()
	if err != nil {
		return err
	}
	if latestWorkDate.After(barber.LastWorkdate) {
		dateRangeToDelete := ent.DateRange{StartDate: barber.LastWorkdate.Add(24 * time.Hour), EndDate: latestWorkDate}
		ctx, cancel := context.WithTimeout(context.Background(), cfg.TimoutWrite)
		defer cancel()
		err := rep.Rep.DeleteWorkdaysByDateRange(ctx, barber.ID, dateRangeToDelete)
		if err != nil {
			return err
		}
		return nil
	}
	workdays := calculateSchedule(latestWorkDate, barber)
	if len(workdays) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.TimoutWrite)
		defer cancel()
		if err = rep.Rep.CreateWorkdays(ctx, workdays...); err != nil {
			return err
		}
	}
	return nil
}

func calculateSchedule(latestWorkDate time.Time, barber ent.Barber) (workdays []ent.Workday) {
	dayDuration := 24 * time.Hour
	today := tm.Today()
	if latestWorkDate.Before(today) {
		latestWorkDate = today.Add(-dayDuration)
	}
	lastSheduledDate := today.Add(time.Duration(cfg.ScheduledWeeks) * dayDuration * 7)
	if lastSheduledDate.After(barber.LastWorkdate) {
		lastSheduledDate = barber.LastWorkdate
	}
	for date := lastSheduledDate; date.After(latestWorkDate); date = date.Add(-dayDuration) {
		if date.Weekday() != cfg.NonWorkingDay {
			workdays = append(workdays, ent.Workday{
				BarberID:  barber.ID,
				Date:      date,
				StartTime: ent.DefaultStart,
				EndTime:   ent.DefaultEnd,
			})
		}
	}
	return workdays
}
