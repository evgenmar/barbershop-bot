package telegram

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	cp "github.com/evgenmar/barbershop-bot/contextprovider"
	ent "github.com/evgenmar/barbershop-bot/entities"
	cfg "github.com/evgenmar/barbershop-bot/lib/config"
	"github.com/evgenmar/barbershop-bot/lib/e"
	tm "github.com/evgenmar/barbershop-bot/lib/time"
	rep "github.com/evgenmar/barbershop-bot/repository"
	m "github.com/evgenmar/barbershop-bot/repository/mappers"
	sched "github.com/evgenmar/barbershop-bot/scheduler"
	sess "github.com/evgenmar/barbershop-bot/sessions"

	tele "gopkg.in/telebot.v3"
)

func onStartBarber(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	mutex.Lock()
	defer mutex.Unlock()
	if err := clearOldMenuForBarber(ctx.Sender().ID); err != nil {
		return logAndMsgErrBarberWithoutMenu(ctx, "can't show main menu", err)
	}
	return sendToBarberMenuAndUpdStoredMessage(ctx, mainMenu, markupBarberMain)
}

func onSignUpClientForAppointment(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	sess.UpdateAppointmentAndBarberState(
		barberID,
		sess.Appointment{BarberID: barberID},
		sess.StateStart,
	)
	services, err := cp.RepoWithContext.GetServicesByBarberID(barberID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't show to barber services for new appointment", err)
	}
	if len(services) == 0 {
		return ctx.Edit(createServiceFirst, markupBarberBackToMain)
	}
	return ctx.Edit(
		barberSelectServiceForAppointment,
		markupSelectService(services, endpntBarberSelectServiceForAppointment, endpntBarberBackToMain),
	)
}

func onBarberSelectServiceForAppointment(ctx tele.Context) error {
	errMsg := "can't show free workdays for new appointment"
	serviceID, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	service, err := cp.RepoWithContext.GetServiceByID(serviceID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	newAppointment := sess.GetAppointmentBarber(ctx.Sender().ID)
	newAppointment.ServiceID = service.ID
	newAppointment.Duration = service.Duration
	return calculateAndShowToBarberFreeWorkdaysForAppointment(ctx, 0, newAppointment)
}

func onBarberSelectMonthForAppointment(ctx tele.Context) error {
	delta, err := strconv.ParseInt(ctx.Callback().Data, 10, 8)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't show free workdays for appointment", err)
	}
	appointment := sess.GetAppointmentBarber(ctx.Sender().ID)
	return calculateAndShowToBarberFreeWorkdaysForAppointment(ctx, int8(delta), appointment)
}

func onBarberSelectWorkdayForAppointment(ctx tele.Context) error {
	errMsg := "can't show free time for appointment"
	workdayID, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	workday, appointments, err := getWorkdayAndAppointments(workdayID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	appointment := sess.GetAppointmentBarber(ctx.Sender().ID)
	appointment.WorkdayID = workdayID
	return calculateAndShowToBarberFreeTimesForAppointment(ctx, workday, appointments, appointment)
}

func onBarberSelectTimeForAppointment(ctx tele.Context) error {
	errMsg := "can't handle select time for new appointment"
	apptTime, err := strconv.ParseUint(ctx.Callback().Data, 10, 64)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	appointmentTime := tm.Duration(apptTime)
	barberID := ctx.Sender().ID
	appointment := sess.GetAppointmentBarber(barberID)
	appointment.Time = appointmentTime
	workday, appointments, err := getWorkdayAndAppointments(appointment.WorkdayID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	if !isTimeForAppointmentAvailable(workday, appointments, appointment) {
		return calculateAndShowToBarberFreeTimesForAppointment(ctx, workday, appointments, appointment)
	}
	sess.UpdateAppointmentAndBarberState(barberID, appointment, sess.StateStart)
	if appointment.ID == 0 {
		service, err := cp.RepoWithContext.GetServiceByID(appointment.ServiceID)
		if err != nil {
			return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
		}
		return ctx.Edit(
			fmt.Sprintf(confirmNewAppointment, service.ShortInfo(), tm.ShowDate(workday.Date), appointmentTime.ShortString()),
			markupBarberConfirmNewAppointment,
		)
	}
	serviceInfo := shortNullServiceInfo(appointment.ServiceID, appointment.Duration)
	return ctx.Edit(
		fmt.Sprintf(confirmRescheduleAppointment, serviceInfo, tm.ShowDate(workday.Date), appointmentTime.ShortString()),
		markupBarberConfirmRescheduleAppointment,
	)
}

func onBarberConfirmNewAppointment(ctx tele.Context) error {
	errMsg := "can't confirm new appointment for barber"
	barberID := ctx.Sender().ID
	appointment := sess.GetAppointmentBarber(barberID)
	ok, err := checkAndCreateAppointmentByBarber(appointment)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	if ok {
		serviceInfo := shortNullServiceInfo(appointment.ServiceID, appointment.Duration)
		workday, err := cp.RepoWithContext.GetWorkdayByID(appointment.WorkdayID)
		if err != nil {
			return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
		}
		return ctx.Edit(
			fmt.Sprintf(newAppointmentSavedByBarber, serviceInfo, tm.ShowDate(workday.Date), appointment.Time.ShortString()),
			markupUpdNote,
		)
	}
	return ctx.Edit(failedToSaveAppointment, markupBarberFailedToSaveOrRescheduleAppointment)
}

func onBarberSelectAnotherTimeForAppointment(ctx tele.Context) error {
	appointment := sess.GetAppointmentBarber(ctx.Sender().ID)
	workday, appointments, err := getWorkdayAndAppointments(appointment.WorkdayID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't handle select another time for appointment", err)
	}
	return calculateAndShowToBarberFreeTimesForAppointment(ctx, workday, appointments, appointment)
}

func onAddNote(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateAddNote)
	return ctx.Edit(enterNote)
}

func onAddWorkday(ctx tele.Context) error {
	delta, err := strconv.ParseInt(ctx.Callback().Data, 10, 8)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't show add workday calendar", err)
	}
	return calculateAndShowChangeDayTypeCalendar(ctx, false, noDaysOff, selectNonWorkingDay, delta)
}

func onAddNonWorkday(ctx tele.Context) error {
	delta, err := strconv.ParseInt(ctx.Callback().Data, 10, 8)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't show delete workday calendar", err)
	}
	return calculateAndShowChangeDayTypeCalendar(ctx, true, noWorkdays, selectWorkingDay, delta)
}

func onCreateWorkday(ctx tele.Context) error {
	errMsg := "can't create workday"
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	newWorkdayDate, err := time.ParseInLocation(time.DateOnly, ctx.Callback().Data, cfg.Location)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	if err := cp.RepoWithContext.CreateWorkdays(ent.Workday{
		BarberID:  ctx.Sender().ID,
		Date:      newWorkdayDate,
		StartTime: ent.DefaultStart,
		EndTime:   ent.DefaultEnd,
	}); err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	return calculateAndShowChangeDayTypeCalendar(ctx, false, noDaysOff, selectNonWorkingDay, 0)
}

func onDeleteWorkday(ctx tele.Context) error {
	errMsg := "can't delete workday"
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	workdayID, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	mutex.Lock()
	if err := cp.RepoWithContext.DeleteWorkdayByID(workdayID); err != nil {
		if errors.Is(err, rep.ErrAppointmentsExists) {
			return ctx.Edit(failedToDeleteWorkday, markupBarberBackToMain)
		}
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	mutex.Unlock()
	return calculateAndShowChangeDayTypeCalendar(ctx, true, noWorkdays, selectWorkingDay, 0)
}

func onMyWorkSchedule(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	appointment := sess.GetAppointmentBarber(barberID)
	if appointment.BarberID == 0 {
		appointment = sess.Appointment{
			BarberID: ctx.Sender().ID,
		}
	}
	return calculateAndShowScheduleCalendar(ctx, 0, appointment)
}

func onSelectMonthFromScheduleCalendar(ctx tele.Context) error {
	delta, err := strconv.ParseInt(ctx.Callback().Data, 10, 8)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't show schedule calendar", err)
	}
	appointment := sess.GetAppointmentBarber(ctx.Sender().ID)
	return calculateAndShowScheduleCalendar(ctx, int8(delta), appointment)
}

func onSelectWorkdayFromScheduleCalendar(ctx tele.Context) error {
	workdayID, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't show schedule workday menu", err)
	}
	barberID := ctx.Sender().ID
	appointment := sess.GetAppointmentBarber(barberID)
	appointment.WorkdayID = workdayID
	sess.UpdateAppointmentAndBarberState(barberID, appointment, sess.StateStart)
	return showScheduledWorkdayMenu(ctx, appointment)
}

func onMakeThisDayNonWorking(ctx tele.Context) error {
	appointment := sess.GetAppointmentBarber(ctx.Sender().ID)
	ok, err := checkAndDeleteWorkday(appointment.WorkdayID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't add certain non-working day from schedule calendar", err)
	}
	if ok {
		return calculateAndShowScheduleCalendar(ctx, 0, appointment)
	}
	return showScheduledWorkdayMenu(ctx, appointment)
}

func onUpdWorkdayStartTime(ctx tele.Context) error {
	appointment := sess.GetAppointmentBarber(ctx.Sender().ID)
	latestStart, err := latestPossibleStart(appointment.WorkdayID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't define latest possible start time for workday", err)
	}
	return ctx.Edit(
		fmt.Sprintf(selectWorkdayStartTime, ent.EarlestStart.ShortString()),
		markupSelectWorkdayStartOrEndTime(ent.EarlestStart, latestStart, endpntSelectWorkdayStartTime),
	)
}

func onSelectWorkdayStartTime(ctx tele.Context) error {
	errMsg := "can't update workday start time"
	start, err := strconv.ParseUint(ctx.Callback().Data, 10, 64)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	startTime := tm.Duration(start)
	appointment := sess.GetAppointmentBarber(ctx.Sender().ID)
	ok, err := checkAndUpdateWorkdayStartTime(startTime, appointment.WorkdayID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	if ok {
		return showScheduledWorkdayMenu(ctx, appointment)
	}
	return ctx.Edit(failToUpdateWorkdayStartTime, markupBackToWorkdayInfo(appointment.WorkdayID))
}

func onUpdWorkdayEndTime(ctx tele.Context) error {
	appointment := sess.GetAppointmentBarber(ctx.Sender().ID)
	earlestEnd, err := earlestPossibleEnd(appointment.WorkdayID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't define earlest possible end time for workday", err)
	}
	return ctx.Edit(
		fmt.Sprintf(selectWorkdayEndTime, ent.LatestEnd.ShortString()),
		markupSelectWorkdayStartOrEndTime(earlestEnd, ent.LatestEnd, endpntSelectWorkdayEndTime),
	)
}

func onSelectWorkdayEndTime(ctx tele.Context) error {
	errMsg := "can't update workday end time"
	end, err := strconv.ParseUint(ctx.Callback().Data, 10, 64)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	endTime := tm.Duration(end)
	appointment := sess.GetAppointmentBarber(ctx.Sender().ID)
	ok, err := checkAndUpdateWorkdayEndTime(endTime, appointment.WorkdayID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	if ok {
		return showScheduledWorkdayMenu(ctx, appointment)
	}
	return ctx.Edit(failToUpdateWorkdayEndTime, markupBackToWorkdayInfo(appointment.WorkdayID))
}

func onSelectAppointment(ctx tele.Context) error {
	errMsg := "can't show appointment options menu"
	splitData := strings.Split(ctx.Callback().Data, "|")
	if len(splitData) != 2 {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, errors.New("invalid appointment data"))
	}
	apptID, err := strconv.Atoi(splitData[0])
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	barberID := ctx.Sender().ID
	appointment := sess.GetAppointmentBarber(barberID)
	appointment.ID = apptID
	appointment.HashStr = splitData[1]
	editedAppointment, err := getVeryfiedAppointment(appointment)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	if editedAppointment.ID == 0 {
		return showScheduledWorkdayMenu(ctx, appointment)
	}
	sess.UpdateAppointmentAndBarberState(barberID, appointment, sess.StateStart)
	return showAppointmentOptionsMenu(ctx, editedAppointment)
}

func onBarberRescheduleAppointment(ctx tele.Context) error {
	appointment := sess.GetAppointmentBarber(ctx.Sender().ID)
	editedAppointment, err := getVeryfiedAppointment(appointment)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't show to barber free workdays for reschedule appointment", err)
	}
	if editedAppointment.ID == 0 {
		return showScheduledWorkdayMenu(ctx, appointment)
	}
	appointment.ServiceID = editedAppointment.ServiceID
	appointment.Duration = editedAppointment.Duration
	return calculateAndShowToBarberFreeWorkdaysForAppointment(ctx, 0, appointment)
}

func onBarberConfirmRescheduleAppointment(ctx tele.Context) error {
	errMsg := "can't confirm reschedule appointment for barber"
	barberID := ctx.Sender().ID
	appointment := sess.GetAppointmentBarber(barberID)
	mutex.Lock()
	defer mutex.Unlock()
	editedAppointment, err := getVeryfiedAppointment(appointment)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	if editedAppointment.ID == 0 {
		return ctx.Edit(appointmentNotRescheduled+reasons, markupBarberBackToMain)
	}
	ok, err := checkAndRescheduleAppointment(appointment)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	if ok {
		editedAppointment.WorkdayID = appointment.WorkdayID
		editedAppointment.Time = appointment.Time
		appointment.HashStr = appointmentHashStr(editedAppointment)
		sess.UpdateAppointmentAndBarberState(barberID, appointment, sess.StateStart)
		if err := sendToUserRescheduleOrCancelNotification(editedAppointment, appointmentRescheduledByBarber); err != nil {
			return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
		}
		return showAppointmentOptionsMenu(ctx, editedAppointment)
	}
	return ctx.Edit(failedToRescheduleAppointment, markupBarberFailedToSaveOrRescheduleAppointment)
}

func onBarberCancelAppointment(ctx tele.Context) error {
	return barberCancelAppointment(ctx, markupBarberConfirmCancelAppointment)
}

func onBarberConfirmCancelAppointment(ctx tele.Context) error {
	errMsg := "can't confirm cancel appointment for barber"
	barberID := ctx.Sender().ID
	appointment := sess.GetAppointmentBarber(barberID)
	mutex.Lock()
	defer mutex.Unlock()
	editedAppointment, err := getVeryfiedAppointment(appointment)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	if editedAppointment.ID == 0 {
		return ctx.Edit(appointmentNotCanceled+reasons, markupBarberBackToMain)
	}
	if err := cp.RepoWithContext.DeleteAppointmentByID(editedAppointment.ID); err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	if err := sendToUserRescheduleOrCancelNotification(editedAppointment, appointmentCanceledByBarber); err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	return showScheduledWorkdayMenu(ctx, appointment)
}

func onCancelAppointmentAndApology(ctx tele.Context) error {
	return barberCancelAppointment(ctx, markupConfirmCancelAppointmentAndApology)
}

func onConfirmCancelAppointmentAndApology(ctx tele.Context) error {
	errMsg := "can't confirm cancel appointment for barber"
	barberID := ctx.Sender().ID
	appointment := sess.GetAppointmentBarber(barberID)
	mutex.Lock()
	defer mutex.Unlock()
	editedAppointment, err := getVeryfiedAppointment(appointment)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	if editedAppointment.ID == 0 {
		return ctx.Edit(appointmentNotCanceled+reasons, markupBarberBackToMain)
	}
	if err := cp.RepoWithContext.DeleteAppointmentByID(editedAppointment.ID); err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	if err := sendToUserCancelNotificationAndApology(editedAppointment); err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	return showScheduledWorkdayMenu(ctx, appointment)
}

func onUpdNote(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateUpdNote)
	return ctx.Edit(enterNote)
}

func onBarberSettings(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(settingsMenu, markupBarberSettings)
}

func onBarbersMemo(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(barbersMemo, markupShortBarberSettings)
}

func onBarberManageAccount(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(manageBarberAccount, markupBarberManageAccount)
}

func onBarberCurrentSettings(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	barber, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't show current barber settings", err)
	}
	return ctx.Edit(currentSettings+barber.Info(), markupBarberBackToMain)
}

func onBarberUpdPersonal(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	barberID := ctx.Sender().ID
	barber, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't show update personal data options for barber", err)
	}
	if barber.Name == ent.NoName && barber.Phone == ent.NoPhone {
		return ctx.Edit(privacyBarber, markupBarberPrivacy)
	}
	return ctx.Edit(selectPersonalData, markupBarberPersonal)
}

func onBarberAgreeWithPrivacy(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(selectPersonalData, markupBarberPersonal)
}

func onBarberUpdName(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateUpdName)
	return ctx.Edit(updName)
}

func onBarberUpdPhone(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateUpdPhone)
	return ctx.Edit(updPhone)
}

func onDeleteAccount(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(deleteAccount, markupDeleteAccount)
}

func onSetLastWorkDate(ctx tele.Context) error {
	errMsg := "can't open select last work date menu"
	barberID := ctx.Sender().ID
	latestAppointmentDate, err := cp.RepoWithContext.GetLatestAppointmentDate(barberID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	delta, err := strconv.ParseInt(ctx.Callback().Data, 10, 8)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	firstDisplayedDateRange := defineFirstDisplayedDateRangeForLastWorkDate(latestAppointmentDate)
	displayedMonthRange := monthRange{
		firstMonth: tm.ParseMonth(firstDisplayedDateRange.LastDate),
		lastMonth:  tm.MonthFromNow(cfg.ScheduledWeeks * 7 / 30),
	}
	lastWorkDate := sess.GetLastWorkDate(barberID)
	displayedDateRange := defineDisplayedDateRange(
		firstDisplayedDateRange,
		displayedMonthRange,
		int8(delta),
		lastWorkDate.LastShownMonth,
	)
	lastWorkDate.LastShownMonth = tm.ParseMonth(displayedDateRange.LastDate)
	sess.UpdateLastWorkDateAndState(barberID, lastWorkDate, sess.StateStart)
	return ctx.Edit(selectLastWorkDate, markupSelectLastWorkDate(displayedDateRange, displayedMonthRange))
}

func onSelectLastWorkDate(ctx tele.Context) error {
	errMsg := "can't save last work date"
	barberID := ctx.Sender().ID
	sess.UpdateLastWorkDateAndState(barberID, sess.LastWorkDate{}, sess.StateStart)
	dateToSave, err := time.ParseInLocation(time.DateOnly, ctx.Callback().Data, cfg.Location)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	barber, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	switch dateToSave.Compare(barber.LastWorkdate) {
	case 0:
		return ctx.Edit(lastWorkDateUnchanged, markupBarberBackToMain)
	case 1:
		if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: barberID, LastWorkdate: dateToSave}); err != nil {
			return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
		}
		if err := sched.MakeSchedule(barberID); err != nil {
			log.Print(e.Wrap(errMsg, err))
			// TODO: ensure atomicity using outbox pattern
			return ctx.Edit(lastWorkDateSavedWithoutSсhedule, markupBarberBackToMain)
		}
		return ctx.Edit(lastWorkDateSaved, markupBarberBackToMain)
	case -1:
		latestWorkDate, err := cp.RepoWithContext.GetLatestWorkDate(barberID)
		if err != nil {
			return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
		}
		dateRangeToDelete := ent.DateRange{FirstDate: dateToSave.Add(24 * time.Hour), LastDate: latestWorkDate}
		err = cp.RepoWithContext.DeleteWorkdaysByDateRange(barberID, dateRangeToDelete)
		if err != nil && !errors.Is(err, rep.ErrInvalidDateRange) {
			if errors.Is(err, rep.ErrAppointmentsExists) {
				return ctx.Edit(haveAppointmentAfterDataToSave, markupBarberBackToMain)
			}
			return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
		}
		if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: barberID, LastWorkdate: dateToSave}); err != nil {
			// TODO: ensure atomicity using outbox pattern
			return ctx.Edit(lastWorkDateNotSavedButScheduleDeleted, markupBarberBackToMain)
		}
		return ctx.Edit(lastWorkDateSaved, markupBarberBackToMain)
	default:
		return nil
	}
}

func onSelfDeleteBarber(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	barberToDelete, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't provide options for barber self deletion", err)
	}
	if barberToDelete.LastWorkdate.Before(tm.Today()) {
		return ctx.Edit(confirmSelfDeletion, markupConfirmSelfDeletion)
	}
	return ctx.Edit(youHaveActiveSchedule+preDeletionBarberInstruction, markupBarberBackToMain)
}

func onSureToSelfDeleteBarber(ctx tele.Context) error {
	errMsg := "can't self delete barber"
	barberIDToDelete := ctx.Sender().ID
	if err := cp.RepoWithContext.DeletePastAppointments(barberIDToDelete); err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	if err := cp.RepoWithContext.DeleteBarberByID(barberIDToDelete); err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	cfg.Barbers.RemoveID(barberIDToDelete)
	return ctx.Edit(goodbuyBarber)
}

func onManageServices(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	services, err := cp.RepoWithContext.GetServicesByBarberID(barberID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't show manage services menu", err)
	}
	if len(services) == 0 {
		return ctx.Edit(youHaveNoServices, markupManageServicesShort)
	}
	return ctx.Edit(manageServices, markupManageServicesFull)
}

func onShowMyServices(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	services, err := cp.RepoWithContext.GetServicesByBarberID(barberID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't show list of services", err)
	}
	if len(services) == 0 {
		return ctx.Edit(youHaveNoServices, markupManageServicesShort)
	}
	servicesInfo := ""
	for _, service := range services {
		servicesInfo = servicesInfo + "\n\n" + service.Info()
	}
	return ctx.Edit(yourServices+servicesInfo, markupShowMyServices)
}

func onCreateService(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	newService := sess.GetNewService(barberID)
	if newService.Name != "" || newService.Desciption != "" || newService.Price != 0 || newService.Duration != 0 {
		return ctx.Edit(continueOldOrMakeNewService, markupСontinueOldOrMakeNewService)
	}
	return showNewServOptsWithEditMsg(ctx, newService)
}

func onСontinueOldService(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	newService := sess.GetNewService(barberID)
	return showNewServOptsWithEditMsg(ctx, newService)
}

func onMakeNewService(ctx tele.Context) error {
	sess.UpdateNewServiceAndState(ctx.Sender().ID, sess.NewService{}, sess.StateStart)
	return showNewServOptsWithEditMsg(ctx, sess.NewService{})
}

func onEnterServiceName(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateEnterServiceName)
	return ctx.Edit(enterServiceName)
}

func onEnterServiceDescription(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateEnterServiceDescription)
	return ctx.Edit(enterServiceDescription)
}

func onEnterServicePrice(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateEnterServicePrice)
	return ctx.Edit(enterServicePrice)
}

func onSelectServiceDurationOnEnter(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(selectServiceDuration, markupSelectServiceDuration(endpntEnterServiceDuration))
}

func onSelectCertainDurationOnEnter(ctx tele.Context) error {
	dur, err := strconv.ParseUint(ctx.Callback().Data, 10, 64)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't select certain service duration", err)
	}
	barberID := ctx.Sender().ID
	newService := sess.GetNewService(barberID)
	newService.Duration = tm.Duration(dur)
	sess.UpdateNewServiceAndState(barberID, newService, sess.StateStart)
	return showNewServOptsWithEditMsg(ctx, newService)
}

func onSaveNewService(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	newService := sess.GetNewService(barberID)
	err := cp.RepoWithContext.CreateService(ent.Service{
		BarberID:   barberID,
		Name:       newService.Name,
		Desciption: newService.Desciption,
		Price:      newService.Price,
		Duration:   newService.Duration,
	})
	if err != nil {
		sess.UpdateBarberState(barberID, sess.StateStart)
		if errors.Is(err, rep.ErrInvalidService) {
			return ctx.Edit(invalidService, markupBarberBackToMain)
		}
		if errors.Is(err, rep.ErrAlreadyExists) {
			return ctx.Edit(nonUniqueServiceName+"\n\n"+newService.Info(), markupEnterServiceName, tele.ModeMarkdown)
		}
		return logAndMsgErrBarberWithEdit(ctx, "can't create service", err)
	}
	sess.UpdateNewServiceAndState(barberID, sess.NewService{}, sess.StateStart)
	return ctx.Edit(serviceCreated, markupManageServicesFull)
}

func onEditService(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	editedService := sess.GetEditedService(barberID)
	if editedService.ID != 0 {
		return ctx.Edit(continueEditingOrSelectService, markupContinueEditingOrSelectService)
	}
	return onSelectServiceToEdit(ctx)
}

func onСontinueEditingService(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	editedService := sess.GetEditedService(barberID)
	return showEditServOptsWithEditMsg(ctx, editedService)
}

func onSelectServiceToEdit(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	services, err := cp.RepoWithContext.GetServicesByBarberID(barberID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't show list of services for editing", err)
	}
	if len(services) == 0 {
		return ctx.Edit(youHaveNoServices, markupManageServicesShort)
	}
	return ctx.Edit(
		selectServiceToEdit,
		markupSelectService(services, endpntServiceToEdit, endpntBarberBackToMain),
	)
}

func onSelectCertainServiceToEdit(ctx tele.Context) error {
	errMsg := "can't select certain service for editing"
	serviceID, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	serviceToEdit, err := cp.RepoWithContext.GetServiceByID(serviceID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	editedService := sess.EditedService{
		ID: serviceID,
		OldService: sess.Service{
			Name:       serviceToEdit.Name,
			Desciption: serviceToEdit.Desciption,
			Price:      serviceToEdit.Price,
			Duration:   serviceToEdit.Duration,
		},
	}
	sess.UpdateEditedServiceAndState(ctx.Sender().ID, editedService, sess.StateStart)
	return showEditServOptsWithEditMsg(ctx, editedService)
}

func onEditServiceName(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateEditServiceName)
	return ctx.Edit(enterServiceName)
}

func onEditServiceDescription(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateEditServiceDescription)
	return ctx.Edit(enterServiceDescription)
}

func onEditServicePrice(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateEditServicePrice)
	return ctx.Edit(enterServicePrice)
}

func onSelectServiceDurationOnEdit(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(selectServiceDuration, markupSelectServiceDuration(endpntEditServiceDuration))
}

func onSelectCertainDurationOnEdit(ctx tele.Context) error {
	dur, err := strconv.ParseUint(ctx.Callback().Data, 10, 64)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't select certain service duration", err)
	}
	barberID := ctx.Sender().ID
	editedService := sess.GetEditedService(barberID)
	editedService.UpdService.Duration = tm.Duration(dur)
	sess.UpdateEditedServiceAndState(barberID, editedService, sess.StateStart)
	return showEditServOptsWithEditMsg(ctx, editedService)
}

func onUpdateService(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	editedService := sess.GetEditedService(barberID)
	err := cp.RepoWithContext.UpdateService(ent.Service{
		ID:         editedService.ID,
		Name:       editedService.UpdService.Name,
		Desciption: editedService.UpdService.Desciption,
		Price:      editedService.UpdService.Price,
		Duration:   editedService.UpdService.Duration,
	})
	if err != nil {
		sess.UpdateBarberState(barberID, sess.StateStart)
		if errors.Is(err, rep.ErrInvalidService) {
			return ctx.Edit(invalidService, markupBarberBackToMain)
		}
		if errors.Is(err, rep.ErrNonUniqueData) {
			return ctx.Edit(nonUniqueServiceName+"\n\n"+editedService.Info(), markupEditServiceName)
		}
		return logAndMsgErrBarberWithEdit(ctx, "can't update service", err)
	}
	sess.UpdateEditedServiceAndState(barberID, sess.EditedService{}, sess.StateStart)
	return ctx.Edit(serviceUpdated, markupManageServicesFull)
}

func onDeleteService(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	services, err := cp.RepoWithContext.GetServicesByBarberID(barberID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't show list of services for delete", err)
	}
	if len(services) == 0 {
		return ctx.Edit(youHaveNoServices, markupManageServicesShort)
	}
	return ctx.Edit(
		selectServiceToDelete,
		markupSelectService(services, endpntServiceToDelete, endpntBarberBackToMain),
	)
}

func onSelectCertainServiceToDelete(ctx tele.Context) error {
	errMsg := "can't select certain service for delete"
	serviceID, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	serviceToDelete, err := cp.RepoWithContext.GetServiceByID(serviceID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	return ctx.Edit(fmt.Sprintf(confirmServiceDeletion, serviceToDelete.Info()), markupConfirmServiceDeletion(serviceID))
}

func onSureToDeleteService(ctx tele.Context) error {
	errMsg := "can't delete service"
	serviceID, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	if err := cp.RepoWithContext.DeleteServiceByID(serviceID); err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	return ctx.Edit(serviceDeleted, markupBarberBackToMain)
}

func onManageBarbers(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(manageBarbers, markupManageBarbers)
}

func onShowAllBarbers(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	barbers, err := cp.RepoWithContext.GetAllBarbers()
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't show all barbers", err)
	}
	barbersInfo := ""
	for _, barber := range barbers {
		barbersInfo = barbersInfo + "\n\n" + barber.PublicInfo()
	}
	return ctx.Edit(listOfBarbers+barbersInfo, markupBarberBackToMain, tele.ModeMarkdown)
}

func onAddBarber(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateAddBarber)
	return ctx.Edit(addBarber)
}

func onDeleteBarber(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	barbers, err := cp.RepoWithContext.GetAllBarbers()
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't suggest actions to delete barber", err)
	}
	if len(barbers) == 1 {
		return ctx.Edit(onlyOneBarberExists, markupBarberBackToMain)
	}
	markupSelectBarber := markupSelectBarberToDeletion(barberID, barbers)
	if len(markupSelectBarber.InlineKeyboard) == 1 {
		return ctx.Edit(noBarbersToDelete+preDeletionBarberInstruction, markupSelectBarber)
	}
	return ctx.Edit(selectBarberToDeletion+preDeletionBarberInstruction, markupSelectBarber)
}

func onDeleteCertainBarber(ctx tele.Context) error {
	errMsg := "can't delete barber"
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	barberIDToDelete, err := strconv.ParseInt(ctx.Callback().Data, 10, 64)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	barberToDelete, err := cp.RepoWithContext.GetBarberByID(barberIDToDelete)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	if barberToDelete.LastWorkdate.Before(tm.Today()) {
		if err := cp.RepoWithContext.DeletePastAppointments(barberIDToDelete); err != nil {
			return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
		}
		if err := cp.RepoWithContext.DeleteBarberByID(barberIDToDelete); err != nil {
			return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
		}
		cfg.Barbers.RemoveID(barberIDToDelete)
		return ctx.Edit(barberDeleted, markupBarberBackToMain)
	}
	return ctx.Edit(barberHaveActiveSchedule, markupBarberBackToMain)
}

func onBarberBackToMain(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(mainMenu, markupBarberMain)
}

func onTextBarber(ctx tele.Context) error {
	mutex.Lock()
	defer mutex.Unlock()
	state := sess.GetBarberState(ctx.Sender().ID)
	switch state {
	case sess.StateUpdName:
		return onUpdateBarberName(ctx)
	case sess.StateUpdPhone:
		return onUpdateBarberPhone(ctx)
	case sess.StateEnterServiceName:
		return onEnterServName(ctx)
	case sess.StateEnterServiceDescription:
		return onEnterServDescription(ctx)
	case sess.StateEnterServicePrice:
		return onEnterServPrice(ctx)
	case sess.StateEditServiceName:
		return onEditServName(ctx)
	case sess.StateEditServiceDescription:
		return onEditServDescription(ctx)
	case sess.StateEditServicePrice:
		return onEditServPrice(ctx)
	case sess.StateAddNote:
		return onCreateNote(ctx)
	case sess.StateUpdNote:
		return onUpdateNote(ctx)
	default:
		return ctx.Send(unknownCommand)
	}
}

func onUpdateBarberName(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: barberID, Name: ctx.Message().Text}); err != nil {
		if errors.Is(err, rep.ErrInvalidBarber) {
			log.Print(e.Wrap("invalid barber name", err))
			return ctx.Send(invalidName)
		}
		if errors.Is(err, rep.ErrNonUniqueData) {
			log.Print(e.Wrap("barber's name must be unique", err))
			return ctx.Send(notUniqueBarberName)
		}
		sess.UpdateBarberState(barberID, sess.StateStart)
		return logAndMsgErrBarberWithSend(ctx, "can't update barber's name", err)
	}
	sess.UpdateBarberState(barberID, sess.StateStart)
	return sendToBarberMenuAndUpdStoredMessage(ctx, updNameSuccess, markupBarberPersonal)
}

func onUpdateBarberPhone(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: barberID, Phone: ctx.Message().Text}); err != nil {
		if errors.Is(err, rep.ErrInvalidBarber) {
			log.Print(e.Wrap("invalid barber phone", err))
			return ctx.Send(invalidPhone)
		}
		if errors.Is(err, rep.ErrNonUniqueData) {
			log.Print(e.Wrap("barber's phone must be unique", err))
			return ctx.Send(notUniqueBarberPhone)
		}
		sess.UpdateBarberState(barberID, sess.StateStart)
		return logAndMsgErrBarberWithSend(ctx, "can't update barber's phone", err)
	}
	sess.UpdateBarberState(barberID, sess.StateStart)
	return sendToBarberMenuAndUpdStoredMessage(ctx, updPhoneSuccess, markupBarberPersonal)
}

func onEnterServName(ctx tele.Context) error {
	text := ctx.Message().Text
	if !m.IsValidServiceName(text) {
		return ctx.Send(invalidServiceName)
	}
	barberID := ctx.Sender().ID
	newService := sess.GetNewService(barberID)
	newService.Name = text
	sess.UpdateNewServiceAndState(barberID, newService, sess.StateStart)
	return showNewServOptsWithSendMsg(ctx, newService)
}

func onEnterServDescription(ctx tele.Context) error {
	text := ctx.Message().Text
	if !m.IsValidDescription(text) {
		return ctx.Send(invalidServiceDescription)
	}
	barberID := ctx.Sender().ID
	newService := sess.GetNewService(barberID)
	newService.Desciption = text
	sess.UpdateNewServiceAndState(barberID, newService, sess.StateStart)
	return showNewServOptsWithSendMsg(ctx, newService)
}

func onEnterServPrice(ctx tele.Context) error {
	text := ctx.Message().Text
	price, err := ent.NewPrice(text)
	if err != nil {
		log.Print(e.Wrap("invalid service price", err))
		return ctx.Send(invalidServicePrice)
	}
	barberID := ctx.Sender().ID
	newService := sess.GetNewService(barberID)
	newService.Price = price
	sess.UpdateNewServiceAndState(barberID, newService, sess.StateStart)
	return showNewServOptsWithSendMsg(ctx, newService)
}

func onEditServName(ctx tele.Context) error {
	text := ctx.Message().Text
	if !m.IsValidServiceName(text) {
		return ctx.Send(invalidServiceName)
	}
	barberID := ctx.Sender().ID
	editedService := sess.GetEditedService(barberID)
	editedService.UpdService.Name = text
	sess.UpdateEditedServiceAndState(barberID, editedService, sess.StateStart)
	return showEditServOptsWithSendMsg(ctx, editedService)
}

func onEditServDescription(ctx tele.Context) error {
	text := ctx.Message().Text
	if !m.IsValidDescription(text) {
		return ctx.Send(invalidServiceDescription)
	}
	barberID := ctx.Sender().ID
	editedService := sess.GetEditedService(barberID)
	editedService.UpdService.Desciption = text
	sess.UpdateEditedServiceAndState(barberID, editedService, sess.StateStart)
	return showEditServOptsWithSendMsg(ctx, editedService)
}

func onEditServPrice(ctx tele.Context) error {
	text := ctx.Message().Text
	price, err := ent.NewPrice(text)
	if err != nil {
		log.Print(e.Wrap("invalid service price", err))
		return ctx.Send(invalidServicePrice)
	}
	barberID := ctx.Sender().ID
	editedService := sess.GetEditedService(barberID)
	editedService.UpdService.Price = price
	sess.UpdateEditedServiceAndState(barberID, editedService, sess.StateStart)
	return showEditServOptsWithSendMsg(ctx, editedService)
}

func onCreateNote(ctx tele.Context) error {
	errMsg := "can't create note"
	barberID := ctx.Sender().ID
	appointment := sess.GetAppointmentBarber(barberID)
	appointmentID, err := cp.RepoWithContext.GetAppointmentIDByWorkdayIDAndTime(appointment.WorkdayID, appointment.Time)
	if err != nil {
		return logAndMsgErrBarberWithSend(ctx, errMsg, err)
	}
	if err := cp.RepoWithContext.UpdateAppointment(ent.Appointment{ID: appointmentID, Note: ctx.Message().Text}); err != nil {
		if errors.Is(err, rep.ErrInvalidAppointment) {
			log.Print(e.Wrap("invalid note", err))
			return ctx.Send(invalidNote)
		}
		sess.UpdateBarberState(barberID, sess.StateStart)
		return logAndMsgErrBarberWithSend(ctx, errMsg, err)
	}
	sess.UpdateBarberState(barberID, sess.StateStart)
	return sendToBarberMenuAndUpdStoredMessage(ctx, updNoteSuccess, markupBarberBackToMain)
}

func onUpdateNote(ctx tele.Context) error {
	errMsg := "can't update note"
	barberID := ctx.Sender().ID
	appointment := sess.GetAppointmentBarber(barberID)
	if err := cp.RepoWithContext.UpdateAppointment(ent.Appointment{ID: appointment.ID, Note: ctx.Message().Text}); err != nil {
		if errors.Is(err, rep.ErrInvalidAppointment) {
			log.Print(e.Wrap("invalid note", err))
			return ctx.Send(invalidNote)
		}
		sess.UpdateBarberState(barberID, sess.StateStart)
		return logAndMsgErrBarberWithSend(ctx, errMsg, err)
	}
	sess.UpdateBarberState(barberID, sess.StateStart)
	editedAppointment, err := getVeryfiedAppointment(appointment)
	if err != nil {
		return logAndMsgErrBarberWithSend(ctx, errMsg, err)
	}
	if editedAppointment.ID == 0 {
		_, err := cp.RepoWithContext.GetAppointmentByID(appointment.ID)
		if err != nil {
			if errors.Is(err, rep.ErrNoSavedAppointment) {
				return sendToBarberMenuAndUpdStoredMessage(ctx, appointmentDeletedByUser, markupBarberBackToMain)
			}
			return logAndMsgErrBarberWithSend(ctx, errMsg, err)
		}
		return sendToBarberMenuAndUpdStoredMessage(ctx, updNoteSuccessAndAppointmentRescheduled, markupBarberBackToMain)
	}
	return sendAppointmentOptionsMenu(ctx, editedAppointment)
}

func onContactBarber(ctx tele.Context) error {
	mutex.Lock()
	defer mutex.Unlock()
	state := sess.GetBarberState(ctx.Sender().ID)
	switch state {
	case sess.StateAddBarber:
		return onAddNewBarber(ctx)
	default:
		return ctx.Send(unknownCommand)
	}
}

func onAddNewBarber(ctx tele.Context) error {
	errMsg := "can't add new barber"
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	newBarberID := ctx.Message().Contact.UserID
	if err := cp.RepoWithContext.CreateBarber(newBarberID); err != nil {
		if errors.Is(err, rep.ErrAlreadyExists) {
			return sendToBarberMenuAndUpdStoredMessage(ctx, userIsAlreadyBarber, markupBarberBackToMain)
		}
		return logAndMsgErrBarberWithSend(ctx, errMsg, err)
	}
	cfg.Barbers.AddID(newBarberID)
	if err := sched.MakeSchedule(newBarberID); err != nil {
		log.Print(e.Wrap(errMsg, err))
		// TODO: ensure atomicity using outbox pattern
		return sendToBarberMenuAndUpdStoredMessage(ctx, addedNewBarberWithoutSсhedule, markupBarberBackToMain)
	}
	return sendToBarberMenuAndUpdStoredMessage(ctx, addedNewBarberWithSсhedule, markupBarberBackToMain)
}

func barberCancelAppointment(ctx tele.Context, markup *tele.ReplyMarkup) error {
	errMsg := "can't handle barber cancel appointment"
	appointment := sess.GetAppointmentBarber(ctx.Sender().ID)
	editedAppointment, err := getVeryfiedAppointment(appointment)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	if editedAppointment.ID == 0 {
		return showScheduledWorkdayMenu(ctx, appointment)
	}
	serviceInfo := shortNullServiceInfo(editedAppointment.ServiceID, editedAppointment.Duration)
	workday, err := cp.RepoWithContext.GetWorkdayByID(editedAppointment.WorkdayID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	return ctx.Edit(
		fmt.Sprintf(barberConfirmCancelAppointment, serviceInfo, tm.ShowDate(workday.Date), editedAppointment.Time.ShortString()),
		markup,
	)
}

func calculateAndShowChangeDayTypeCalendar(ctx tele.Context, showWorkdays bool, noDaysMsg, selectDayMsg string, delta int64) error {
	errMsg := "can't show add or delete workday calendar"
	barberID := ctx.Sender().ID
	barber, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	firstDisplayedDateRange, err := defineFirstDisplayedDateRangeForDayTypeChange(barber, showWorkdays)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	displayedMonthRange := defineDisplayedMonthRangeForDayTypeChange(firstDisplayedDateRange, barber.LastWorkdate)
	if displayedMonthRange.lastMonth < displayedMonthRange.firstMonth {
		return ctx.Edit(noDaysMsg, markupBackToMyWorkSchedule)
	}
	appointment := sess.GetAppointmentBarber(barberID)
	displayedDateRange := defineDisplayedDateRange(
		firstDisplayedDateRange,
		displayedMonthRange,
		int8(delta),
		appointment.LastShownMonth,
	)
	if displayedDateRange.LastDate.After(barber.LastWorkdate) {
		displayedDateRange.LastDate = barber.LastWorkdate
	}
	appointment.LastShownMonth = tm.ParseMonth(displayedDateRange.LastDate)
	sess.UpdateAppointmentAndBarberState(barberID, appointment, sess.StateStart)
	markupSelectDay, err := markupSelectDayForDayTypeChange(displayedDateRange, displayedMonthRange, barberID, showWorkdays)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	return ctx.Edit(selectDayMsg, markupSelectDay)
}

func calculateAndShowScheduleCalendar(ctx tele.Context, deltaDisplayedMonth int8, appointment sess.Appointment) error {
	errMsg := "can't show schedule calendar"
	displayedDateRange, displayedMonthRange, err := calculateDisplayedRangesForScheduleCalendar(deltaDisplayedMonth, appointment)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	if displayedMonthRange.firstMonth > displayedMonthRange.lastMonth {
		return ctx.Edit(scheduleCalendarIsEmpty, markupBarberBackToMain)
	}
	appointment.LastShownMonth = tm.ParseMonth(displayedDateRange.LastDate)
	sess.UpdateAppointmentAndBarberState(ctx.Sender().ID, appointment, sess.StateStart)
	markupSelectWorkday, err := markupSelectWorkdayFromScheduleCalendar(
		displayedDateRange,
		displayedMonthRange,
		appointment.BarberID,
	)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	return ctx.Edit(selectWorkday, markupSelectWorkday)
}

func calculateAndShowToBarberFreeTimesForAppointment(ctx tele.Context, workday ent.Workday, appointments []ent.Appointment, appointment sess.Appointment) error {
	freeTimes := freeTimesForAppointment(workday, appointments, appointment)
	if len(freeTimes) == 0 {
		return calculateAndShowToBarberFreeWorkdaysForAppointment(ctx, 0, appointment)
	}
	sess.UpdateAppointmentAndBarberState(ctx.Sender().ID, appointment, sess.StateStart)
	serviceInfo := shortNullServiceInfo(appointment.ServiceID, appointment.Duration)
	return ctx.Edit(
		fmt.Sprintf(selectTimeForAppointment, serviceInfo, tm.ShowDate(workday.Date)),
		markupSelectTimeForAppointment(
			freeTimes,
			endpntBarberSelectTimeForAppointment,
			endpntBarberSelectMonthForAppointment,
			endpntBarberBackToMain,
		),
	)
}

func calculateAndShowToBarberFreeWorkdaysForAppointment(ctx tele.Context, deltaDisplayedMonth int8, appointment sess.Appointment) error {
	displayedDateRange, displayedMonthRange, err := calculateDisplayedRangesForAppointment(deltaDisplayedMonth, appointment)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't show to barber free workdays for new appointment", err)
	}
	if displayedMonthRange.lastMonth == 0 {
		return ctx.Edit(informBarberNoFreeTimeForAppointment, markupBarberBackToMain)
	}
	appointment.LastShownMonth = tm.ParseMonth(displayedDateRange.LastDate)
	sess.UpdateAppointmentAndBarberState(ctx.Sender().ID, appointment, sess.StateStart)
	return showToBarberFreeWorkdaysForAppointment(ctx, displayedDateRange, displayedMonthRange, appointment)
}

func calculateDisplayedRangesForScheduleCalendar(deltaDisplayedMonth int8, appointment sess.Appointment) (
	displayedDateRange ent.DateRange, displayedMonthRange monthRange, err error,
) {
	firstDisplayedDateRange, err := defineFirstDisplayedDateRangeForScheduleCalendar(appointment.BarberID)
	if err != nil {
		return
	}
	displayedMonthRange = monthRange{
		firstMonth: tm.ParseMonth(firstDisplayedDateRange.LastDate),
		lastMonth:  tm.MonthFromNow(cfg.ScheduledWeeks * 7 / 30),
	}
	displayedDateRange = defineDisplayedDateRange(
		firstDisplayedDateRange,
		displayedMonthRange,
		deltaDisplayedMonth,
		appointment.LastShownMonth,
	)
	return
}

func checkAndCreateAppointmentByBarber(appointment sess.Appointment) (ok bool, err error) {
	mutex.Lock()
	defer mutex.Unlock()
	workday, appointments, err := getWorkdayAndAppointments(appointment.WorkdayID)
	if err != nil {
		return
	}
	ok = isTimeForAppointmentAvailable(workday, appointments, appointment)
	if !ok {
		return
	}
	err = cp.RepoWithContext.CreateAppointment(ent.Appointment{
		WorkdayID: appointment.WorkdayID,
		ServiceID: appointment.ServiceID,
		Time:      appointment.Time,
		Duration:  appointment.Duration,
		CreatedAt: time.Now().Unix(),
	})
	return
}

func checkAndDeleteWorkday(workdayID int) (ok bool, err error) {
	mutex.Lock()
	defer mutex.Unlock()
	workday, appointments, err := getWorkdayAndAppointments(workdayID)
	if err != nil {
		return
	}
	ok = len(appointments) == 0
	if !ok {
		return
	}
	err = cp.RepoWithContext.DeleteWorkdaysByDateRange(
		workday.BarberID,
		ent.DateRange{FirstDate: workday.Date, LastDate: workday.Date},
	)
	return
}

func checkAndUpdateWorkdayEndTime(endTime tm.Duration, workdayID int) (ok bool, err error) {
	mutex.Lock()
	defer mutex.Unlock()
	earlestEnd, err := earlestPossibleEnd(workdayID)
	if err != nil {
		return
	}
	ok = endTime >= earlestEnd
	if !ok {
		return
	}
	err = cp.RepoWithContext.UpdateWorkday(ent.Workday{ID: workdayID, EndTime: endTime})
	return
}

func checkAndUpdateWorkdayStartTime(startTime tm.Duration, workdayID int) (ok bool, err error) {
	mutex.Lock()
	defer mutex.Unlock()
	latestStart, err := latestPossibleStart(workdayID)
	if err != nil {
		return
	}
	ok = startTime <= latestStart
	if !ok {
		return
	}
	err = cp.RepoWithContext.UpdateWorkday(ent.Workday{ID: workdayID, StartTime: startTime})
	return
}

func clearOldMenuForBarber(barberID int64) error {
	barber, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return err
	}
	if barber.ChatID != 0 {
		if _, err := Bot.Edit(barber.StoredMessage, textReplacingMenu); err != nil {
			return err
		}
		Bot.Delete(barber.StoredMessage)
	}
	return nil
}

func defineDisplayedMonthRangeForDayTypeChange(firstDisplayedDateRange ent.DateRange, lastWorkdate time.Time) monthRange {
	if tm.MonthFromNow(cfg.ScheduledWeeks*7/30) <= tm.ParseMonth(lastWorkdate) {
		return monthRange{
			firstMonth: tm.ParseMonth(firstDisplayedDateRange.LastDate),
			lastMonth:  tm.MonthFromNow(cfg.ScheduledWeeks * 7 / 30),
		}
	}
	return monthRange{
		firstMonth: tm.ParseMonth(firstDisplayedDateRange.LastDate),
		lastMonth:  tm.ParseMonth(lastWorkdate),
	}
}

func defineFirstDisplayedDateRangeForDayTypeChange(barber ent.Barber, showWorkdays bool) (firstDisplayedDateRange ent.DateRange, err error) {
	firstDisplayedMonth := tm.ParseMonth(tm.Today())
	lastDisplayedMonth := tm.MonthFromNow(cfg.ScheduledWeeks * 7 / 30)
	lastWorkdateMonth := tm.ParseMonth(barber.LastWorkdate)
	if lastWorkdateMonth < lastDisplayedMonth {
		lastDisplayedMonth = lastWorkdateMonth
	}
	firstDisplayedDateRange = ent.Month(firstDisplayedMonth)
	if firstDisplayedDateRange.LastDate.After(barber.LastWorkdate) {
		firstDisplayedDateRange.LastDate = barber.LastWorkdate
	}
	for firstDisplayedMonth <= lastDisplayedMonth {
		earlestDisplayedDate, err := earlestDisplayedDateForDayTypeChange(barber.ID, showWorkdays, firstDisplayedDateRange)
		if err != nil {
			return ent.DateRange{}, err
		}
		if !earlestDisplayedDate.Equal(time.Time{}) {
			if earlestDisplayedDate.After(firstDisplayedDateRange.FirstDate) {
				firstDisplayedDateRange.FirstDate = earlestDisplayedDate
			}
			break
		}
		firstDisplayedMonth++
		firstDisplayedDateRange = ent.Month(firstDisplayedMonth)
		if firstDisplayedDateRange.LastDate.After(barber.LastWorkdate) {
			firstDisplayedDateRange.LastDate = barber.LastWorkdate
		}
	}
	return
}

func defineFirstDisplayedDateRangeForLastWorkDate(latestAppointmentDate time.Time) (firstDisplayedDateRange ent.DateRange) {
	firstDisplayedMonth := tm.ParseMonth(tm.Today())
	for firstDisplayedMonth <= tm.MonthFromNow(cfg.MaxAppointmentBookingMonths) {
		firstDisplayedDateRange = ent.Month(firstDisplayedMonth)
		if latestAppointmentDate.Compare(firstDisplayedDateRange.LastDate) <= 0 {
			if latestAppointmentDate.After(firstDisplayedDateRange.FirstDate) {
				firstDisplayedDateRange.FirstDate = latestAppointmentDate
			}
			break
		}
		firstDisplayedMonth++
	}
	return
}

func defineFirstDisplayedDateRangeForScheduleCalendar(barberID int64) (firstDisplayedDateRange ent.DateRange, err error) {
	firstDisplayedMonth := tm.ParseMonth(tm.Today())
	firstDisplayedDateRange = ent.Month(firstDisplayedMonth)
	for firstDisplayedMonth <= tm.MonthFromNow(cfg.ScheduledWeeks*7/30) {
		earlestWorkDate, err := earlestScheduledDate(barberID, firstDisplayedDateRange)
		if err != nil {
			return ent.DateRange{}, err
		}
		if !earlestWorkDate.Equal(time.Time{}) {
			if earlestWorkDate.After(firstDisplayedDateRange.FirstDate) {
				firstDisplayedDateRange.FirstDate = earlestWorkDate
			}
			break
		}
		firstDisplayedMonth++
		firstDisplayedDateRange = ent.Month(firstDisplayedMonth)
	}
	return
}

func earlestDisplayedDateForDayTypeChange(barberID int64, showWorkdays bool, dateRange ent.DateRange) (earlestDate time.Time, err error) {
	workdays, err := cp.RepoWithContext.GetWorkdaysByDateRange(barberID, dateRange)
	if err != nil {
		return time.Time{}, err
	}
	if showWorkdays {
		if len(workdays) == 0 {
			return time.Time{}, nil
		}
		appts, err := cp.RepoWithContext.GetAppointmentsByDateRange(barberID, dateRange)
		if err != nil {
			return time.Time{}, err
		}
		appointments := make(map[int][]ent.Appointment)
		for _, appt := range appts {
			appointments[appt.WorkdayID] = append(appointments[appt.WorkdayID], appt)
		}
		for _, workday := range workdays {
			if len(appointments[workday.ID]) == 0 {
				return workday.Date, nil
			}
		}
		return time.Time{}, nil
	}
	for i, date := 0, dateRange.FirstDate; date.Compare(dateRange.LastDate) <= 0; i, date = i+1, date.Add(24*time.Hour) {
		if i > len(workdays)-1 {
			return date, nil
		}
		if date.Before(workdays[i].Date) {
			return date, nil
		}
	}
	return time.Time{}, nil
}

func earlestPossibleEnd(workdayID int) (tm.Duration, error) {
	workday, appointments, err := getWorkdayAndAppointments(workdayID)
	if err != nil {
		return 0, err
	}
	if len(appointments) == 0 {
		return workday.StartTime + 30*tm.Minute, nil
	}
	return appointments[len(appointments)-1].Time + appointments[len(appointments)-1].Duration, nil
}

func earlestScheduledDate(barberID int64, dateRange ent.DateRange) (earlestWorkDate time.Time, err error) {
	workdays, err := cp.RepoWithContext.GetWorkdaysByDateRange(barberID, dateRange)
	if err != nil {
		return time.Time{}, err
	}
	if len(workdays) == 0 {
		return time.Time{}, nil
	}
	if workdays[0].Date.Equal(tm.Today()) && workdays[0].EndTime < tm.CurrentDayTime() {
		if len(workdays) > 1 {
			return workdays[1].Date, nil
		}
		return time.Time{}, nil
	}
	return workdays[0].Date, nil
}

func getVeryfiedAppointment(target sess.Appointment) (ent.Appointment, error) {
	appointment, err := cp.RepoWithContext.GetAppointmentByID(target.ID)
	if err != nil {
		if errors.Is(err, rep.ErrNoSavedAppointment) {
			return ent.Appointment{}, nil
		}
		return ent.Appointment{}, err
	}
	if appointmentHashStr(appointment) != target.HashStr {
		return ent.Appointment{}, nil
	}
	workday, err := cp.RepoWithContext.GetWorkdayByID(appointment.WorkdayID)
	if err != nil {
		return ent.Appointment{}, err
	}
	today := tm.Today()
	if workday.Date.Before(today) ||
		(workday.Date.Equal(today) && appointment.Time+appointment.Duration <= tm.CurrentDayTime()) {
		return ent.Appointment{}, nil
	}
	return appointment, nil
}

func latestPossibleStart(workdayID int) (tm.Duration, error) {
	workday, appointments, err := getWorkdayAndAppointments(workdayID)
	if err != nil {
		return 0, err
	}
	if len(appointments) == 0 {
		return workday.EndTime - 30*tm.Minute, nil
	}
	return appointments[0].Time, nil
}

func logAndMsgErrBarberWithEdit(ctx tele.Context, msg string, err error) error {
	log.Print(e.Wrap(msg, err))
	return ctx.Edit(errorBarber, markupBarberBackToMain)
}

func logAndMsgErrBarberWithSend(ctx tele.Context, msg string, err error) error {
	log.Print(e.Wrap(msg, err))
	return sendToBarberMenuAndUpdStoredMessage(ctx, errorBarber, markupBarberBackToMain)
}

func logAndMsgErrBarberWithoutMenu(ctx tele.Context, msg string, err error) error {
	log.Print(e.Wrap(msg, err))
	return ctx.Send(errorBarber)
}

func sendAppointmentOptionsMenu(ctx tele.Context, appointment ent.Appointment) error {
	errMsg := "can't send appointment options menu"
	serviceInfo := shortNullServiceInfo(appointment.ServiceID, appointment.Duration)
	workday, err := cp.RepoWithContext.GetWorkdayByID(appointment.WorkdayID)
	if err != nil {
		return logAndMsgErrBarberWithSend(ctx, errMsg, err)
	}
	userInfo, err := userInfo(appointment)
	if err != nil {
		return logAndMsgErrBarberWithSend(ctx, errMsg, err)
	}
	return sendToBarberMenuAndUpdStoredMessage(ctx,
		fmt.Sprintf(appointmentInfoForBarber, serviceInfo, tm.ShowDate(workday.Date), appointment.Time.ShortString(), userInfo),
		markupEditAppointment(appointment.WorkdayID, appointment.UserID),
		tele.ModeMarkdown,
	)
}

func sendToBarberMenuAndUpdStoredMessage(ctx tele.Context, what interface{}, opts ...interface{}) error {
	errMsg := "can't send menu to barber"
	message, err := Bot.Send(ctx.Recipient(), what, opts...)
	if err != nil {
		return logAndMsgErrBarberWithoutMenu(ctx, errMsg, err)
	}
	storedMessage := tele.StoredMessage{MessageID: strconv.Itoa(message.ID), ChatID: message.Chat.ID}
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{
		ID:            ctx.Sender().ID,
		StoredMessage: storedMessage,
	}); err != nil {
		if err := Bot.Delete(storedMessage); err != nil {
			return logAndMsgErrBarberWithoutMenu(ctx, errMsg, err)
		}
		return logAndMsgErrBarberWithoutMenu(ctx, errMsg, err)
	}
	return nil
}

func sendToUserCancelNotificationAndApology(appointment ent.Appointment) error {
	if appointment.UserID != 0 {
		user, err := cp.RepoWithContext.GetUserByID(appointment.UserID)
		if err != nil {
			return err
		}
		serviceInfo := shortNullServiceInfo(appointment.ServiceID, appointment.Duration)
		workday, err := cp.RepoWithContext.GetWorkdayByID(appointment.WorkdayID)
		if err != nil {
			return err
		}
		barber, err := cp.RepoWithContext.GetBarberByID(workday.BarberID)
		if err != nil {
			return err
		}
		if _, err := Bot.Send(user,
			fmt.Sprintf(appointmentCanceledByBarberWithApology,
				tm.ShowDate(workday.Date),
				appointment.Time.ShortString(),
				barber.Name,
				barber.Contacts(),
				serviceInfo,
			),
			tele.ModeMarkdown,
		); err != nil {
			return err
		}
	}
	return nil
}

func sendToUserRescheduleOrCancelNotification(appointment ent.Appointment, text string) error {
	if appointment.UserID != 0 {
		user, err := cp.RepoWithContext.GetUserByID(appointment.UserID)
		if err != nil {
			return err
		}
		serviceInfo := shortNullServiceInfo(appointment.ServiceID, appointment.Duration)
		workday, err := cp.RepoWithContext.GetWorkdayByID(appointment.WorkdayID)
		if err != nil {
			return err
		}
		barber, err := cp.RepoWithContext.GetBarberByID(workday.BarberID)
		if err != nil {
			return err
		}
		if _, err := Bot.Send(user,
			fmt.Sprintf(text, serviceInfo, barber.Name, tm.ShowDate(workday.Date), appointment.Time.ShortString()),
		); err != nil {
			return err
		}
	}
	return nil
}

func shortNullServiceInfo(serviceID int, appointmentDuration tm.Duration) string {
	service, err := cp.RepoWithContext.GetServiceByID(serviceID)
	if err != nil {
		return "Длительность услуги: " + appointmentDuration.LongString()
	}
	return service.ShortInfo()
}

func showAppointmentOptionsMenu(ctx tele.Context, appointment ent.Appointment) error {
	errMsg := "can't show appointment options menu"
	serviceInfo := shortNullServiceInfo(appointment.ServiceID, appointment.Duration)
	workday, err := cp.RepoWithContext.GetWorkdayByID(appointment.WorkdayID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	userInfo, err := userInfo(appointment)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, errMsg, err)
	}
	return ctx.Edit(
		fmt.Sprintf(appointmentInfoForBarber, serviceInfo, tm.ShowDate(workday.Date), appointment.Time.ShortString(), userInfo),
		markupEditAppointment(appointment.WorkdayID, appointment.UserID),
		tele.ModeMarkdown,
	)
}

func showEditServOptsWithEditMsg(ctx tele.Context, editedService sess.EditedService) error {
	if editedService.UpdService.Name != "" ||
		editedService.UpdService.Desciption != "" ||
		editedService.UpdService.Price != 0 ||
		editedService.UpdService.Duration != 0 {
		return ctx.Edit(editServiceParams+editedService.Info()+readyToUpdateService, markupReadyToUpdateService)
	}
	return ctx.Edit(editServiceParams+editedService.Info(), markupEditServiceParams)
}

func showEditServOptsWithSendMsg(ctx tele.Context, editedService sess.EditedService) error {
	if editedService.UpdService.Name != "" ||
		editedService.UpdService.Desciption != "" ||
		editedService.UpdService.Price != 0 ||
		editedService.UpdService.Duration != 0 {
		return sendToBarberMenuAndUpdStoredMessage(ctx, editServiceParams+editedService.Info()+readyToUpdateService, markupReadyToUpdateService)
	}
	return sendToBarberMenuAndUpdStoredMessage(ctx, editServiceParams+editedService.Info(), markupEditServiceParams)
}

func showNewServOptsWithEditMsg(ctx tele.Context, newService sess.NewService) error {
	if newService.Name != "" && newService.Desciption != "" && newService.Price != 0 && newService.Duration != 0 {
		return ctx.Edit(enterServiceParams+newService.Info()+readyToCreateService, markupReadyToCreateService, tele.ModeMarkdown)
	}
	return ctx.Edit(enterServiceParams+newService.Info(), markupEnterServiceParams, tele.ModeMarkdown)
}

func showNewServOptsWithSendMsg(ctx tele.Context, newService sess.NewService) error {
	if newService.Name != "" && newService.Desciption != "" && newService.Price != 0 && newService.Duration != 0 {
		return sendToBarberMenuAndUpdStoredMessage(ctx,
			enterServiceParams+newService.Info()+readyToCreateService,
			markupReadyToCreateService,
			tele.ModeMarkdown,
		)
	}
	return sendToBarberMenuAndUpdStoredMessage(ctx,
		enterServiceParams+newService.Info(),
		markupEnterServiceParams,
		tele.ModeMarkdown,
	)
}

func showScheduledWorkdayMenu(ctx tele.Context, appointment sess.Appointment) error {
	workday, appointments, err := getWorkdayAndAppointments(appointment.WorkdayID)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't show schedule workday menu", err)
	}
	if len(appointments) == 0 {
		return ctx.Edit(
			fmt.Sprintf(workdayIsFree, tm.ShowDate(workday.Date), workday.StartTime.ShortString(), workday.EndTime.ShortString()),
			markupWorkdayIsFree,
		)
	}
	today := tm.Today()
	if workday.Date.Before(today) {
		return calculateAndShowScheduleCalendar(ctx, 0, appointment)
	}
	if workday.Date.Equal(today) {
		appointments = excludePastAppointments(appointments, tm.CurrentDayTime())
	}
	return ctx.Edit(
		fmt.Sprintf(selectAppointment, tm.ShowDate(workday.Date), workday.StartTime.ShortString(), workday.EndTime.ShortString()),
		markupSelectAppointment(appointments),
	)
}

func showToBarberFreeWorkdaysForAppointment(
	ctx tele.Context,
	displayedDateRange ent.DateRange,
	displayedMonthRange monthRange,
	appointment sess.Appointment,
) error {
	markupSelectWorkday, err := markupSelectWorkdayForAppointment(
		displayedDateRange,
		displayedMonthRange,
		appointment,
		endpntBarberSelectWorkdayForAppointment,
		endpntBarberSelectMonthForAppointment,
		endpntBarberBackToMain,
	)
	if err != nil {
		return logAndMsgErrBarberWithEdit(ctx, "can't show to barber free workdays for appointment", err)
	}
	serviceInfo := shortNullServiceInfo(appointment.ServiceID, appointment.Duration)
	return ctx.Edit(fmt.Sprintf(selectDateForAppointment, serviceInfo), markupSelectWorkday)
}

func userInfo(appointment ent.Appointment) (info string, err error) {
	userData := make([]string, 0, 4)
	if appointment.UserID != 0 {
		user, err := cp.RepoWithContext.GetUserByID(appointment.UserID)
		if err != nil {
			return "", err
		}
		if user.Name != ent.NoName {
			userData = append(userData, "Имя : "+user.Name)
		}
		if user.Phone != ent.NoPhone {
			userData = append(userData, "Телефон: "+user.Phone)
		}
		userData = append(userData, fmt.Sprintf("[Ссылка на профиль](tg://user?id=%d)", user.ID))
	}
	if appointment.Note != "" {
		userData = append(userData, "Заметка : "+appointment.Note)
	}
	if len(userData) == 0 {
		return "Нет информации о клиенте", nil
	}
	return "Информация о клиенте:\n" + strings.Join(userData, "\n"), nil
}
