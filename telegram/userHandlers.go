package telegram

import (
	cp "barbershop-bot/contextprovider"
	ent "barbershop-bot/entities"
	"barbershop-bot/lib/e"
	tm "barbershop-bot/lib/time"
	rep "barbershop-bot/repository"
	sess "barbershop-bot/sessions"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"
)

func onStartUser(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
	return ctx.Send(mainMenu, markupUserMain)
}

func onSignUpForAppointment(ctx tele.Context) error {
	errMsg := "can't handle sign up for appointment"
	userID := ctx.Sender().ID
	upcomingAppointment, err := cp.RepoWithContext.GetUpcomingAppointment(userID)
	if err != nil {
		if errors.Is(err, rep.ErrNoSavedAppointment) {
			return showBarbersForAppointment(ctx, userID)
		}
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	return informUserAppointmentAlreadyExists(ctx, upcomingAppointment)
}

func onSelectBarberForAppointment(ctx tele.Context) error {
	errMsg := "can't show to user services for appointment"
	barberID, err := strconv.ParseInt(ctx.Callback().Data, 10, 64)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	barber, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	return showToUserServicesForNewAppointment(ctx, barber)
}

func onUserSelectServiceForAppointment(ctx tele.Context) error {
	errMsg := "can't show free workdays for new appointment"
	serviceID, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	service, err := cp.RepoWithContext.GetServiceByID(serviceID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	newAppointment := sess.GetAppointmentUser(ctx.Sender().ID)
	newAppointment.ServiceID = service.ID
	newAppointment.Duration = service.Duration
	return calculateAndShowToUserFreeWorkdaysForAppointment(ctx, 0, newAppointment)
}

func onUserSelectMonthForAppointment(ctx tele.Context) error {
	delta, err := strconv.ParseInt(ctx.Callback().Data, 10, 8)
	if err != nil {
		return logAndMsgErrUser(ctx, "can't show free workdays for appointment", err)
	}
	appointment := sess.GetAppointmentUser(ctx.Sender().ID)
	return calculateAndShowToUserFreeWorkdaysForAppointment(ctx, int8(delta), appointment)
}

func onUserSelectWorkdayForAppointment(ctx tele.Context) error {
	errMsg := "can't show free time for appointment"
	workdayID, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	workday, appointments, err := getWorkdayAndAppointments(workdayID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	appointment := sess.GetAppointmentUser(ctx.Sender().ID)
	appointment.WorkdayID = workdayID
	return calculateAndShowToUserFreeTimesForAppointment(ctx, workday, appointments, appointment)
}

func onUserSelectTimeForAppointment(ctx tele.Context) error {
	errMsg := "can't handle select time for new appointment"
	apptTime, err := strconv.ParseUint(ctx.Callback().Data, 10, 64)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	appointmentTime := tm.Duration(apptTime)
	userID := ctx.Sender().ID
	appointment := sess.GetAppointmentUser(userID)
	appointment.Time = appointmentTime
	workday, appointments, err := getWorkdayAndAppointments(appointment.WorkdayID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	if !isTimeForAppointmentAvailable(workday, appointments, appointment) {
		return calculateAndShowToUserFreeTimesForAppointment(ctx, workday, appointments, appointment)
	}
	sess.UpdateAppointmentAndUserState(userID, appointment, sess.StateStart)
	if appointment.ID == 0 {
		service, err := cp.RepoWithContext.GetServiceByID(appointment.ServiceID)
		if err != nil {
			return logAndMsgErrUser(ctx, errMsg, err)
		}
		return ctx.Edit(
			fmt.Sprintf(confirmNewAppointment, service.Info(), tm.ShowDate(workday.Date), appointmentTime.ShortString()),
			markupUserConfirmNewAppointment,
		)
	}
	serviceInfo := getNullServiceInfo(appointment.ServiceID, appointment.Duration)
	return ctx.Edit(
		fmt.Sprintf(confirmRescheduleAppointment, serviceInfo, tm.ShowDate(workday.Date), appointmentTime.ShortString()),
		markupUserConfirmRescheduleAppointment,
	)
}

func onUserConfirmNewAppointment(ctx tele.Context) error {
	errMsg := "can't confirm new appointment for user"
	userID := ctx.Sender().ID
	upcomingAppointment, err := cp.RepoWithContext.GetUpcomingAppointment(userID)
	if err != nil {
		if errors.Is(err, rep.ErrNoSavedAppointment) {
			appointment := sess.GetAppointmentUser(userID)
			ok, err := checkAndCreateAppointmentByUser(userID, appointment)
			if err != nil {
				return logAndMsgErrUser(ctx, errMsg, err)
			}
			if ok {
				serviceInfo, barberName, date, time, err := getAppointmentInfo(ent.Appointment{
					WorkdayID: appointment.WorkdayID,
					ServiceID: appointment.ServiceID,
					Time:      appointment.Time,
					Duration:  appointment.Duration,
				})
				if err != nil {
					return logAndMsgErrUser(ctx, errMsg, err)
				}
				return ctx.Edit(
					fmt.Sprintf(newAppointmentSavedByUser, serviceInfo, barberName, date, time),
					markupUserBackToMainSend,
				)
			}
			return ctx.Edit(failedToSaveAppointment, markupUserFailedToSaveOrRescheduleAppointment)
		}
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	return informUserAppointmentAlreadyExists(ctx, upcomingAppointment)
}

func onUserSelectAnotherTimeForAppointment(ctx tele.Context) error {
	appointment := sess.GetAppointmentUser(ctx.Sender().ID)
	workday, appointments, err := getWorkdayAndAppointments(appointment.WorkdayID)
	if err != nil {
		return logAndMsgErrUser(ctx, "can't handle select another time for appointment", err)
	}
	return calculateAndShowToUserFreeTimesForAppointment(ctx, workday, appointments, appointment)
}

func onRescheduleOrCancelAppointment(ctx tele.Context) error {
	errMsg := "can't handle reschedule or cancel appointment"
	userID := ctx.Sender().ID
	upcomingAppointment, err := cp.RepoWithContext.GetUpcomingAppointment(userID)
	if err != nil {
		if errors.Is(err, rep.ErrNoSavedAppointment) {
			return ctx.Edit(youHaveNoAppointments, markupUserBackToMain)
		}
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	serviceInfo, barberName, date, time, err := getAppointmentInfo(upcomingAppointment)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	return ctx.Edit(
		fmt.Sprintf(rescheduleOrCancelAppointment, barberName, date, time, serviceInfo),
		markupRescheduleOrCancelAppointment,
	)
}

func onUserRescheduleAppointment(ctx tele.Context) error {
	errMsg := "can't show free workdays for reschedule appointment"
	userID := ctx.Sender().ID
	upcomingAppointment, err := cp.RepoWithContext.GetUpcomingAppointment(userID)
	if err != nil {
		if errors.Is(err, rep.ErrNoSavedAppointment) {
			return ctx.Edit(youHaveNoAppointments, markupUserBackToMain)
		}
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	workday, err := cp.RepoWithContext.GetWorkdayByID(upcomingAppointment.WorkdayID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	barber, err := cp.RepoWithContext.GetBarberByID(workday.BarberID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	editedAppointment := sess.Appointment{
		ID:        upcomingAppointment.ID,
		ServiceID: upcomingAppointment.ServiceID,
		Duration:  upcomingAppointment.Duration,
		BarberID:  barber.ID,
	}
	return calculateAndShowToUserFreeWorkdaysForAppointment(ctx, 0, editedAppointment)
}

func onUserConfirmRescheduleAppointment(ctx tele.Context) error {
	errMsg := "can't confirm reschedule appointment for user"
	userID := ctx.Sender().ID
	_, err := cp.RepoWithContext.GetUpcomingAppointment(userID)
	if err != nil {
		if errors.Is(err, rep.ErrNoSavedAppointment) {
			return ctx.Edit(youHaveNoAppointments, markupUserBackToMain)
		}
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	appointment := sess.GetAppointmentUser(userID)
	ok, err := checkAndRescheduleAppointmentByUser(appointment)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	if ok {
		serviceInfo, barberName, date, time, err := getAppointmentInfo(ent.Appointment{
			WorkdayID: appointment.WorkdayID,
			ServiceID: appointment.ServiceID,
			Time:      appointment.Time,
			Duration:  appointment.Duration,
		})
		if err != nil {
			return logAndMsgErrUser(ctx, errMsg, err)
		}
		return ctx.Edit(
			fmt.Sprintf(appointmentRescheduledByUser, serviceInfo, barberName, date, time),
			markupUserBackToMainSend,
		)
	}
	return ctx.Edit(failedToRescheduleAppointment, markupUserFailedToSaveOrRescheduleAppointment)
}

func onUserCancelAppointment(ctx tele.Context) error {
	errMsg := "can't handle user cancel appointment"
	userID := ctx.Sender().ID
	upcomingAppointment, err := cp.RepoWithContext.GetUpcomingAppointment(userID)
	if err != nil {
		if errors.Is(err, rep.ErrNoSavedAppointment) {
			return ctx.Edit(youHaveNoAppointments, markupUserBackToMain)
		}
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	serviceInfo, barberName, date, time, err := getAppointmentInfo(upcomingAppointment)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	return ctx.Edit(
		fmt.Sprintf(confirmCancelAppointment, barberName, date, time, serviceInfo),
		markupUserConfirmCancelAppointment,
	)
}

func onUserConfirmCancelAppointment(ctx tele.Context) error {
	errMsg := "can't confirm cancel appointment for user"
	userID := ctx.Sender().ID
	upcomingAppointment, err := cp.RepoWithContext.GetUpcomingAppointment(userID)
	if err != nil {
		if errors.Is(err, rep.ErrNoSavedAppointment) {
			return ctx.Edit(youHaveNoAppointments, markupUserBackToMain)
		}
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	if err := cp.RepoWithContext.DeleteAppointmentByID(upcomingAppointment.ID); err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	return ctx.Edit(appointmentCanceled, markupUserBackToMain)
}

func onUserSettings(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(settingsMenu, markupUserSettings)
}

func onUserUpdPersonal(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
	userID := ctx.Sender().ID
	user, err := cp.RepoWithContext.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, rep.ErrNoSavedUser) {
			return ctx.Edit(privacyExplanation, markupPrivacyExplanation)
		}
		return logAndMsgErrUser(ctx, "can't show update personal data options for user", err)
	}
	if user.Name == ent.NoName && user.Phone == ent.NoPhone {
		return ctx.Edit(privacyExplanation, markupPrivacyExplanation)
	}
	return ctx.Edit(selectPersonalData, markupUserPersonal)
}

func onUserPrivacy(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(privacyUser, markupUserPrivacy)
}

func onUserAgreeWithPrivacy(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(selectPersonalData, markupUserPersonal)
}

func onUserUpdName(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateUpdName)
	return ctx.Edit(updName)
}

func onUserUpdPhone(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateUpdPhone)
	return ctx.Edit(updPhone)
}

func onUserBackToMain(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(mainMenu, markupUserMain)
}

func onUserBackToMainSend(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
	return ctx.Send(mainMenu, markupUserMain)
}

func onTextUser(ctx tele.Context) error {
	state := sess.GetUserState(ctx.Sender().ID)
	switch state {
	case sess.StateUpdName:
		return onUpdateUserName(ctx)
	case sess.StateUpdPhone:
		return onUpdateUserPhone(ctx)
	default:
		return ctx.Send(unknownCommand)
	}
}

func onUpdateUserName(ctx tele.Context) error {
	errMsg := "can't update user's name"
	userID := ctx.Sender().ID
	text := ctx.Message().Text
	if err := cp.RepoWithContext.CreateUser(ent.User{ID: userID, Name: text}); err != nil {
		if errors.Is(err, rep.ErrAlreadyExists) {
			if err := cp.RepoWithContext.UpdateUser(ent.User{ID: userID, Name: text}); err != nil {
				if errors.Is(err, rep.ErrInvalidUser) {
					return ctx.Send(invalidName)
				}
				sess.UpdateUserState(userID, sess.StateStart)
				return logAndMsgErrUser(ctx, errMsg, err)
			}
			sess.UpdateUserState(userID, sess.StateStart)
			return ctx.Send(updNameSuccess, markupUserPersonal)
		}
		if errors.Is(err, rep.ErrInvalidUser) {
			return ctx.Send(invalidName)
		}
		sess.UpdateUserState(userID, sess.StateStart)
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	sess.UpdateUserState(userID, sess.StateStart)
	return ctx.Send(updNameSuccess, markupUserPersonal)
}

func onUpdateUserPhone(ctx tele.Context) error {
	errMsg := "can't update user's phone"
	userID := ctx.Sender().ID
	text := ctx.Message().Text
	if err := cp.RepoWithContext.CreateUser(ent.User{ID: userID, Phone: text}); err != nil {
		if errors.Is(err, rep.ErrAlreadyExists) {
			if err := cp.RepoWithContext.UpdateUser(ent.User{ID: userID, Phone: text}); err != nil {
				if errors.Is(err, rep.ErrInvalidUser) {
					return ctx.Send(invalidPhone)
				}
				sess.UpdateUserState(userID, sess.StateStart)
				return logAndMsgErrUser(ctx, errMsg, err)
			}
			sess.UpdateUserState(userID, sess.StateStart)
			return ctx.Send(updPhoneSuccess, markupUserPersonal)
		}
		if errors.Is(err, rep.ErrInvalidUser) {
			return ctx.Send(invalidPhone)
		}
		sess.UpdateUserState(userID, sess.StateStart)
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	sess.UpdateUserState(userID, sess.StateStart)
	return ctx.Send(updPhoneSuccess, markupUserPersonal)
}

func calculateAndShowToUserFreeTimesForAppointment(ctx tele.Context, workday ent.Workday, appointments []ent.Appointment, appointment sess.Appointment) error {
	freeTimes := freeTimesForAppointment(workday, appointments, appointment)
	if len(freeTimes) == 0 {
		return calculateAndShowToUserFreeWorkdaysForAppointment(ctx, 0, appointment)
	}
	sess.UpdateAppointmentAndUserState(ctx.Sender().ID, appointment, sess.StateStart)
	serviceInfo := getNullServiceInfo(appointment.ServiceID, appointment.Duration)
	return ctx.Edit(
		fmt.Sprintf(selectTimeForAppointment, serviceInfo, tm.ShowDate(workday.Date)),
		markupSelectTimeForAppointment(
			freeTimes,
			endpntUserSelectTimeForAppointment,
			endpntUserSelectMonthForAppointment,
			endpntUserBackToMain,
		),
	)
}

func calculateAndShowToUserFreeWorkdaysForAppointment(ctx tele.Context, deltaDisplayedMonth int8, appointment sess.Appointment) error {
	displayedDateRange, displayedMonthRange, err := calculateDisplayedRangesForAppointment(deltaDisplayedMonth, appointment)
	if err != nil {
		return logAndMsgErrUser(ctx, "can't show to user free workdays for new appointment", err)
	}
	if displayedMonthRange.firstMonth > displayedMonthRange.lastMonth {
		return informUserNoFreeTime(ctx, appointment.BarberID)
	}
	appointment.LastShownMonth = tm.ParseMonth(displayedDateRange.LastDate)
	sess.UpdateAppointmentAndUserState(ctx.Sender().ID, appointment, sess.StateStart)
	return showToUserFreeWorkdaysForAppointment(ctx, displayedDateRange, displayedMonthRange, appointment)
}

func checkAndCreateAppointmentByUser(userID int64, appointment sess.Appointment) (ok bool, err error) {
	appointmentsMutex.Lock()
	defer appointmentsMutex.Unlock()
	workday, appointments, err := getWorkdayAndAppointments(appointment.WorkdayID)
	if err != nil {
		return
	}
	ok = isTimeForAppointmentAvailable(workday, appointments, appointment)
	if !ok {
		return
	}
	if err = cp.RepoWithContext.CreateUser(ent.User{ID: userID}); err != nil {
		if !errors.Is(err, rep.ErrAlreadyExists) {
			return
		}
	}
	err = cp.RepoWithContext.CreateAppointment(ent.Appointment{
		UserID:    userID,
		WorkdayID: appointment.WorkdayID,
		ServiceID: appointment.ServiceID,
		Time:      appointment.Time,
		Duration:  appointment.Duration,
		CreatedAt: time.Now().Unix(),
	})
	return
}

func checkAndRescheduleAppointmentByUser(appointment sess.Appointment) (ok bool, err error) {
	appointmentsMutex.Lock()
	defer appointmentsMutex.Unlock()
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

func getAppointmentInfo(appointment ent.Appointment) (serviceInfo, barberName, date, time string, err error) {
	defer func() { err = e.WrapIfErr("can't get appointment info", err) }()
	serviceInfo = getNullServiceInfo(appointment.ServiceID, appointment.Duration)
	workday, err := cp.RepoWithContext.GetWorkdayByID(appointment.WorkdayID)
	if err != nil {
		return
	}
	barber, err := cp.RepoWithContext.GetBarberByID(workday.BarberID)
	if err != nil {
		return
	}
	return serviceInfo, barber.Name, tm.ShowDate(workday.Date), appointment.Time.ShortString(), nil
}

func getWorkingBarbers() (workingBarbers []ent.Barber, err error) {
	barbers, err := cp.RepoWithContext.GetAllBarbers()
	if err != nil {
		return nil, err
	}
	today := tm.Today()
	for _, barber := range barbers {
		if barber.Name == ent.NoName || barber.LastWorkdate.Before(today) {
			continue
		}
		services, err := cp.RepoWithContext.GetServicesByBarberID(barber.ID)
		if err != nil {
			return nil, err
		}
		if len(services) > 0 {
			workingBarbers = append(workingBarbers, barber)
		}
	}
	return
}

func informUserAppointmentAlreadyExists(ctx tele.Context, upcomingAppointment ent.Appointment) error {
	serviceInfo, barberName, date, time, err := getAppointmentInfo(upcomingAppointment)
	if err != nil {
		return logAndMsgErrUser(ctx, "can't inform user appointment already exists", err)
	}
	return ctx.Edit(
		fmt.Sprintf(appointmentAlreadyExists, serviceInfo, barberName, date, time),
		markupUserBackToMain,
	)
}

func informUserNoFreeTime(ctx tele.Context, barberID int64) error {
	barber, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return logAndMsgErrUser(ctx, "can't inform user no free time for appointment", err)
	}
	return ctx.Edit(
		fmt.Sprintf(informUserNoFreeTimeForAppointment, barber.Phone, barber.ID),
		markupUserBackToMain, tele.ModeMarkdown,
	)
}

func logAndMsgErrUser(ctx tele.Context, msg string, err error) error {
	log.Print(e.Wrap(msg, err))
	return ctx.Send(errorUser, markupUserBackToMain)
}

func showBarbersForAppointment(ctx tele.Context, userID int64) error {
	workingBarbers, err := getWorkingBarbers()
	if err != nil {
		return logAndMsgErrUser(ctx, "can't show barbers for appointment", err)
	}
	switch len(workingBarbers) {
	case 0:
		sess.UpdateUserState(userID, sess.StateStart)
		return ctx.Edit(noWorkingBarbers, markupUserBackToMain)
	case 1:
		return showToUserServicesForNewAppointment(ctx, workingBarbers[0])
	default:
		sess.UpdateUserState(userID, sess.StateStart)
		return ctx.Edit(selectBarberForAppointment, markupSelectBarberForAppointment(workingBarbers))
	}
}

func showToUserFreeWorkdaysForAppointment(ctx tele.Context, displayedDateRange ent.DateRange, displayedMonthRange monthRange, appointment sess.Appointment) error {
	markupSelectWorkday, err := markupSelectWorkdayForAppointment(
		displayedDateRange,
		displayedMonthRange,
		appointment,
		endpntUserSelectWorkdayForAppointment,
		endpntUserSelectMonthForAppointment,
		endpntUserBackToMain,
	)
	if err != nil {
		return logAndMsgErrUser(ctx, "can't show to user free workdays for appointment", err)
	}
	serviceInfo := getNullServiceInfo(appointment.ServiceID, appointment.Duration)
	return ctx.Edit(fmt.Sprintf(selectDateForAppointment, serviceInfo), markupSelectWorkday)
}

func showToUserServicesForNewAppointment(ctx tele.Context, barber ent.Barber) error {
	sess.UpdateAppointmentAndUserState(
		ctx.Sender().ID,
		sess.Appointment{BarberID: barber.ID},
		sess.StateStart,
	)
	services, err := cp.RepoWithContext.GetServicesByBarberID(barber.ID)
	if err != nil {
		return logAndMsgErrUser(ctx, "can't show to user services for new appointment", err)
	}
	return ctx.Edit(
		fmt.Sprintf(userSelectServiceForAppointment, barber.Name),
		markupSelectService(services, endpntUserSelectServiceForAppointment, endpntUserBackToMain),
	)
}
