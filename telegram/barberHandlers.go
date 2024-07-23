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
	return ctx.Send(mainMenu, markupBarberMain)
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
		return logAndMsgErrBarber(ctx, "can't show to barber services for new appointment", err)
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
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	service, err := cp.RepoWithContext.GetServiceByID(serviceID)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	newAppointment := sess.GetAppointmentBarber(ctx.Sender().ID)
	newAppointment.ServiceID = service.ID
	newAppointment.Duration = service.Duration
	return calculateAndShowToBarberFreeWorkdaysForAppointment(ctx, 0, newAppointment)
}

func onBarberSelectMonthForAppointment(ctx tele.Context) error {
	delta, err := strconv.ParseInt(ctx.Callback().Data, 10, 8)
	if err != nil {
		return logAndMsgErrBarber(ctx, "can't show free workdays for appointment", err)
	}
	appointment := sess.GetAppointmentBarber(ctx.Sender().ID)
	return calculateAndShowToBarberFreeWorkdaysForAppointment(ctx, int8(delta), appointment)
}

func onBarberSelectWorkdayForAppointment(ctx tele.Context) error {
	errMsg := "can't show free time for appointment"
	workdayID, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	workday, appointments, err := getWorkdayAndAppointments(workdayID)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	appointment := sess.GetAppointmentBarber(ctx.Sender().ID)
	appointment.WorkdayID = workdayID
	return calculateAndShowToBarberFreeTimesForAppointment(ctx, workday, appointments, appointment)
}

func onBarberSelectTimeForAppointment(ctx tele.Context) error {
	errMsg := "can't handle select time for new appointment"
	apptTime, err := strconv.ParseUint(ctx.Callback().Data, 10, 64)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	appointmentTime := tm.Duration(apptTime)
	barberID := ctx.Sender().ID
	appointment := sess.GetAppointmentBarber(barberID)
	appointment.Time = appointmentTime
	workday, appointments, err := getWorkdayAndAppointments(appointment.WorkdayID)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	if !isTimeForAppointmentAvailable(workday, appointments, appointment) {
		return calculateAndShowToBarberFreeTimesForAppointment(ctx, workday, appointments, appointment)
	}
	sess.UpdateAppointmentAndBarberState(barberID, appointment, sess.StateStart)
	if appointment.ID == 0 {
		service, err := cp.RepoWithContext.GetServiceByID(appointment.ServiceID)
		if err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		return ctx.Edit(
			fmt.Sprintf(confirmNewAppointment, service.Info(), tm.ShowDate(workday.Date), appointmentTime.ShortString()),
			markupBarberConfirmNewAppointment,
		)
	}
	serviceInfo := getNullServiceInfo(appointment.ServiceID, appointment.Duration)
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
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	if ok {
		serviceInfo := getNullServiceInfo(appointment.ServiceID, appointment.Duration)
		workday, err := cp.RepoWithContext.GetWorkdayByID(appointment.WorkdayID)
		if err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
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
		return logAndMsgErrBarber(ctx, "can't handle select another time for appointment", err)
	}
	return calculateAndShowToBarberFreeTimesForAppointment(ctx, workday, appointments, appointment)
}

func onUpdNote(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateUpdNote)
	return ctx.Edit(enterNote)
}

func onBarberSettings(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(settingsMenu, markupBarberSettings)
}

func onListOfNecessarySettings(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(listOfNecessarySettings, markupShortBarberSettings)
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
		return logAndMsgErrBarber(ctx, "can't show current barber settings", err)
	}
	return ctx.Edit(currentSettings+barber.Info(), markupBarberBackToMain)
}

func onBarberUpdPersonal(ctx tele.Context) error {
	sess.UpdateBarberState(ctx.Sender().ID, sess.StateStart)
	barberID := ctx.Sender().ID
	barber, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return logAndMsgErrBarber(ctx, "can't show update personal data options for barber", err)
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
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	delta, err := strconv.ParseInt(ctx.Callback().Data, 10, 8)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	firstDisplayedDateRange := defineFirstDisplayedDateRangeForLastWorkDate(latestAppointmentDate)
	displayedMonthRange := ent.DateRange{
		StartDate: firstDisplayedDateRange.EndDate,
		EndDate:   ent.MonthFromNow(cfg.ScheduledWeeks * 7 / 30).EndDate,
	}
	lastWorkDate := sess.GetLastWorkDate(barberID)
	displayedDateRange := defineDisplayedDateRangeForLastWorkDate(
		firstDisplayedDateRange,
		displayedMonthRange,
		int8(delta),
		lastWorkDate,
	)
	lastWorkDate.LastShownDate = displayedDateRange.EndDate
	sess.UpdateLastWorkDateAndState(barberID, lastWorkDate, sess.StateStart)
	return ctx.Edit(selectLastWorkDate, markupSelectLastWorkDate(displayedDateRange, displayedMonthRange))
}

func onSelectLastWorkDate(ctx tele.Context) error {
	errMsg := "can't save last work date"
	barberID := ctx.Sender().ID
	sess.UpdateLastWorkDateAndState(barberID, sess.LastWorkDate{}, sess.StateStart)
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
		return ctx.Edit(lastWorkDateUnchanged, markupBarberBackToMain)
	case 1:
		if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: barberID, LastWorkdate: dateToSave}); err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		if err := sched.MakeSchedule(barberID); err != nil {
			log.Print(e.Wrap(errMsg, err))
			// TODO: ensure atomicity using outbox pattern
			return ctx.Send(lastWorkDateSavedWithoutSсhedule, markupBarberBackToMain)
		}
		return ctx.Edit(lastWorkDateSaved, markupBarberBackToMain)
	case -1:
		latestWorkDate, err := cp.RepoWithContext.GetLatestWorkDate(barberID)
		if err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		dateRangeToDelete := ent.DateRange{StartDate: dateToSave.Add(24 * time.Hour), EndDate: latestWorkDate}
		err = cp.RepoWithContext.DeleteWorkdaysByDateRange(barberID, dateRangeToDelete)
		if err != nil && !errors.Is(err, rep.ErrInvalidDateRange) {
			if errors.Is(err, rep.ErrAppointmentsExists) {
				return ctx.Edit(haveAppointmentAfterDataToSave, markupBarberBackToMain)
			}
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: barberID, LastWorkdate: dateToSave}); err != nil {
			// TODO: ensure atomicity using outbox pattern
			return ctx.Send(lastWorkDateNotSavedButScheduleDeleted, markupBarberBackToMain)
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
		return logAndMsgErrBarber(ctx, "can't provide options for barber self deletion", err)
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
	return ctx.Edit(
		selectServiceToEdit,
		markupSelectService(services, endpntServiceToEdit, endpntBarberBackToMain),
	)
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
	return ctx.Edit(
		selectServiceToDelete,
		markupSelectService(services, endpntServiceToDelete, endpntBarberBackToMain),
	)
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
		return logAndMsgErrBarber(ctx, "can't show all barbers", err)
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
		return logAndMsgErrBarber(ctx, "can't suggest actions to delete barber", err)
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
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	barberToDelete, err := cp.RepoWithContext.GetBarberByID(barberIDToDelete)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	if barberToDelete.LastWorkdate.Before(tm.Today()) {
		if err := cp.RepoWithContext.DeletePastAppointments(barberIDToDelete); err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		if err := cp.RepoWithContext.DeleteBarberByID(barberIDToDelete); err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
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
		return logAndMsgErrBarber(ctx, "can't update barber's name", err)
	}
	sess.UpdateBarberState(barberID, sess.StateStart)
	return ctx.Send(updNameSuccess, markupBarberPersonal)
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
		return logAndMsgErrBarber(ctx, "can't update barber's phone", err)
	}
	sess.UpdateBarberState(barberID, sess.StateStart)
	return ctx.Send(updPhoneSuccess, markupBarberPersonal)
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

func onUpdateNote(ctx tele.Context) error {
	errMsg := "can't update note"
	barberID := ctx.Sender().ID
	appointment := sess.GetAppointmentBarber(barberID)
	appointmentID, err := cp.RepoWithContext.GetAppointmentIDByWorkdayIDAndTime(appointment.WorkdayID, appointment.Time)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	if err := cp.RepoWithContext.UpdateAppointment(ent.Appointment{ID: appointmentID, Note: ctx.Message().Text}); err != nil {
		if errors.Is(err, rep.ErrInvalidAppointment) {
			log.Print(e.Wrap("invalid note", err))
			return ctx.Send(invalidNote)
		}
		sess.UpdateBarberState(barberID, sess.StateStart)
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	sess.UpdateBarberState(barberID, sess.StateStart)
	return ctx.Send(updNoteSuccess, markupBarberBackToMain)
}

func onContactBarber(ctx tele.Context) error {
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
			return ctx.Send(userIsAlreadyBarber, markupBarberBackToMain)
		}
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	cfg.Barbers.AddID(newBarberID)
	if err := sched.MakeSchedule(newBarberID); err != nil {
		log.Print(e.Wrap(errMsg, err))
		// TODO: ensure atomicity using outbox pattern
		return ctx.Send(addedNewBarberWithoutSсhedule, markupBarberBackToMain)
	}
	return ctx.Send(addedNewBarberWithSсhedule, markupBarberBackToMain)
}

func calculateAndShowToBarberFreeTimesForAppointment(ctx tele.Context, workday ent.Workday, appointments []ent.Appointment, appointment sess.Appointment) error {
	freeTimes := freeTimesForAppointment(workday, appointments, appointment)
	if len(freeTimes) == 0 {
		return calculateAndShowToBarberFreeWorkdaysForAppointment(ctx, 0, appointment)
	}
	sess.UpdateAppointmentAndBarberState(ctx.Sender().ID, appointment, sess.StateStart)
	serviceInfo := getNullServiceInfo(appointment.ServiceID, appointment.Duration)
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
	displayedDateRange, displayedMonthRange, ok, err := calculateAndCheckDisplayedRanges(deltaDisplayedMonth, appointment)
	if err != nil {
		return logAndMsgErrBarber(ctx, "can't show to barber free workdays for new appointment", err)
	}
	if !ok {
		return ctx.Edit(informBarberNoFreeTimeForAppointment, markupBarberBackToMain)
	}
	appointment.LastShownDate = displayedDateRange.EndDate
	sess.UpdateAppointmentAndBarberState(ctx.Sender().ID, appointment, sess.StateStart)
	return showToBarberFreeWorkdaysForAppointment(ctx, displayedDateRange, displayedMonthRange, appointment)
}

func checkAndCreateAppointmentByBarber(appointment sess.Appointment) (ok bool, err error) {
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
	err = cp.RepoWithContext.CreateAppointment(ent.Appointment{
		WorkdayID: appointment.WorkdayID,
		ServiceID: appointment.ServiceID,
		Time:      appointment.Time,
		Duration:  appointment.Duration,
		CreatedAt: time.Now().Unix(),
	})
	return
}

func defineDisplayedDateRangeForLastWorkDate(
	firstDisplayedDateRange ent.DateRange,
	displayedMonthRange ent.DateRange,
	deltaDisplayedMonth int8,
	lastWorkDate sess.LastWorkDate,
) ent.DateRange {
	displayedDateRange := firstDisplayedDateRange
	if deltaDisplayedMonth != 0 {
		newDateRange := ent.MonthFromBase(lastWorkDate.LastShownDate, deltaDisplayedMonth)
		if newDateRange.EndDate.After(displayedMonthRange.EndDate) || newDateRange.EndDate.Before(displayedMonthRange.StartDate) {
			if lastWorkDate.LastShownDate.After(displayedDateRange.EndDate) {
				displayedDateRange = ent.MonthFromBase(lastWorkDate.LastShownDate, 0)
			}
		}
		if newDateRange.EndDate.After(displayedDateRange.EndDate) {
			displayedDateRange = newDateRange
		}
	}
	return displayedDateRange
}

func defineFirstDisplayedDateRangeForLastWorkDate(latestAppointmentDate time.Time) (firstDisplayedDateRange ent.DateRange) {
	var relativeFirstDisplayedMonth byte = 0
	for relativeFirstDisplayedMonth <= cfg.MaxAppointmentBookingMonths {
		firstDisplayedDateRange = ent.MonthFromNow(relativeFirstDisplayedMonth)
		if latestAppointmentDate.Compare(firstDisplayedDateRange.EndDate) <= 0 {
			if latestAppointmentDate.After(firstDisplayedDateRange.StartDate) {
				firstDisplayedDateRange.StartDate = latestAppointmentDate
			}
			break
		}
		relativeFirstDisplayedMonth++
	}
	return
}

func logAndMsgErrBarber(ctx tele.Context, msg string, err error) error {
	log.Print(e.Wrap(msg, err))
	return ctx.Send(errorBarber, markupBarberBackToMain)
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

func showToBarberFreeWorkdaysForAppointment(ctx tele.Context, displayedDateRange, displayedMonthRange ent.DateRange, appointment sess.Appointment) error {
	markupSelectWorkday, err := markupSelectWorkdayForAppointment(
		displayedDateRange,
		displayedMonthRange,
		appointment,
		endpntBarberSelectWorkdayForAppointment,
		endpntBarberSelectMonthForAppointment,
		endpntBarberBackToMain,
	)
	if err != nil {
		return logAndMsgErrBarber(ctx, "can't show to barber free workdays for appointment", err)
	}
	serviceInfo := getNullServiceInfo(appointment.ServiceID, appointment.Duration)
	return ctx.Edit(fmt.Sprintf(selectDateForAppointment, serviceInfo), markupSelectWorkday)
}
