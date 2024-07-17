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
		return ctx.Send(noWorkingBarbers, markupBackToMainUser)
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
		return ctx.Send(
			fmt.Sprintf(selectServiceForAppointment, workingBarbers[0].Name),
			markupSelectService(services, endpntServiceForAppointmentUser, endpntBackToMainUser),
		)
	default:
		sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
		return ctx.Send(selectBarberForAppointment, markupSelectBarberForAppointment(workingBarbers))
	}
}

func onSelectBarberForAppointment(ctx tele.Context) error {
	errMsg := "can't show barber's services for appointment"
	barberID, err := strconv.ParseInt(ctx.Callback().Data, 10, 64)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
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
	return ctx.Send(
		fmt.Sprintf(selectServiceForAppointment, barber.Name),
		markupSelectService(services, endpntServiceForAppointmentUser, endpntBackToMainUser),
	)
}

func onSelectServiceForAppointmentUser(ctx tele.Context) error {
	errMsg := "can't show free workdays for appointment"
	serviceID, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	service, err := cp.RepoWithContext.GetServiceByID(serviceID)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	userID := ctx.Sender().ID
	newAppointment := sess.GetNewAppointmentUser(userID)
	newAppointment.ServiceID = service.ID
	newAppointment.Duration = service.Duration
	if err := showFreeWorkdaysForNewAppointmentUser(ctx, 0, newAppointment); err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	return nil
}

func onSelectMonthForNewAppointmentUser(ctx tele.Context) error {
	errMsg := "can't show free workdays for appointment"
	delta, err := strconv.ParseInt(ctx.Callback().Data, 10, 8)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	newAppointment := sess.GetNewAppointmentUser(ctx.Sender().ID)
	if err := showFreeWorkdaysForNewAppointmentUser(ctx, int8(delta), newAppointment); err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	return nil
}

func onSelectWorkdayForNewAppointmentUser(ctx tele.Context) error {
	errMsg := "can't show free time for appointment"
	workdayID, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	workday, err := cp.RepoWithContext.GetWorkdayByID(workdayID)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	dateRange := ent.DateRange{StartDate: workday.Date, EndDate: workday.Date}
	appointments, err := cp.RepoWithContext.GetAppointmentsByDateRange(workday.BarberID, dateRange)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	userID := ctx.Sender().ID
	newAppointment := sess.GetNewAppointmentUser(userID)
	newAppointment.WorkdayID = workdayID
	freeTimes := freeTimesForAppointment(workday, appointments, newAppointment.Duration)
	if len(freeTimes) == 0 {
		if err := showFreeWorkdaysForNewAppointmentUser(ctx, 0, newAppointment); err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		return nil
	}
	sess.UpdateNewAppointmentAndUserState(userID, newAppointment, sess.StateStart)
	markupSelectTime := markupSelectTimeForAppointment(
		freeTimes,
		endpntTimeForNewAppointmentUser,
		endpntMonthForNewAppointmentUser,
		endpntBackToMainUser,
	)
	service, err := cp.RepoWithContext.GetServiceByID(newAppointment.ServiceID)
	if err != nil {
		return err
	}
	return ctx.Edit(fmt.Sprintf(selectTimeForAppointment, service.Info(), tm.ShowDate(workday.Date)), markupSelectTime)
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
