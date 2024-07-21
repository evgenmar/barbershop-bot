package telegram

import (
	"log"
	"os"
	"sync"
	"time"

	tele "gopkg.in/telebot.v3"
	mw "gopkg.in/telebot.v3/middleware"
)

var appointmentsMutex sync.Mutex

var Bot *tele.Bot

func InitBot() {
	Bot = setHandlers(setMiddleware(newBot()))
}

// botWithMiddleware creates bot with Recover(), AutoRespond() and withStorage(rep) global middleware.
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

func setMiddleware(bot *tele.Bot) *tele.Bot {
	bot.Use(mw.Recover())
	bot.Use(mw.AutoRespond())
	return bot
}

func setHandlers(bot *tele.Bot) *tele.Bot {
	barbers := bot.Group()
	barbers.Use(whitelist())
	users := bot.Group()
	users.Use(notInWhitelist())

	bot.Handle(&btnEmpty, noAction)

	bot.Handle("/start", noAction, onStartRestrict())
	bot.Handle(tele.OnText, noAction, onTextRestrict())

	barbers.Handle(&btnBarberSettings, onBarberSettings)

	barbers.Handle(&btnListOfNecessarySettings, onListOfNecessarySettings)

	barbers.Handle(&btnBarberManageAccount, onBarberManageAccount)

	barbers.Handle(&btnBarberCurrentSettings, onBarberCurrentSettings)

	barbers.Handle(&btnBarberUpdPersonal, onBarberUpdPersonal)
	barbers.Handle(&btnBarberAgreeWithPrivacy, onBarberAgreeWithPrivacy)
	barbers.Handle(&btnBarberUpdName, onBarberUpdName)
	barbers.Handle(&btnBarberUpdPhone, onBarberUpdPhone)

	barbers.Handle(&btnDeleteAccount, onDeleteAccount)
	barbers.Handle(&btnSetLastWorkDate, onSetLastWorkDate)
	barbers.Handle(&btnSelectLastWorkDate, onSelectLastWorkDate)
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
	barbers.Handle(&btnSelectCertainDurationOnEnter, onSelectCertainDurationOnEnter)
	barbers.Handle(&btnSaveNewService, onSaveNewService)

	barbers.Handle(&btnEditService, onEditService)
	barbers.Handle(&btnСontinueEditingService, onСontinueEditingService)
	barbers.Handle(&btnSelectServiceToEdit, onSelectServiceToEdit)
	barbers.Handle(&btnSelectCertainServiceToEdit, onSelectCertainServiceToEdit)
	barbers.Handle(&btnEditServiceName, onEditServiceName)
	barbers.Handle(&btnEditServiceDescription, onEditServiceDescription)
	barbers.Handle(&btnEditServicePrice, onEditServicePrice)
	barbers.Handle(&btnSelectServiceDurationOnEdit, onSelectServiceDurationOnEdit)
	barbers.Handle(&btnSelectCertainDurationOnEdit, onSelectCertainDurationOnEdit)
	barbers.Handle(&btnUpdateService, onUpdateService)

	barbers.Handle(&btnDeleteService, onDeleteService)
	barbers.Handle(&btnSelectCertainServiceToDelete, onSelectCertainServiceToDelete)
	barbers.Handle(&btnSureToDeleteService, onSureToDeleteService)

	barbers.Handle(&btnManageBarbers, onManageBarbers)
	barbers.Handle(&btnShowAllBurbers, onShowAllBarbers)
	barbers.Handle(&btnAddBarber, onAddBarber)
	barbers.Handle(&btnDeleteBarber, onDeleteBarber)
	barbers.Handle(&btnDeleteCertainBarber, onDeleteCertainBarber)

	barbers.Handle(&btnBarberBackToMain, onBarberBackToMain)

	barbers.Handle(tele.OnContact, onContactBarber)
	//TODO barberHandlers

	users.Handle(&btnSignUpForAppointment, onSignUpForAppointment)
	users.Handle(&btnSelectBarberForAppointment, onSelectBarberForAppointment)
	users.Handle(&btnUserSelectServiceForAppointment, onUserSelectServiceForAppointment)
	users.Handle(&btnUserSelectMonthForAppointment, onUserSelectMonthForAppointment)
	users.Handle(&btnUserSelectWorkdayForAppointment, onUserSelectWorkdayForAppointment)
	users.Handle(&btnUserSelectTimeForAppointment, onUserSelectTimeForAppointment)
	users.Handle(&btnUserConfirmNewAppointment, onUserConfirmNewAppointment)
	users.Handle(&btnUserSelectAnotherTimeForAppointment, onUserSelectAnotherTimeForAppointment)

	users.Handle(&btnRescheduleOrCancelAppointment, onRescheduleOrCancelAppointment)
	users.Handle(&btnUserRescheduleAppointment, onUserRescheduleAppointment)
	users.Handle(&btnUserConfirmRescheduleAppointment, onUserConfirmRescheduleAppointment)

	users.Handle(&btnUserSettings, onUserSettings)
	users.Handle(&btnUserUpdPersonal, onUserUpdPersonal)
	users.Handle(&btnUserPrivacy, onUserPrivacy)
	users.Handle(&btnUserAgreeWithPrivacy, onUserAgreeWithPrivacy)
	users.Handle(&btnUserUpdName, onUserUpdName)
	users.Handle(&btnUserUpdPhone, onUserUpdPhone)

	users.Handle(&btnUserBackToMain, onUserBackToMain)
	users.Handle(&btnUserBackToMainSend, onUserBackToMainSend)
	//TODO userHandlers

	return bot
}
