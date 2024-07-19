package telegram

import (
	cp "barbershop-bot/contextprovider"
	ent "barbershop-bot/entities"
	cfg "barbershop-bot/lib/config"
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
	return ctx.Send(mainMenu, markupMainUser)
}

func onSignUpForAppointment(ctx tele.Context) error {
	userID := ctx.Sender().ID
	upcomingAppointment, err := cp.RepoWithContext.GetUpcomingAppointment(userID)
	if err != nil {
		if errors.Is(err, rep.ErrNoSavedAppointment) {
			return showBarbersForAppointment(ctx, userID)
		}
		return logAndMsgErrUser(ctx, "can't handle sign up for appointment", err)
	}
	return informUpcomingAppointmentExists(ctx, upcomingAppointment)
}

func onSelectBarberForAppointment(ctx tele.Context) error {
	errMsg := "can't show barber's services for appointment"
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

func onSelectServiceForAppointmentUser(ctx tele.Context) error {
	errMsg := "can't show free workdays for new appointment"
	serviceID, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	service, err := cp.RepoWithContext.GetServiceByID(serviceID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	newAppointment := sess.GetNewAppointmentUser(ctx.Sender().ID)
	newAppointment.ServiceID = service.ID
	newAppointment.Duration = service.Duration
	return showToUserFreeWorkdaysForNewAppointment(ctx, 0, newAppointment)
}

func onSelectMonthForNewAppointmentUser(ctx tele.Context) error {
	delta, err := strconv.ParseInt(ctx.Callback().Data, 10, 8)
	if err != nil {
		return logAndMsgErrUser(ctx, "can't show free workdays for new appointment", err)
	}
	newAppointment := sess.GetNewAppointmentUser(ctx.Sender().ID)
	return showToUserFreeWorkdaysForNewAppointment(ctx, int8(delta), newAppointment)
}

func onSelectWorkdayForNewAppointmentUser(ctx tele.Context) error {
	errMsg := "can't show free time for new appointment"
	workdayID, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	workday, appointments, err := getWorkdayAndAppointments(workdayID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	newAppointment := sess.GetNewAppointmentUser(ctx.Sender().ID)
	newAppointment.WorkdayID = workdayID
	return showToUserFreeTimesForNewAppointment(ctx, workday, appointments, newAppointment)
}

func onSelectTimeForNewAppointmentUser(ctx tele.Context) error {
	errMsg := "can't handle select time for new appointment"
	apptTime, err := strconv.ParseUint(ctx.Callback().Data, 10, 64)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	appointmentTime := tm.Duration(apptTime)
	userID := ctx.Sender().ID
	newAppointment := sess.GetNewAppointmentUser(userID)
	workday, appointments, err := getWorkdayAndAppointments(newAppointment.WorkdayID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	if !isTimeForAppointmentAvailable(appointmentTime, newAppointment.Duration, workday, appointments) {
		return showToUserFreeTimesForNewAppointment(ctx, workday, appointments, newAppointment)
	}
	newAppointment.Time = appointmentTime
	sess.UpdateNewAppointmentAndUserState(userID, newAppointment, sess.StateStart)
	service, err := cp.RepoWithContext.GetServiceByID(newAppointment.ServiceID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	return ctx.Edit(
		fmt.Sprintf(confirmNewAppointment, service.Info(), tm.ShowDate(workday.Date), appointmentTime.ShortString()),
		markupConfirmNewAppointmentUser,
	)
}

func onConfirmNewAppointmentUser(ctx tele.Context) error {
	errMsg := "can't confirm new appointment for user"
	userID := ctx.Sender().ID
	newAppointment := sess.GetNewAppointmentUser(userID)
	ok, date, err := checkAndCreateAppointmentByUser(ctx, newAppointment)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	if ok {
		service, err := cp.RepoWithContext.GetServiceByID(newAppointment.ServiceID)
		if err != nil {
			return logAndMsgErrUser(ctx, errMsg, err)
		}
		barber, err := cp.RepoWithContext.GetBarberByID(newAppointment.BarberID)
		if err != nil {
			return logAndMsgErrUser(ctx, errMsg, err)
		}
		return ctx.Edit(
			fmt.Sprintf(
				newAppointmentSavedByUser,
				service.Info(), barber.Name, tm.ShowDate(date), newAppointment.Time.ShortString(),
			),
			markupBackToMainUserSend,
		)
	}
	return ctx.Edit(newAppointmentFailed, markupNewAppointmentFailedUser)
}

func onSelectAnotherTimeForNewAppointmentUser(ctx tele.Context) error {
	newAppointment := sess.GetNewAppointmentUser(ctx.Sender().ID)
	workday, appointments, err := getWorkdayAndAppointments(newAppointment.WorkdayID)
	if err != nil {
		return logAndMsgErrUser(ctx, "can't handle select another time for new appointment", err)
	}
	return showToUserFreeTimesForNewAppointment(ctx, workday, appointments, newAppointment)
}

func onSettingsUser(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(settingsMenu, markupSettingsUser)
}

func onUpdPersonalUser(ctx tele.Context) error {
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
	return ctx.Edit(selectPersonalData, markupPersonalUser)
}

func onPrivacyUser(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(privacyUser, markupPrivacyUser)
}

func onUserAgreeWithPrivacy(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(selectPersonalData, markupPersonalUser)
}

func onUpdNameUser(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateUpdName)
	return ctx.Edit(updName)
}

func onUpdPhoneUser(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateUpdPhone)
	return ctx.Edit(updPhone)
}

func onBackToMainUser(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(mainMenu, markupMainUser)
}

func onBackToMainUserSend(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
	return ctx.Send(mainMenu, markupMainUser)
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
			return ctx.Send(updNameSuccess, markupPersonalUser)
		}
		if errors.Is(err, rep.ErrInvalidUser) {
			return ctx.Send(invalidName)
		}
		sess.UpdateUserState(userID, sess.StateStart)
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	sess.UpdateUserState(userID, sess.StateStart)
	return ctx.Send(updNameSuccess, markupPersonalUser)
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
			return ctx.Send(updPhoneSuccess, markupPersonalUser)
		}
		if errors.Is(err, rep.ErrInvalidUser) {
			return ctx.Send(invalidPhone)
		}
		sess.UpdateUserState(userID, sess.StateStart)
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	sess.UpdateUserState(userID, sess.StateStart)
	return ctx.Send(updPhoneSuccess, markupPersonalUser)
}

func checkAndCreateAppointmentByUser(ctx tele.Context, newAppointment sess.Appointment) (ok bool, date time.Time, err error) {
	appointmentsMutex.Lock()
	defer appointmentsMutex.Unlock()
	ok, date, err = checkAppointment(newAppointment)
	if err != nil || !ok {
		return
	}
	userID := ctx.Sender().ID
	if err = cp.RepoWithContext.CreateUser(ent.User{ID: userID}); err != nil {
		if !errors.Is(err, rep.ErrAlreadyExists) {
			return
		}
	}
	err = cp.RepoWithContext.CreateAppointment(ent.Appointment{
		UserID:    userID,
		WorkdayID: newAppointment.WorkdayID,
		ServiceID: newAppointment.ServiceID,
		Time:      newAppointment.Time,
		Duration:  newAppointment.Duration,
		CreatedAt: time.Now().Unix(),
	})
	return
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

func informUpcomingAppointmentExists(ctx tele.Context, appointment ent.Appointment) error {
	errMsg := "can't inform that appointment already exists"
	service, err := cp.RepoWithContext.GetServiceByID(appointment.ServiceID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	workday, err := cp.RepoWithContext.GetWorkdayByID(appointment.WorkdayID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	barber, err := cp.RepoWithContext.GetBarberByID(workday.BarberID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	return ctx.Edit(
		fmt.Sprintf(
			appointmentAlreadyExists,
			service.Info(), barber.Name, tm.ShowDate(workday.Date), appointment.Time.ShortString(),
		),
		markupBackToMainUser,
	)
}

func logAndMsgErrUser(ctx tele.Context, msg string, err error) error {
	log.Print(e.Wrap(msg, err))
	return ctx.Send(errorUser, markupBackToMainUser)
}

func showBarbersForAppointment(ctx tele.Context, userID int64) error {
	workingBarbers, err := getWorkingBarbers()
	if err != nil {
		return logAndMsgErrUser(ctx, "can't show barbers for appointment", err)
	}
	switch len(workingBarbers) {
	case 0:
		sess.UpdateUserState(userID, sess.StateStart)
		return ctx.Edit(noWorkingBarbers, markupBackToMainUser)
	case 1:
		return showToUserServicesForNewAppointment(ctx, workingBarbers[0])
	default:
		sess.UpdateUserState(userID, sess.StateStart)
		return ctx.Edit(selectBarberForAppointment, markupSelectBarberForAppointment(workingBarbers))
	}
}

func showToUserFreeTimesForNewAppointment(ctx tele.Context, workday ent.Workday, appointments []ent.Appointment, newAppointment sess.Appointment) error {
	freeTimes := freeTimesForAppointment(workday, appointments, newAppointment.Duration)
	if len(freeTimes) == 0 {
		return showToUserFreeWorkdaysForNewAppointment(ctx, 0, newAppointment)
	}
	sess.UpdateNewAppointmentAndUserState(ctx.Sender().ID, newAppointment, sess.StateStart)
	service, err := cp.RepoWithContext.GetServiceByID(newAppointment.ServiceID)
	if err != nil {
		return logAndMsgErrUser(ctx, "can't show to user free times for new appointment", err)
	}
	return ctx.Edit(
		fmt.Sprintf(selectTimeForAppointment, service.Info(), tm.ShowDate(workday.Date)),
		markupSelectTimeForAppointment(
			freeTimes,
			endpntTimeForNewAppointmentUser,
			endpntMonthForNewAppointmentUser,
			endpntBackToMainUser,
		),
	)
}

func showToUserFreeWorkdaysForNewAppointment(ctx tele.Context, deltaDisplayedMonth int8, newAppointment sess.Appointment) error {
	errMsg := "can't show to user free workdays for new appointment"
	firstDisplayedDateRange, err := defineFirstDisplayedDateRangeForAppointment(newAppointment)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	if firstDisplayedDateRange.EndDate.After(ent.MonthFromNow(cfg.MaxAppointmentBookingMonths).EndDate) {
		barber, err := cp.RepoWithContext.GetBarberByID(newAppointment.BarberID)
		if err != nil {
			return logAndMsgErrUser(ctx, errMsg, err)
		}
		return ctx.Edit(fmt.Sprintf(noFreeTimeForNewAppointmentUser, barber.Phone, barber.ID), markupBackToMainUser)
	}
	displayedMonthRange := ent.DateRange{
		StartDate: firstDisplayedDateRange.EndDate,
		EndDate:   ent.MonthFromNow(cfg.MaxAppointmentBookingMonths).EndDate,
	}
	displayedDateRange := defineDisplayedDateRangeForAppointment(
		firstDisplayedDateRange,
		displayedMonthRange,
		deltaDisplayedMonth,
		newAppointment,
	)
	newAppointment.LastShownDate = displayedDateRange.EndDate
	sess.UpdateNewAppointmentAndUserState(ctx.Sender().ID, newAppointment, sess.StateStart)
	markupSelectWorkday, err := markupSelectWorkdayForAppointment(
		displayedDateRange,
		displayedMonthRange,
		newAppointment,
		endpntWorkdayForNewAppointmentUser,
		endpntMonthForNewAppointmentUser,
		endpntBackToMainUser,
	)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	service, err := cp.RepoWithContext.GetServiceByID(newAppointment.ServiceID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	return ctx.Edit(fmt.Sprintf(selectDateForAppointment, service.Info()), markupSelectWorkday)
}

func showToUserServicesForNewAppointment(ctx tele.Context, barber ent.Barber) error {
	sess.UpdateNewAppointmentAndUserState(
		ctx.Sender().ID,
		sess.Appointment{BarberID: barber.ID},
		sess.StateStart,
	)
	services, err := cp.RepoWithContext.GetServicesByBarberID(barber.ID)
	if err != nil {
		return logAndMsgErrUser(ctx, "can't show to user services for new appointment", err)
	}
	return ctx.Edit(
		fmt.Sprintf(selectServiceForAppointment, barber.Name),
		markupSelectService(services, endpntServiceForAppointmentUser, endpntBackToMainUser),
	)
}
