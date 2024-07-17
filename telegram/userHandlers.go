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
	errMsg := "can't handle sign up for appointment"
	barbers, err := cp.RepoWithContext.GetAllBarbers()
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	var workingBarbers []ent.Barber
	today := tm.Today()
	for _, barber := range barbers {
		if barber.Name == ent.NoName || barber.LastWorkdate.Before(today) {
			continue
		}
		services, err := cp.RepoWithContext.GetServicesByBarberID(barber.ID)
		if err != nil {
			return logAndMsgErrUser(ctx, errMsg, err)
		}
		if len(services) > 0 {
			workingBarbers = append(workingBarbers, barber)
		}
	}
	switch len(workingBarbers) {
	case 0:
		sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
		return ctx.Edit(noWorkingBarbers, markupBackToMainUser)
	case 1:
		sess.UpdateNewAppointmentAndUserState(
			ctx.Sender().ID,
			sess.Appointment{BarberID: workingBarbers[0].ID},
			sess.StateStart,
		)
		services, err := cp.RepoWithContext.GetServicesByBarberID(workingBarbers[0].ID)
		if err != nil {
			return logAndMsgErrUser(ctx, errMsg, err)
		}
		return ctx.Edit(
			fmt.Sprintf(selectServiceForAppointment, workingBarbers[0].Name),
			markupSelectService(services, endpntServiceForAppointmentUser, endpntBackToMainUser),
		)
	default:
		sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
		return ctx.Edit(selectBarberForAppointment, markupSelectBarberForAppointment(workingBarbers))
	}
}

func onSelectBarberForAppointment(ctx tele.Context) error {
	errMsg := "can't show barber's services for appointment"
	barberID, err := strconv.ParseInt(ctx.Callback().Data, 10, 64)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	sess.UpdateNewAppointmentAndUserState(
		ctx.Sender().ID,
		sess.Appointment{BarberID: barberID},
		sess.StateStart,
	)
	barber, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	services, err := cp.RepoWithContext.GetServicesByBarberID(barberID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	return ctx.Edit(
		fmt.Sprintf(selectServiceForAppointment, barber.Name),
		markupSelectService(services, endpntServiceForAppointmentUser, endpntBackToMainUser),
	)
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
	userID := ctx.Sender().ID
	newAppointment := sess.GetNewAppointmentUser(userID)
	newAppointment.ServiceID = service.ID
	newAppointment.Duration = service.Duration
	if err := showFreeWorkdaysForNewAppointmentUser(ctx, 0, newAppointment); err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	return nil
}

func onSelectMonthForNewAppointmentUser(ctx tele.Context) error {
	errMsg := "can't show free workdays for new appointment"
	delta, err := strconv.ParseInt(ctx.Callback().Data, 10, 8)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	newAppointment := sess.GetNewAppointmentUser(ctx.Sender().ID)
	if err := showFreeWorkdaysForNewAppointmentUser(ctx, int8(delta), newAppointment); err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	return nil
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
	if err := showFreeTimesForNewAppointmentUser(ctx, workday, appointments, newAppointment); err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	return nil
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
		if err := showFreeTimesForNewAppointmentUser(ctx, workday, appointments, newAppointment); err != nil {
			return logAndMsgErrUser(ctx, errMsg, err)
		}
		return nil
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
	appointmentsMutex.Lock()
	workday, appointments, err := getWorkdayAndAppointments(newAppointment.WorkdayID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	if !isTimeForAppointmentAvailable(newAppointment.Time, newAppointment.Duration, workday, appointments) {
		return ctx.Edit(newAppointmentFailed, markupNewAppointmentFailedUser)
	}
	if err := cp.RepoWithContext.CreateUser(ent.User{ID: userID}); err != nil {
		if !errors.Is(err, rep.ErrAlreadyExists) {
			return logAndMsgErrUser(ctx, errMsg, err)
		}
	}
	if err := cp.RepoWithContext.CreateAppointment(ent.Appointment{
		UserID:    userID,
		WorkdayID: newAppointment.WorkdayID,
		ServiceID: newAppointment.ServiceID,
		Time:      newAppointment.Time,
		Duration:  newAppointment.Duration,
		CreatedAt: time.Now().Unix(),
	}); err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	appointmentsMutex.Unlock()
	service, err := cp.RepoWithContext.GetServiceByID(newAppointment.ServiceID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	barber, err := cp.RepoWithContext.GetBarberByID(newAppointment.BarberID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	return ctx.Edit(
		fmt.Sprintf(newAppointmentSavedByUser, service.Info(), barber.Name, tm.ShowDate(workday.Date), newAppointment.Time.ShortString()),
		markupBackToMainUserSend,
	)
}

func onSelectAnotherTimeForNewAppointmentUser(ctx tele.Context) error {
	errMsg := "can't handle select another time for new appointment"
	newAppointment := sess.GetNewAppointmentUser(ctx.Sender().ID)
	workday, appointments, err := getWorkdayAndAppointments(newAppointment.WorkdayID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	if err := showFreeTimesForNewAppointmentUser(ctx, workday, appointments, newAppointment); err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	return nil
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

func logAndMsgErrUser(ctx tele.Context, msg string, err error) error {
	log.Print(e.Wrap(msg, err))
	return ctx.Send(errorUser, markupBackToMainUser)
}

func noTimeForNewAppointmentUser(ctx tele.Context, barberID int64) error {
	barber, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return err
	}
	return ctx.Edit(fmt.Sprintf(noFreeTimeForNewAppointmentUser, barber.Phone, barber.ID), markupBackToMainUser)
}

func showFreeTimesForNewAppointmentUser(ctx tele.Context, workday ent.Workday, appointments []ent.Appointment, newAppointment sess.Appointment) error {
	freeTimes := freeTimesForAppointment(workday, appointments, newAppointment.Duration)
	if len(freeTimes) == 0 {
		if err := showFreeWorkdaysForNewAppointmentUser(ctx, 0, newAppointment); err != nil {
			return err
		}
		return nil
	}
	sess.UpdateNewAppointmentAndUserState(ctx.Sender().ID, newAppointment, sess.StateStart)
	service, err := cp.RepoWithContext.GetServiceByID(newAppointment.ServiceID)
	if err != nil {
		return err
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

func showFreeWorkdaysForNewAppointmentUser(ctx tele.Context, deltaDisplayedMonth int8, newAppointment sess.Appointment) error {
	firstDisplayedDateRange, err := defineFirstDisplayedDateRangeForAppointment(newAppointment)
	if err != nil {
		return err
	}
	if firstDisplayedDateRange.EndDate.After(ent.MonthFromNow(cfg.MaxAppointmentBookingMonths).EndDate) {
		return noTimeForNewAppointmentUser(ctx, newAppointment.BarberID)
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
		return err
	}
	service, err := cp.RepoWithContext.GetServiceByID(newAppointment.ServiceID)
	if err != nil {
		return err
	}
	return ctx.Edit(fmt.Sprintf(selectDateForAppointment, service.Info()), markupSelectWorkday)
}
