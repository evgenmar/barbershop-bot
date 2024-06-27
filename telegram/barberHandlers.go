package telegram

import (
	cp "barbershop-bot/contextprovider"
	ent "barbershop-bot/entities"
	cfg "barbershop-bot/lib/config"
	"barbershop-bot/lib/e"
	tm "barbershop-bot/lib/time"
	rep "barbershop-bot/repository"
	m "barbershop-bot/repository/mappers"
	sched "barbershop-bot/scheduler"
	sess "barbershop-bot/sessions"
	"errors"
	"log"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"
)

func onStartBarber(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Send(mainBarber, markupMainBarber)
}

func onSettingsBarber(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(settingsBarber, markupSettingsBarber)
}

func onManageAccountBarber(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(manageAccountBarber, markupManageAccountBarber)
}

func onShowCurrentSettingsBarber(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	barber, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return logAndMsgErrBarber(ctx, "can't show current barber settings", err)
	}
	return ctx.Edit(currentSettings+barber.PersonalInfo(), markupBackToMainBarber)
}

func onUpdPersonalBarber(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(personalBarber, markupPersonalBarber)
}

func onUpdNameBarber(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateUpdName)
	return ctx.Edit(updNameBarber)
}

func onUpdPhoneBarber(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateUpdPhone)
	return ctx.Edit(updPhoneBarber)
}

func onDeleteAccount(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(deleteAccount, markupDeleteAccount)
}

func onSetLastWorkDate(ctx tele.Context) error {
	errMsg := "can't open select last work date menu"
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	latestAppointmentDate, err := cp.RepoWithContext.GetLatestAppointmentDate(barberID)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	var firstDisplayedDateRange ent.DateRange
	relativeFirstDisplayedMonth := 0
	for relativeFirstDisplayedMonth <= cfg.MaxAppointmentBookingMonths {
		firstDisplayedDateRange = ent.Month(byte(relativeFirstDisplayedMonth))
		if latestAppointmentDate.Compare(firstDisplayedDateRange.EndDate) <= 0 {
			if latestAppointmentDate.After(firstDisplayedDateRange.StartDate) {
				firstDisplayedDateRange.StartDate = latestAppointmentDate
			}
			break
		}
		relativeFirstDisplayedMonth++
	}
	deltaDisplayedMonth, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	markupSelectDate := markupSelectLastWorkDate(firstDisplayedDateRange, relativeFirstDisplayedMonth, deltaDisplayedMonth)
	return ctx.Edit(selectLastWorkDate, markupSelectDate)
}

func markupSelectLastWorkDate(firstDisplayedDateRange ent.DateRange, relativeFirstDisplayedMonth, deltaDisplayedMonth int) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	var btnPrevMonth, btnNextMonth tele.Btn
	var displayedDateRange ent.DateRange
	if deltaDisplayedMonth == 0 {
		displayedDateRange = firstDisplayedDateRange
		btnPrevMonth = btnEmpty
		btnNextMonth = markup.Data(next, endpntSelectMonthOfLastWorkDate, strconv.Itoa(1))
	} else {
		displayedDateRange = ent.Month(byte(relativeFirstDisplayedMonth + deltaDisplayedMonth))
		btnPrevMonth = markup.Data(prev, endpntSelectMonthOfLastWorkDate, strconv.Itoa(deltaDisplayedMonth-1))
		relativeMaxDisplayedMonth := int(cfg.ScheduledWeeks) * 7 / 30
		if relativeFirstDisplayedMonth+deltaDisplayedMonth == relativeMaxDisplayedMonth {
			btnNextMonth = btnEmpty
		} else {
			btnNextMonth = markup.Data(next, endpntSelectMonthOfLastWorkDate, strconv.Itoa(deltaDisplayedMonth+1))
		}
	}
	rowSelectMonth := markup.Row(btnPrevMonth, btnMonth(displayedDateRange.Month()), btnNextMonth)
	var btnsDatesToSelect []tele.Btn
	for i := 1; i < displayedDateRange.StartWeekday(); i++ {
		btnsDatesToSelect = append(btnsDatesToSelect, btnEmpty)
	}
	for date := displayedDateRange.StartDate; date.Compare(displayedDateRange.EndDate) <= 0; date = date.Add(24 * time.Hour) {
		btnDateToSelect := markup.Data(strconv.Itoa(date.Day()), endpntSelectLastWorkDate, m.MapToStorage.Date(date))
		btnsDatesToSelect = append(btnsDatesToSelect, btnDateToSelect)
	}
	for i := 7; i > displayedDateRange.EndWeekday(); i-- {
		btnsDatesToSelect = append(btnsDatesToSelect, btnEmpty)
	}
	rowsSelectDate := markup.Split(7, btnsDatesToSelect)
	rowRestoreDefaultDate := markup.Row(markup.Data("Установить бессрочную дату окончания работы", endpntSelectLastWorkDate, cfg.InfiniteWorkDate))
	rowBackToMainBarber := markup.Row(btnBackToMainBarber)
	var rows []tele.Row
	rows = append(rows, rowSelectMonth, rowWeekdays)
	rows = append(rows, rowsSelectDate...)
	rows = append(rows, rowRestoreDefaultDate, rowBackToMainBarber)
	markup.Inline(rows...)
	return markup
}

func onSelectLastWorkDate(ctx tele.Context) error {
	errMsg := "can't save last work date"
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	dateToSave, err := m.MapToEntity.Date(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	barber, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	switch dateToSave.Compare(barber.LastWorkdate) {
	case 0:
		return ctx.Edit(lastWorkDateUnchanged, markupBackToMainBarber)
	case 1:
		if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: barberID, LastWorkdate: dateToSave}); err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		if err := sched.MakeSchedule(barberID); err != nil {
			log.Print(e.Wrap(errMsg, err))
			// TODO: ensure atomicity using outbox pattern
			return ctx.Send(lastWorkDateSavedWithoutSсhedule, markupBackToMainBarber)
		}
		return ctx.Edit(lastWorkDateSaved, markupBackToMainBarber)
	case -1:
		latestWorkDate, err := cp.RepoWithContext.GetLatestWorkDate(barberID)
		if err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		dateRangeToDelete := ent.DateRange{StartDate: dateToSave.Add(24 * time.Hour), EndDate: latestWorkDate}
		err = cp.RepoWithContext.DeleteWorkdaysByDateRange(barberID, dateRangeToDelete)
		if err != nil && !errors.Is(err, rep.ErrInvalidDateRange) {
			if errors.Is(err, rep.ErrAppointmentsExists) {
				return ctx.Edit(haveAppointmentAfterDataToSave, markupBackToMainBarber)
			}
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: barberID, LastWorkdate: dateToSave}); err != nil {
			// TODO: ensure atomicity using outbox pattern
			return ctx.Send(lastWorkDateNotSavedButScheduleDeleted, markupBackToMainBarber)
		}
		return ctx.Edit(lastWorkDateSaved, markupBackToMainBarber)
	default:
		return nil
	}
}

func onSelfDeleteBarber(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	barberToDelete, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return logAndMsgErrBarber(ctx, "can't provide options for barber self deletion", err)
	}
	if barberToDelete.LastWorkdate.Before(tm.Today()) {
		return ctx.Edit(confirmSelfDeletion, markupConfirmSelfDeletion)
	}
	return ctx.Edit(youHaveActiveSchedule+preDeletionBarberInstruction, markupBackToMainBarber)
}

func onSureToDelete(ctx tele.Context) error {
	errMsg := "can't self delete barber"
	barberIDToDelete := ctx.Sender().ID
	if err := cp.RepoWithContext.DeleteAppointmentsBeforeDate(barberIDToDelete, tm.Today()); err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	if err := cp.RepoWithContext.DeleteBarberByID(barberIDToDelete); err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	cfg.Barbers.RemoveID(barberIDToDelete)
	return ctx.Edit(goodbuyBarber)
}

func onManageServices(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(manageServices, markupManageServices)
}

func onShowMyServices(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	services, err := cp.RepoWithContext.GetServicesByBarberID(ctx.Sender().ID)
	if err != nil {
		return logAndMsgErrBarber(ctx, "can't show list of services", err)
	}
	if len(services) == 0 {
		return ctx.Edit(youHaveNoServices, markupShowNoServices)
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
	return showNewServiceOptions(ctx, newService)
}

func onСontinueOldService(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	newService := sess.GetNewService(barberID)
	return showNewServiceOptions(ctx, newService)
}

func onMakeNewService(ctx tele.Context) error {
	sess.UpdateNewServiceAndState(ctx.Sender().ID, sess.NewService{}, sess.StateStart)
	return showNewServiceOptions(ctx, sess.NewService{})
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

func onSelectServiceDuration(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(selectServiceDuration, markupSelectServiceDuration())
}

func markupSelectServiceDuration() *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	btnsDurationsToSelect := make([]tele.Btn, 4)
	for duration := 30 * tm.Minute; duration <= 2*tm.Hour; duration += 30 * tm.Minute {
		btnsDurationsToSelect = append(btnsDurationsToSelect, btnServiceDuration(duration))
	}
	rows := markup.Split(2, btnsDurationsToSelect)
	rows = append(rows, markup.Row(btnBackToMainBarber))
	markup.Inline(rows...)
	return markup
}

func onSelectCertainDuration(ctx tele.Context) error {
	dur, err := strconv.ParseUint(ctx.Callback().Data, 10, 64)
	if err != nil {
		return logAndMsgErrBarber(ctx, "can't handle select certain service duration", err)
	}
	barberID := ctx.Sender().ID
	newService := sess.GetNewService(barberID)
	newService.Duration = tm.Duration(dur)
	sess.UpdateNewServiceAndState(barberID, newService, sess.StateStart)
	return showNewServiceOptions(ctx, newService)
}

func showNewServiceOptions(ctx tele.Context, newService sess.NewService) error {
	if newService.Name != "" && newService.Desciption != "" && newService.Price != 0 && newService.Duration != 0 {
		return ctx.Edit(readyToCreateService+"\n\n"+enterServiceParams+"\n\n"+newService.Info(), markupReadyToCreateService, tele.ModeMarkdown)
	}
	return ctx.Edit(enterServiceParams+"\n\n"+newService.Info(), markupEnterServiceParams, tele.ModeMarkdown)
}

func onSaveNewService(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	newService := sess.GetNewService(barberID)
	serviceToCreate := ent.Service{
		BarberID:   barberID,
		Name:       newService.Name,
		Desciption: newService.Desciption,
		Price:      newService.Price,
		Duration:   newService.Duration,
	}
	if err := cp.RepoWithContext.CreateService(serviceToCreate); err != nil {
		sess.UpdateBarberState(barberID, sess.StateStart)
		if errors.Is(err, rep.ErrInvalidService) {
			return ctx.Edit(invalidService, markupBackToMainBarber)
		}
		if errors.Is(err, rep.ErrAlreadyExists) {
			return ctx.Edit(nonUniqueServiceName+"\n\n"+newService.Info(), markupEnterServiceName, tele.ModeMarkdown)
		}
		return logAndMsgErrBarber(ctx, "can't create service", err)
	}
	sess.UpdateNewServiceAndState(barberID, sess.NewService{}, sess.StateStart)
	return ctx.Edit(serviceCreated, markupManageServices)
}

func onManageBarbers(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(manageBarbers, markupManageBarbers)
}

func onShowAllBarbers(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	barberIDs := cfg.Barbers.IDs()
	barbersInfo := ""
	for _, barberID := range barberIDs {
		barber, err := cp.RepoWithContext.GetBarberByID(barberID)
		if err != nil {
			return logAndMsgErrBarber(ctx, "can't show all barbers", err)
		}
		barbersInfo = barbersInfo + "\n\n" + barber.PuplicInfo()
	}
	return ctx.Edit(listOfBarbers+barbersInfo, markupBackToMainBarber)
}

func onAddBarber(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateAddBarber)
	return ctx.Edit(addBarber)
}

func onDeleteBarber(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	barberIDs := cfg.Barbers.IDs()
	if len(barberIDs) == 1 {
		return ctx.Edit(onlyOneBarberExists, markupBackToMainBarber)
	}
	markupSelectBarber, notEmpty, err := markupSelectBarberToDeletion(ctx, barberIDs)
	if err != nil {
		return logAndMsgErrBarber(ctx, "can't suggest actions to delete barber", err)
	}
	if !notEmpty {
		return ctx.Edit(noBarbersToDelete+preDeletionBarberInstruction, markupSelectBarber)
	}
	return ctx.Edit(selectBarberToDeletion+preDeletionBarberInstruction, markupSelectBarber)
}

func markupSelectBarberToDeletion(ctx tele.Context, barberIDs []int64) (*tele.ReplyMarkup, bool, error) {
	today := tm.Today()
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	for _, barberID := range barberIDs {
		if barberID != ctx.Sender().ID {
			barber, err := cp.RepoWithContext.GetBarberByID(barberID)
			if err != nil {
				return &tele.ReplyMarkup{}, false, e.Wrap("can't make reply markup", err)
			}
			if barber.LastWorkdate.Before(today) {
				barberIDString := strconv.FormatInt(barberID, 10)
				row := markup.Row(markup.Data(barber.Name, endpntBarberToDeletion, barberIDString))
				rows = append(rows, row)
			}
		}
	}
	notEmpty := len(rows) > 0
	rows = append(rows, markup.Row(btnBackToMainBarber))
	markup.Inline(rows...)
	return markup, notEmpty, nil
}

func onDeleteCertainBarber(ctx tele.Context) error {
	errMsg := "can't delete barber"
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	barberIDToDelete, err := strconv.ParseInt(ctx.Callback().Data, 10, 64)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	barberToDelete, err := cp.RepoWithContext.GetBarberByID(barberIDToDelete)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	today := tm.Today()
	if barberToDelete.LastWorkdate.Before(today) {
		if err := cp.RepoWithContext.DeleteAppointmentsBeforeDate(barberIDToDelete, today); err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		if err := cp.RepoWithContext.DeleteBarberByID(barberIDToDelete); err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		cfg.Barbers.RemoveID(barberIDToDelete)
		return ctx.Edit(barberDeleted, markupBackToMainBarber)
	}
	return ctx.Edit(barberHaveActiveSchedule, markupBackToMainBarber)
}

func onBackToMainBarber(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(mainBarber, markupMainBarber)
}

func onTextBarber(ctx tele.Context) error {
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
	default:
		return ctx.Send(unknownCmdBarber)
	}
}

func onUpdateBarberName(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: barberID, Name: ctx.Message().Text}); err != nil {
		if errors.Is(err, rep.ErrInvalidBarber) {
			log.Print(e.Wrap("invalid barber name", err))
			return ctx.Send(invalidBarberName)
		}
		if errors.Is(err, rep.ErrNonUniqueData) {
			log.Print(e.Wrap("barber's name must be unique", err))
			return ctx.Send(notUniqueBarberName)
		}
		sess.UpdateBarberState(barberID, sess.StateStart)
		return logAndMsgErrBarber(ctx, "can't update barber's name", err)
	}
	sess.UpdateBarberState(barberID, sess.StateStart)
	return ctx.Send(updNameSuccessBarber, markupPersonalBarber)
}

func onUpdateBarberPhone(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: barberID, Phone: ctx.Message().Text}); err != nil {
		if errors.Is(err, rep.ErrInvalidBarber) {
			log.Print(e.Wrap("invalid barber phone", err))
			return ctx.Send(invalidBarberPhone)
		}
		if errors.Is(err, rep.ErrNonUniqueData) {
			log.Print(e.Wrap("barber's phone must be unique", err))
			return ctx.Send(notUniqueBarberPhone)
		}
		sess.UpdateBarberState(barberID, sess.StateStart)
		return logAndMsgErrBarber(ctx, "can't update barber's phone", err)
	}
	sess.UpdateBarberState(barberID, sess.StateStart)
	return ctx.Send(updPhoneSuccessBarber, markupPersonalBarber)
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
	return showNewServOptions(ctx, newService)
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
	return showNewServOptions(ctx, newService)
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
	return showNewServOptions(ctx, newService)
}

func showNewServOptions(ctx tele.Context, newService sess.NewService) error {
	if newService.Name != "" && newService.Desciption != "" && newService.Price != 0 && newService.Duration != 0 {
		return ctx.Send(readyToCreateService+"\n\n"+enterServiceParams+"\n\n"+newService.Info(), markupReadyToCreateService, tele.ModeMarkdown)
	}
	return ctx.Send(enterServiceParams+"\n\n"+newService.Info(), markupEnterServiceParams, tele.ModeMarkdown)
}

func onContactBarber(ctx tele.Context) error {
	state := sess.GetBarberState(ctx.Sender().ID)
	switch state {
	case sess.StateAddBarber:
		return onAddNewBarber(ctx)
	default:
		return ctx.Send(unknownCmdBarber)
	}
}

func onAddNewBarber(ctx tele.Context) error {
	errMsg := "can't add new barber"
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	newBarberID := ctx.Message().Contact.UserID
	if err := cp.RepoWithContext.CreateBarber(newBarberID); err != nil {
		if errors.Is(err, rep.ErrAlreadyExists) {
			return ctx.Send(userIsAlreadyBarber, markupBackToMainBarber)
		}
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	cfg.Barbers.AddID(newBarberID)
	if err := sched.MakeSchedule(newBarberID); err != nil {
		log.Print(e.Wrap(errMsg, err))
		// TODO: ensure atomicity using outbox pattern
		return ctx.Send(addedNewBarberWithoutSсhedule, markupBackToMainBarber)
	}
	return ctx.Send(addedNewBarberWithSсhedule, markupBackToMainBarber)
}

func logAndMsgErrBarber(ctx tele.Context, msg string, err error) error {
	log.Print(e.Wrap(msg, err))
	return ctx.Send(errorBarber, markupBackToMainBarber)
}
