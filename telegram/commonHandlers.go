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

func calculateAndCheckDisplayedRanges(deltaDisplayedMonth int8, appointment sess.Appointment) (
	displayedDateRange, displayedMonthRange ent.DateRange, ok bool, err error,
) {
	firstDisplayedDateRange, err := defineFirstDisplayedDateRangeForAppointment(appointment)
	if err != nil {
		return
	}
	if firstDisplayedDateRange.EndDate.After(ent.MonthFromNow(cfg.MaxAppointmentBookingMonths).EndDate) {
		return
	}
	ok = true
	displayedMonthRange = ent.DateRange{
		StartDate: firstDisplayedDateRange.EndDate,
		EndDate:   ent.MonthFromNow(cfg.MaxAppointmentBookingMonths).EndDate,
	}
	displayedDateRange = defineDisplayedDateRangeForAppointment(
		firstDisplayedDateRange,
		displayedMonthRange,
		deltaDisplayedMonth,
		appointment,
	)
	return
}

func callbackUnique(endpnt string) string {
	return "\f" + endpnt
}

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
		if haveFreeTimeForAppointment(workday, appointments[workday.ID], appointment) {
			return workday.Date, true, nil
		}
	}
	return time.Time{}, false, nil
}

func freeTimesForAppointment(workday ent.Workday, appointments []ent.Appointment, appt sess.Appointment) (freeTimes []tm.Duration) {
	analyzedTime, appointments := prepareAnalizedTimeAndAppointments(workday, appointments, appt.ID)
	analyzedApptIndex := 0
	for analyzedTime < workday.EndTime {
		if analyzedApptIndex < len(appointments) {
			if (appointments[analyzedApptIndex].Time - analyzedTime) >= appt.Duration {
				freeTimes = append(freeTimes, analyzedTime)
				analyzedTime += 30 * tm.Minute
			} else {
				analyzedTime = appointments[analyzedApptIndex].Time + appointments[analyzedApptIndex].Duration
				analyzedApptIndex++
			}
		} else {
			if (workday.EndTime - analyzedTime) >= appt.Duration {
				freeTimes = append(freeTimes, analyzedTime)
				analyzedTime += 30 * tm.Minute
			} else {
				break
			}
		}
	}
	return
}

func getFutureAppointments(appointments []ent.Appointment, currentDayTime tm.Duration) []ent.Appointment {
	for i, appointment := range appointments {
		if appointment.Time > currentDayTime {
			return appointments[i:]
		}
	}
	return nil
}

func getNullServiceInfo(serviceID int, appointmentDuration tm.Duration) string {
	service, err := cp.RepoWithContext.GetServiceByID(serviceID)
	if err != nil {
		return "Длительность услуги: " + appointmentDuration.LongString()
	}
	return service.Info()
}

func getWorkdayAndAppointments(workdayID int) (ent.Workday, []ent.Appointment, error) {
	workday, err := cp.RepoWithContext.GetWorkdayByID(workdayID)
	if err != nil {
		return ent.Workday{}, nil, err
	}
	appointments, err := cp.RepoWithContext.GetAppointmentsByDateRange(
		workday.BarberID,
		ent.DateRange{StartDate: workday.Date, EndDate: workday.Date},
	)
	if err != nil {
		return ent.Workday{}, nil, err
	}
	return workday, appointments, nil
}

func haveFreeTimeForAppointment(workday ent.Workday, appointments []ent.Appointment, appt sess.Appointment) bool {
	analyzedTime, appointments := prepareAnalizedTimeAndAppointments(workday, appointments, appt.ID)

	if workday.Date.Equal(tm.Today()) {
		currentDayTime := tm.CurrentDayTime()
		if currentDayTime > workday.StartTime {
			analyzedTime = currentDayTime
			appointments = getFutureAppointments(appointments, currentDayTime)
		}
	}
	for _, appointment := range appointments {
		if (appointment.Time - analyzedTime) >= appt.Duration {
			return true
		}
		analyzedTime = appointment.Time + appointment.Duration
	}
	return (workday.EndTime - analyzedTime) >= appt.Duration
}

func isTimeForAppointmentAvailable(workday ent.Workday, appointments []ent.Appointment, appt sess.Appointment) bool {
	if workday.Date.Equal(tm.Today()) && appt.Time < tm.CurrentDayTime() {
		return false
	}
	freeTimes := freeTimesForAppointment(workday, appointments, appt)
	for _, freeTime := range freeTimes {
		if appt.Time == freeTime {
			return true
		}
	}
	return false
}

func noAction(tele.Context) error { return nil }

func prepareAnalizedTimeAndAppointments(workday ent.Workday, appointments []ent.Appointment, oldAppointmentID int) (tm.Duration, []ent.Appointment) {
	if oldAppointmentID != 0 {
		for i, apappointment := range appointments {
			if apappointment.ID == oldAppointmentID {
				appointments = append(appointments[:i], appointments[i+1:]...)
				break
			}
		}
	}
	analyzedTime := workday.StartTime
	if workday.Date.Equal(tm.Today()) {
		currentDayTime := tm.CurrentDayTime()
		if currentDayTime > workday.StartTime {
			analyzedTime = currentDayTime.RoundUpToMultipleOf30()
			appointments = getFutureAppointments(appointments, currentDayTime)
		}
	}
	return analyzedTime, appointments
}
