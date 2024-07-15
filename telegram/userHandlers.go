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
		return ctx.Send(noWorkingBarbers, markupBackToMainUser)
	case 1:
		sess.UpdateNewAppointmentAndState(ctx.Sender().ID, sess.NewAppointment{BarberID: workingBarbers[0].ID}, sess.StateStart)
		services, err := cp.RepoWithContext.GetServicesByBarberID(workingBarbers[0].ID)
		if err != nil {
			return logAndMsgErrUser(ctx, errMsg, err)
		}
		return ctx.Send(fmt.Sprintf(selectServiceForAppointment, workingBarbers[0].Name), markupSelectServiceForAppointment(services))
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
	sess.UpdateNewAppointmentAndState(ctx.Sender().ID, sess.NewAppointment{BarberID: barberID}, sess.StateStart)
	barber, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	services, err := cp.RepoWithContext.GetServicesByBarberID(barberID)
	if err != nil {
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	return ctx.Send(fmt.Sprintf(selectServiceForAppointment, barber.Name), markupSelectServiceForAppointment(services))
}

func onSelectServiceForAppointment(ctx tele.Context) error {
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
	newAppointment := sess.GetNewAppointment(userID)
	newAppointment.ServiceID = service.ID
	newAppointment.Duration = service.Duration
	return showFreeWorkdaysForAppointment(ctx, 0, newAppointment, errMsg)
}

func onSelectMonthForAppointment(ctx tele.Context) error {
	errMsg := "can't show free workdays for appointment"
	delta, err := strconv.ParseInt(ctx.Callback().Data, 10, 8)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	newAppointment := sess.GetNewAppointment(ctx.Sender().ID)
	return showFreeWorkdaysForAppointment(ctx, int8(delta), newAppointment, errMsg)
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

func earlestDateWithFreeTime(newAppointment sess.NewAppointment, dateRange ent.DateRange) (earlestFreeDate time.Time, ok bool, err error) {
	workdays, err := cp.RepoWithContext.GetWorkdaysByDateRange(newAppointment.BarberID, dateRange)
	if err != nil {
		return time.Time{}, false, err
	}
	appts, err := cp.RepoWithContext.GetAppointmentsByDateRange(newAppointment.BarberID, dateRange)
	if err != nil {
		return time.Time{}, false, err
	}
	appointments := make(map[int][]ent.Appointment)
	for _, appt := range appts {
		appointments[appt.WorkdayID] = append(appointments[appt.WorkdayID], appt)
	}
	for _, workday := range workdays {
		if haveFreeTimeForAppointment(workday, appointments[workday.ID], newAppointment.Duration) {
			return workday.Date, true, nil
		}
	}
	return time.Time{}, false, nil
}

func logAndMsgErrUser(ctx tele.Context, msg string, err error) error {
	log.Print(e.Wrap(msg, err))
	return ctx.Send(errorUser, markupBackToMainUser)
}

func showFreeWorkdaysForAppointment(ctx tele.Context, deltaDisplayedMonth int8, newAppointment sess.NewAppointment, errMsg string) error {
	var relativeFirstDisplayedMonth byte = 0
	var displayedDateRange ent.DateRange
	for relativeFirstDisplayedMonth <= cfg.MaxAppointmentBookingMonths {
		displayedDateRange = ent.MonthFromNow(relativeFirstDisplayedMonth)
		earlestFreeDate, ok, err := earlestDateWithFreeTime(newAppointment, displayedDateRange)
		if err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		if ok {
			if earlestFreeDate.After(displayedDateRange.StartDate) {
				displayedDateRange.StartDate = earlestFreeDate
			}
			break
		}
		relativeFirstDisplayedMonth++
	}
	if relativeFirstDisplayedMonth > cfg.MaxAppointmentBookingMonths {
		barber, err := cp.RepoWithContext.GetBarberByID(newAppointment.BarberID)
		if err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		return ctx.Edit(fmt.Sprintf(noFreeTimeForAppointment, barber.Phone, barber.ID), markupBackToMainUser)
	}
	firstDisplayedMonth := displayedDateRange
	lastDisplayedMonth := ent.MonthFromNow(cfg.MaxAppointmentBookingMonths)
	if deltaDisplayedMonth != 0 {
		newDateRange := ent.MonthFromBase(newAppointment.LastShownDate, deltaDisplayedMonth)
		if newDateRange.EndDate.After(lastDisplayedMonth.EndDate) || newDateRange.EndDate.Before(firstDisplayedMonth.EndDate) {
			if newAppointment.LastShownDate.After(displayedDateRange.EndDate) {
				displayedDateRange = ent.MonthFromBase(newAppointment.LastShownDate, 0)
			}
		}
		if newDateRange.EndDate.After(displayedDateRange.EndDate) {
			displayedDateRange = newDateRange
		}
	}
	newAppointment.LastShownDate = displayedDateRange.EndDate
	sess.UpdateNewAppointmentAndState(ctx.Sender().ID, newAppointment, sess.StateStart)
	markupSelectWorkday, err := markupSelectWorkdayForAppointment(displayedDateRange, firstDisplayedMonth.EndDate, lastDisplayedMonth.EndDate, newAppointment)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	service, err := cp.RepoWithContext.GetServiceByID(newAppointment.ServiceID)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	return ctx.Edit(fmt.Sprintf(selectDateForAppointment, service.Info()), markupSelectWorkday)
}
