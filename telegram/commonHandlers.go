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
	displayedMonthRange = monthRange{
		firstMonth: tm.ParseMonth(firstDisplayedDateRange.LastDate),
		lastMonth:  tm.MonthFromNow(cfg.MaxAppointmentBookingMonths),
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
	displayedMonthRange monthRange,
	deltaDisplayedMonth int8,
	appointment sess.Appointment,
) ent.DateRange {
	if displayedMonthRange.firstMonth > displayedMonthRange.lastMonth {
		return ent.DateRange{}
	}
	if displayedMonthRange.firstMonth == displayedMonthRange.lastMonth {
		return firstDisplayedDateRange
	}
	if deltaDisplayedMonth == 0 {
		if appointment.WorkdayID != 0 {
			if appointment.LastShownMonth > tm.ParseMonth(firstDisplayedDateRange.LastDate) {
				return ent.Month(appointment.LastShownMonth)
			}
		}
		return firstDisplayedDateRange
	}
	newDisplayedMonth := appointment.LastShownMonth + tm.Month(deltaDisplayedMonth)
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

func defineFirstDisplayedDateRangeForAppointment(appointment sess.Appointment) (firstDisplayedDateRange ent.DateRange, err error) {
	var relativeFirstDisplayedMonth byte = 0
	firstDisplayedDateRange = ent.MonthFromNow(relativeFirstDisplayedMonth)
	for relativeFirstDisplayedMonth <= cfg.MaxAppointmentBookingMonths {
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
		relativeFirstDisplayedMonth++
		firstDisplayedDateRange = ent.MonthFromNow(relativeFirstDisplayedMonth)
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
		if haveFreeTimeForAppointment(workday, appointments[workday.ID], appointment) {
			return workday.Date, nil
		}
	}
	return time.Time{}, nil
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
		ent.DateRange{FirstDate: workday.Date, LastDate: workday.Date},
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
