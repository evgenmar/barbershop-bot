package telegram

import (
	cp "barbershop-bot/contextprovider"
	ent "barbershop-bot/entities"
	cfg "barbershop-bot/lib/config"
	tm "barbershop-bot/lib/time"
	sess "barbershop-bot/sessions"
	"time"

	tele "gopkg.in/telebot.v3"
)

func defineDisplayedDateRangeForAppointment(
	firstDisplayedDateRange ent.DateRange,
	displayedMonthRange ent.DateRange,
	deltaDisplayedMonth int8,
	appointment sess.Appointment,
) ent.DateRange {
	displayedDateRange := firstDisplayedDateRange
	if deltaDisplayedMonth == 0 {
		if appointment.WorkdayID != 0 {
			if appointment.LastShownDate.After(displayedDateRange.EndDate) {
				displayedDateRange = ent.MonthFromBase(appointment.LastShownDate, 0)
			}
		}
	} else {
		newDateRange := ent.MonthFromBase(appointment.LastShownDate, deltaDisplayedMonth)
		if newDateRange.EndDate.After(displayedMonthRange.EndDate) || newDateRange.EndDate.Before(displayedMonthRange.StartDate) {
			if appointment.LastShownDate.After(displayedDateRange.EndDate) {
				displayedDateRange = ent.MonthFromBase(appointment.LastShownDate, 0)
			}
		}
		if newDateRange.EndDate.After(displayedDateRange.EndDate) {
			displayedDateRange = newDateRange
		}
	}
	return displayedDateRange
}

func defineFirstDisplayedDateRangeForAppointment(appointment sess.Appointment) (firstDisplayedDateRange ent.DateRange, err error) {
	var relativeFirstDisplayedMonth byte = 0
	for relativeFirstDisplayedMonth <= cfg.MaxAppointmentBookingMonths {
		firstDisplayedDateRange = ent.MonthFromNow(relativeFirstDisplayedMonth)
		earlestFreeDate, ok, err := earlestDateWithFreeTime(appointment, firstDisplayedDateRange)
		if err != nil {
			return ent.DateRange{}, err
		}
		if ok {
			if earlestFreeDate.After(firstDisplayedDateRange.StartDate) {
				firstDisplayedDateRange.StartDate = earlestFreeDate
			}
			break
		}
		relativeFirstDisplayedMonth++
		firstDisplayedDateRange = ent.MonthFromNow(relativeFirstDisplayedMonth)
	}
	return
}

func earlestDateWithFreeTime(appointment sess.Appointment, dateRange ent.DateRange) (earlestFreeDate time.Time, ok bool, err error) {
	workdays, err := cp.RepoWithContext.GetWorkdaysByDateRange(appointment.BarberID, dateRange)
	if err != nil {
		return time.Time{}, false, err
	}
	appts, err := cp.RepoWithContext.GetAppointmentsByDateRange(appointment.BarberID, dateRange)
	if err != nil {
		return time.Time{}, false, err
	}
	appointments := make(map[int][]ent.Appointment)
	for _, appt := range appts {
		appointments[appt.WorkdayID] = append(appointments[appt.WorkdayID], appt)
	}
	for _, workday := range workdays {
		if haveFreeTimeForAppointment(workday, appointments[workday.ID], appointment.Duration) {
			return workday.Date, true, nil
		}
	}
	return time.Time{}, false, nil
}

func freeTimesForAppointment(workday ent.Workday, appointments []ent.Appointment, duration tm.Duration) (freeTimes []tm.Duration) {
	var analyzedTime tm.Duration
	if workday.Date.Equal(tm.Today()) {
		currentDayTime := tm.CurrentDayTime()
		if currentDayTime > workday.StartTime {
			analyzedTime = currentDayTime.RoundUpToMultipleOf30()
		} else {
			analyzedTime = workday.StartTime
		}
	} else {
		analyzedTime = workday.StartTime
	}
	i := 0
	for analyzedTime < workday.EndTime {
		if i < len(appointments) {
			if (appointments[i].Time - analyzedTime) >= duration {
				freeTimes = append(freeTimes, analyzedTime)
				analyzedTime += 30 * tm.Minute
			} else {
				analyzedTime = appointments[i].Time + appointments[i].Duration
				i++
			}
		} else {
			if (workday.EndTime - analyzedTime) >= duration {
				freeTimes = append(freeTimes, analyzedTime)
				analyzedTime += 30 * tm.Minute
			} else {
				break
			}
		}
	}
	return
}

func noAction(tele.Context) error { return nil }
