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
	"fmt"
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

func onListOfNecessarySettings(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(listOfNecessarySettings, markupShortSettingsBarber)
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
	return ctx.Edit(currentSettings+barber.Info(), markupBackToMainBarber)
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
	delta, err := strconv.ParseUint(ctx.Callback().Data, 10, 8)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	deltaDisplayedMonth := byte(delta)
	var firstDisplayedDateRange, displayedDateRange ent.DateRange
	var relativeFirstDisplayedMonth byte = 0
	var maxDeltaDisplayedMonth byte
	for relativeFirstDisplayedMonth <= cfg.MaxAppointmentBookingMonths {
		firstDisplayedDateRange = ent.Month(relativeFirstDisplayedMonth)
		if latestAppointmentDate.Compare(firstDisplayedDateRange.EndDate) <= 0 {
			if latestAppointmentDate.After(firstDisplayedDateRange.StartDate) {
				firstDisplayedDateRange.StartDate = latestAppointmentDate
			}
			break
		}
		relativeFirstDisplayedMonth++
	}
	if deltaDisplayedMonth == 0 {
		displayedDateRange = firstDisplayedDateRange
	} else {
		displayedDateRange = ent.Month(relativeFirstDisplayedMonth + deltaDisplayedMonth)
	}
	if cfg.ScheduledWeeks*7/30 > relativeFirstDisplayedMonth {
		maxDeltaDisplayedMonth = cfg.ScheduledWeeks*7/30 - relativeFirstDisplayedMonth
	} else {
		maxDeltaDisplayedMonth = 0
	}
	return ctx.Edit(selectLastWorkDate, markupSelectLastWorkDate(displayedDateRange, deltaDisplayedMonth, maxDeltaDisplayedMonth))
}

func onSelectLastWorkDate(ctx tele.Context) error {
	errMsg := "can't save last work date"
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	dateToSave, err := time.ParseInLocation(time.DateOnly, ctx.Callback().Data, cfg.Location)
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

func onSureToSelfDeleteBarber(ctx tele.Context) error {
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
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	services, err := cp.RepoWithContext.GetServicesByBarberID(barberID)
	if err != nil {
		return logAndMsgErrBarber(ctx, "can't show manage services menu", err)
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
		return logAndMsgErrBarber(ctx, "can't show list of services", err)
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
		return logAndMsgErrBarber(ctx, "can't select certain service duration", err)
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
		return logAndMsgErrBarber(ctx, "can't show list of services for editing", err)
	}
	if len(services) == 0 {
		return ctx.Edit(youHaveNoServices, markupManageServicesShort)
	}
	return ctx.Edit(selectServiceToEdit, markupSelectServiceBarber(services, endpntServiceToEdit))
}

func onSelectCertainServiceToEdit(ctx tele.Context) error {
	errMsg := "can't select certain service for editing"
	serviceID, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	serviceToEdit, err := cp.RepoWithContext.GetServiceByID(serviceID)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
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
		return logAndMsgErrBarber(ctx, "can't select certain service duration", err)
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
	serviceToUpdate := ent.Service{
		ID:         editedService.ID,
		Name:       editedService.UpdService.Name,
		Desciption: editedService.UpdService.Desciption,
		Price:      editedService.UpdService.Price,
		Duration:   editedService.UpdService.Duration,
	}
	if err := cp.RepoWithContext.UpdateService(serviceToUpdate); err != nil {
		sess.UpdateBarberState(barberID, sess.StateStart)
		if errors.Is(err, rep.ErrInvalidService) {
			return ctx.Edit(invalidService, markupBackToMainBarber)
		}
		if errors.Is(err, rep.ErrNonUniqueData) {
			return ctx.Edit(nonUniqueServiceName+"\n\n"+editedService.Info(), markupEditServiceName)
		}
		return logAndMsgErrBarber(ctx, "can't update service", err)
	}
	sess.UpdateEditedServiceAndState(barberID, sess.EditedService{}, sess.StateStart)
	return ctx.Edit(serviceUpdated, markupManageServicesFull)
}

func onDeleteService(ctx tele.Context) error {
	barberID := ctx.Sender().ID
	sess.UpdateBarberState(barberID, sess.StateStart)
	services, err := cp.RepoWithContext.GetServicesByBarberID(barberID)
	if err != nil {
		return logAndMsgErrBarber(ctx, "can't show list of services for delete", err)
	}
	if len(services) == 0 {
		return ctx.Edit(youHaveNoServices, markupManageServicesShort)
	}
	return ctx.Edit(selectServiceToDelete, markupSelectServiceBarber(services, endpntServiceToDelete))
}

func onSelectCertainServiceToDelete(ctx tele.Context) error {
	errMsg := "can't select certain service for delete"
	serviceID, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	serviceToDelete, err := cp.RepoWithContext.GetServiceByID(serviceID)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	return ctx.Edit(fmt.Sprintf(confirmServiceDeletion, serviceToDelete.Info()), markupConfirmServiceDeletion(serviceID))
}

func onSureToDeleteService(ctx tele.Context) error {
	errMsg := "can't delete service"
	serviceID, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	if err := cp.RepoWithContext.DeleteServiceByID(serviceID); err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	return ctx.Edit(serviceDeleted, markupBackToMainBarber)
}

func onManageBarbers(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(manageBarbers, markupManageBarbers)
}

func onShowAllBarbers(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	barbers, err := cp.RepoWithContext.GetAllBarbers()
	if err != nil {
		return logAndMsgErrBarber(ctx, "can't show all barbers", err)
	}
	barbersInfo := ""
	for _, barber := range barbers {
		barbersInfo = barbersInfo + "\n\n" + barber.Info()
	}
	return ctx.Edit(listOfBarbers+barbersInfo, markupBackToMainBarber)
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
		return logAndMsgErrBarber(ctx, "can't suggest actions to delete barber", err)
	}
	if len(barbers) == 1 {
		return ctx.Edit(onlyOneBarberExists, markupBackToMainBarber)
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
	case sess.StateEditServiceName:
		return onEditServName(ctx)
	case sess.StateEditServiceDescription:
		return onEditServDescription(ctx)
	case sess.StateEditServicePrice:
		return onEditServPrice(ctx)
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

func showEditServOptsWithEditMsg(ctx tele.Context, editedService sess.EditedService) error {
	if editedService.UpdService.Name != "" || editedService.UpdService.Desciption != "" || editedService.UpdService.Price != 0 || editedService.UpdService.Duration != 0 {
		return ctx.Edit(editServiceParams+editedService.Info()+readyToUpdateService, markupReadyToUpdateService)
	}
	return ctx.Edit(editServiceParams+editedService.Info(), markupEditServiceParams)
}

func showEditServOptsWithSendMsg(ctx tele.Context, editedService sess.EditedService) error {
	if editedService.UpdService.Name != "" || editedService.UpdService.Desciption != "" || editedService.UpdService.Price != 0 || editedService.UpdService.Duration != 0 {
		return ctx.Send(editServiceParams+editedService.Info()+readyToUpdateService, markupReadyToUpdateService)
	}
	return ctx.Send(editServiceParams+editedService.Info(), markupEditServiceParams)
}

func showNewServOptsWithEditMsg(ctx tele.Context, newService sess.NewService) error {
	if newService.Name != "" && newService.Desciption != "" && newService.Price != 0 && newService.Duration != 0 {
		return ctx.Edit(enterServiceParams+newService.Info()+readyToCreateService, markupReadyToCreateService, tele.ModeMarkdown)
	}
	return ctx.Edit(enterServiceParams+newService.Info(), markupEnterServiceParams, tele.ModeMarkdown)
}

func showNewServOptsWithSendMsg(ctx tele.Context, newService sess.NewService) error {
	if newService.Name != "" && newService.Desciption != "" && newService.Price != 0 && newService.Duration != 0 {
		return ctx.Send(enterServiceParams+newService.Info()+readyToCreateService, markupReadyToCreateService, tele.ModeMarkdown)
	}
	return ctx.Send(enterServiceParams+newService.Info(), markupEnterServiceParams, tele.ModeMarkdown)
}
