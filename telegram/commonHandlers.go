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

type monthRange struct {
	firstMonth tm.Month
	lastMonth  tm.Month
}

func calculateDisplayedRangesForAppointment(deltaDisplayedMonth int8, appointment sess.Appointment) (
	displayedDateRange ent.DateRange, displayedMonthRange monthRange, err error,
) {
	firstDisplayedDateRange, err := defineFirstDisplayedDateRangeForAppointment(appointment)
	if err != nil {
		return
	}
	displayedMonthRange, err = defineDisplayedMonthRangeForAppointment(firstDisplayedDateRange, appointment)
	if err != nil {
		return
	}
	displayedDateRange = defineDisplayedDateRange(
		firstDisplayedDateRange,
		displayedMonthRange,
		deltaDisplayedMonth,
		appointment.LastShownMonth,
	)
	return
}

func callbackUnique(endpnt string) string {
	return "\f" + endpnt
}

func checkAndRescheduleAppointment(appointment sess.Appointment) (ok bool, err error) {
	workday, appointments, err := getWorkdayAndAppointments(appointment.WorkdayID)
	if err != nil {
		return
	}
	ok = isTimeForAppointmentAvailable(workday, appointments, appointment)
	if !ok {
		return
	}
	err = cp.RepoWithContext.UpdateAppointment(ent.Appointment{
		ID:        appointment.ID,
		WorkdayID: appointment.WorkdayID,
		Time:      appointment.Time,
	})
	return
}

func defineDisplayedDateRange(
	firstDisplayedDateRange ent.DateRange,
	displayedMonthRange monthRange,
	deltaDisplayedMonth int8,
	lastShownMonth tm.Month,
) ent.DateRange {
	if displayedMonthRange.lastMonth == 0 {
		return ent.DateRange{}
	}
	if displayedMonthRange.firstMonth == displayedMonthRange.lastMonth {
		return firstDisplayedDateRange
	}
	newDisplayedMonth := lastShownMonth + tm.Month(deltaDisplayedMonth)
	if newDisplayedMonth > displayedMonthRange.lastMonth {
		return ent.Month(displayedMonthRange.lastMonth)
	}
	if newDisplayedMonth < displayedMonthRange.firstMonth {
		return firstDisplayedDateRange
	}
	if newDisplayedMonth > displayedMonthRange.firstMonth {
		return ent.Month(newDisplayedMonth)
	}
	return firstDisplayedDateRange
}

func defineDisplayedMonthRangeForAppointment(firstDisplayedDateRange ent.DateRange, appointment sess.Appointment) (monthRange, error) {
	firstMonth := tm.ParseMonth(firstDisplayedDateRange.LastDate)
	lastPossibleMonth := tm.MonthFromNow(cfg.MaxAppointmentBookingMonths)
	if firstMonth > lastPossibleMonth {
		return monthRange{}, nil
	}
	for lastMonth := lastPossibleMonth; lastMonth > firstMonth; lastMonth-- {
		ok, err := monthHaveFreeTimeForAppointment(lastMonth, appointment)
		if err != nil {
			return monthRange{}, err
		}
		if ok {
			return monthRange{firstMonth: firstMonth, lastMonth: lastMonth}, nil
		}
	}
	return monthRange{firstMonth: firstMonth, lastMonth: firstMonth}, nil
}

func defineFirstDisplayedDateRangeForAppointment(appointment sess.Appointment) (firstDisplayedDateRange ent.DateRange, err error) {
	firstDisplayedMonth := tm.ParseMonth(tm.Today())
	firstDisplayedDateRange = ent.Month(firstDisplayedMonth)
	for firstDisplayedMonth <= tm.MonthFromNow(cfg.MaxAppointmentBookingMonths) {
		earlestFreeDate, err := earlestDateWithFreeTime(appointment, firstDisplayedDateRange)
		if err != nil {
			return ent.DateRange{}, err
		}
		if !earlestFreeDate.Equal(time.Time{}) {
			if earlestFreeDate.After(firstDisplayedDateRange.FirstDate) {
				firstDisplayedDateRange.FirstDate = earlestFreeDate
			}
			break
		}
		firstDisplayedMonth++
		firstDisplayedDateRange = ent.Month(firstDisplayedMonth)
	}
	return
}

func earlestDateWithFreeTime(appointment sess.Appointment, dateRange ent.DateRange) (earlestFreeDate time.Time, err error) {
	workdays, err := cp.RepoWithContext.GetWorkdaysByDateRange(appointment.BarberID, dateRange)
	if err != nil {
		return time.Time{}, err
	}
	appts, err := cp.RepoWithContext.GetAppointmentsByDateRange(appointment.BarberID, dateRange)
	if err != nil {
		return time.Time{}, err
	}
	appointments := make(map[int][]ent.Appointment)
	for _, appt := range appts {
		appointments[appt.WorkdayID] = append(appointments[appt.WorkdayID], appt)
	}
	for _, workday := range workdays {
		if workdayHaveFreeTimeForAppointment(workday, appointments[workday.ID], appointment) {
			return workday.Date, nil
		}
	}
	return time.Time{}, nil
}

func excludePastAppointments(appointments []ent.Appointment, currentDayTime tm.Duration) []ent.Appointment {
	for i, appointment := range appointments {
		if appointment.Time+appointment.Duration > currentDayTime {
			return appointments[i:]
		}
	}
	return nil
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

func getWorkdayAndAppointments(workdayID int) (ent.Workday, []ent.Appointment, error) {
	workday, err := cp.RepoWithContext.GetWorkdayByID(workdayID)
	if err != nil {
		return ent.Workday{}, nil, err
	}
	appointments, err := cp.RepoWithContext.GetAppointmentsByDateRange(
		workday.BarberID,
		ent.DateRange{FirstDate: workday.Date, LastDate: workday.Date},
	)
	if err != nil {
		return ent.Workday{}, nil, err
	}
	return workday, appointments, nil
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

func monthHaveFreeTimeForAppointment(month tm.Month, appointment sess.Appointment) (bool, error) {
	dateRange := ent.Month(month)
	workdays, err := cp.RepoWithContext.GetWorkdaysByDateRange(appointment.BarberID, dateRange)
	if err != nil {
		return false, err
	}
	appts, err := cp.RepoWithContext.GetAppointmentsByDateRange(appointment.BarberID, dateRange)
	if err != nil {
		return false, err
	}
	appointments := make(map[int][]ent.Appointment)
	for _, appt := range appts {
		appointments[appt.WorkdayID] = append(appointments[appt.WorkdayID], appt)
	}
	for _, workday := range workdays {
		if workdayHaveFreeTimeForAppointment(workday, appointments[workday.ID], appointment) {
			return true, nil
		}
	}
	return false, nil
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
			appointments = excludePastAppointments(appointments, currentDayTime)
		}
	}
	return analyzedTime, appointments
}

func workdayHaveFreeTimeForAppointment(workday ent.Workday, appointments []ent.Appointment, appt sess.Appointment) bool {
	analyzedTime, appointments := prepareAnalizedTimeAndAppointments(workday, appointments, appt.ID)
	for _, appointment := range appointments {
		if (appointment.Time - analyzedTime) >= appt.Duration {
			return true
		}
		analyzedTime = appointment.Time + appointment.Duration
	}
	return (workday.EndTime - analyzedTime) >= appt.Duration
}
