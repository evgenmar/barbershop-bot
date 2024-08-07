package telegram

import (
	"log"
	"os"
	"sync"
	"time"

	tele "gopkg.in/telebot.v3"
	mw "gopkg.in/telebot.v3/middleware"
)

var (
	mutex sync.Mutex
)

var Bot *tele.Bot

func InitBot() {
	Bot = withHandlers(withMiddleware(newBot()))
}

func newBot() *tele.Bot {
	pref := tele.Settings{
		Token: os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second,
			AllowedUpdates: []string{"message", "callback_query"}},
	}
	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}
	return bot
}

func withMiddleware(bot *tele.Bot) *tele.Bot {
	bot.Use(mw.Recover())
	bot.Use(mw.AutoRespond())
	return bot
}

func withHandlers(bot *tele.Bot) *tele.Bot {
	barbers := bot.Group()
	barbers.Use(whitelist())
	users := bot.Group()
	users.Use(notInWhitelist())

	bot.Handle(&btnEmpty, noAction)

	bot.Handle("/start", noAction, onStartRestrict())
	bot.Handle(tele.OnText, noAction, onTextRestrict())

	barbers.Handle(&btnSignUpClientForAppointment, onSignUpClientForAppointment)
	barbers.Handle(callbackUnique(endpntBarberSelectServiceForAppointment), onBarberSelectServiceForAppointment)
	barbers.Handle(callbackUnique(endpntBarberSelectMonthForAppointment), onBarberSelectMonthForAppointment)
	barbers.Handle(callbackUnique(endpntBarberSelectWorkdayForAppointment), onBarberSelectWorkdayForAppointment)
	barbers.Handle(callbackUnique(endpntBarberSelectTimeForAppointment), onBarberSelectTimeForAppointment)
	barbers.Handle(&btnBarberConfirmNewAppointment, onBarberConfirmNewAppointment)
	barbers.Handle(&btnBarberSelectAnotherTimeForAppointment, onBarberSelectAnotherTimeForAppointment)
	barbers.Handle(&btnAddNote, onAddNote)

	barbers.Handle(&btnAddWorkday, onAddWorkday)
	barbers.Handle(&btnAddNonWorkday, onAddNonWorkday)
	barbers.Handle(callbackUnique(endpntCreateWorkday), onCreateWorkday)
	barbers.Handle(callbackUnique(endpntDeleteWorkday), onDeleteWorkday)

	barbers.Handle(callbackUnique(endpntMyWorkSchedule), onMyWorkSchedule)
	barbers.Handle(callbackUnique(endpntSelectMonthFromScheduleCalendar), onSelectMonthFromScheduleCalendar)
	barbers.Handle(callbackUnique(endpntSelectWorkdayFromScheduleCalendar), onSelectWorkdayFromScheduleCalendar)

	barbers.Handle(&btnMakeThisDayNonWorking, onMakeThisDayNonWorking)
	barbers.Handle(&btnUpdWorkdayStartTime, onUpdWorkdayStartTime)
	barbers.Handle(callbackUnique(endpntSelectWorkdayStartTime), onSelectWorkdayStartTime)
	barbers.Handle(&btnUpdWorkdayEndTime, onUpdWorkdayEndTime)
	barbers.Handle(callbackUnique(endpntSelectWorkdayEndTime), onSelectWorkdayEndTime)
	barbers.Handle(callbackUnique(endpntSelectAppointment), onSelectAppointment)

	barbers.Handle(callbackUnique(endpntBarberRescheduleAppointment), onBarberRescheduleAppointment)
	barbers.Handle(&btnBarberConfirmRescheduleAppointment, onBarberConfirmRescheduleAppointment)
	barbers.Handle(callbackUnique(endpntBarberCancelAppointment), onBarberCancelAppointment)
	barbers.Handle(&btnBarberConfirmCancelAppointment, onBarberConfirmCancelAppointment)
	barbers.Handle(&btnCancelAppointmentAndApology, onCancelAppointmentAndApology)
	barbers.Handle(&btnConfirmCancelAppointmentAndApology, onConfirmCancelAppointmentAndApology)
	barbers.Handle(&btnUpdNote, onUpdNote)

	barbers.Handle(&btnBarberSettings, onBarberSettings)

	barbers.Handle(&btnBarbersMemo, onBarbersMemo)

	barbers.Handle(&btnBarberManageAccount, onBarberManageAccount)

	barbers.Handle(&btnBarberCurrentSettings, onBarberCurrentSettings)

	barbers.Handle(&btnBarberUpdPersonal, onBarberUpdPersonal)
	barbers.Handle(&btnBarberAgreeWithPrivacy, onBarberAgreeWithPrivacy)
	barbers.Handle(&btnBarberUpdName, onBarberUpdName)
	barbers.Handle(&btnBarberUpdPhone, onBarberUpdPhone)

	barbers.Handle(&btnDeleteAccount, onDeleteAccount)
	barbers.Handle(&btnSetLastWorkDate, onSetLastWorkDate)
	barbers.Handle(callbackUnique(endpntSelectLastWorkDate), onSelectLastWorkDate)
	barbers.Handle(&btnSelfDeleteBarber, onSelfDeleteBarber)
	barbers.Handle(&btnSureToSelfDeleteBarber, onSureToSelfDeleteBarber)

	barbers.Handle(&btnManageServices, onManageServices)
	barbers.Handle(&btnShowMyServices, onShowMyServices)

	barbers.Handle(&btnCreateService, onCreateService)
	barbers.Handle(&btnСontinueOldService, onСontinueOldService)
	barbers.Handle(&btnMakeNewService, onMakeNewService)
	barbers.Handle(&btnEnterServiceName, onEnterServiceName)
	barbers.Handle(&btnEnterServiceDescription, onEnterServiceDescription)
	barbers.Handle(&btnEnterServicePrice, onEnterServicePrice)
	barbers.Handle(&btnSelectServiceDurationOnEnter, onSelectServiceDurationOnEnter)
	barbers.Handle(callbackUnique(endpntEnterServiceDuration), onSelectCertainDurationOnEnter)
	barbers.Handle(&btnSaveNewService, onSaveNewService)

	barbers.Handle(&btnEditService, onEditService)
	barbers.Handle(&btnСontinueEditingService, onСontinueEditingService)
	barbers.Handle(&btnSelectServiceToEdit, onSelectServiceToEdit)
	barbers.Handle(callbackUnique(endpntServiceToEdit), onSelectCertainServiceToEdit)
	barbers.Handle(&btnEditServiceName, onEditServiceName)
	barbers.Handle(&btnEditServiceDescription, onEditServiceDescription)
	barbers.Handle(&btnEditServicePrice, onEditServicePrice)
	barbers.Handle(&btnSelectServiceDurationOnEdit, onSelectServiceDurationOnEdit)
	barbers.Handle(callbackUnique(endpntEditServiceDuration), onSelectCertainDurationOnEdit)
	barbers.Handle(&btnUpdateService, onUpdateService)

	barbers.Handle(&btnDeleteService, onDeleteService)
	barbers.Handle(callbackUnique(endpntServiceToDelete), onSelectCertainServiceToDelete)
	barbers.Handle(callbackUnique(endpntSureToDeleteService), onSureToDeleteService)

	barbers.Handle(&btnManageBarbers, onManageBarbers)
	barbers.Handle(&btnShowAllBurbers, onShowAllBarbers)
	barbers.Handle(&btnAddBarber, onAddBarber)
	barbers.Handle(&btnDeleteBarber, onDeleteBarber)
	barbers.Handle(callbackUnique(endpntBarberToDeletion), onDeleteCertainBarber)

	barbers.Handle(&btnBarberBackToMain, onBarberBackToMain)

	barbers.Handle(tele.OnContact, onContactBarber)
	//TODO barberHandlers

	users.Handle(&btnSignUpForAppointment, onSignUpForAppointment)
	users.Handle(callbackUnique(endpntBarberForAppointment), onSelectBarberForAppointment)
	users.Handle(callbackUnique(endpntUserSelectServiceForAppointment), onUserSelectServiceForAppointment)
	users.Handle(callbackUnique(endpntUserSelectMonthForAppointment), onUserSelectMonthForAppointment)
	users.Handle(callbackUnique(endpntUserSelectWorkdayForAppointment), onUserSelectWorkdayForAppointment)
	users.Handle(callbackUnique(endpntUserSelectTimeForAppointment), onUserSelectTimeForAppointment)
	users.Handle(&btnUserConfirmNewAppointment, onUserConfirmNewAppointment)
	users.Handle(&btnUserSelectAnotherTimeForAppointment, onUserSelectAnotherTimeForAppointment)

	users.Handle(&btnRescheduleOrCancelAppointment, onRescheduleOrCancelAppointment)
	users.Handle(&btnUserRescheduleAppointment, onUserRescheduleAppointment)
	users.Handle(&btnUserConfirmRescheduleAppointment, onUserConfirmRescheduleAppointment)
	users.Handle(&btnUserCancelAppointment, onUserCancelAppointment)
	users.Handle(&btnUserConfirmCancelAppointment, onUserConfirmCancelAppointment)

	users.Handle(&btnUserSettings, onUserSettings)
	users.Handle(&btnUserUpdPersonal, onUserUpdPersonal)
	users.Handle(&btnUserPrivacy, onUserPrivacy)
	users.Handle(&btnUserAgreeWithPrivacy, onUserAgreeWithPrivacy)
	users.Handle(&btnUserUpdName, onUserUpdName)
	users.Handle(&btnUserUpdPhone, onUserUpdPhone)

	users.Handle(&btnUserBackToMain, onUserBackToMain)
	//TODO userHandlers

	return bot
}
