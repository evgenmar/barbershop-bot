package telegram

import (
	"log"
	"os"
	"time"

	tele "gopkg.in/telebot.v3"
	mw "gopkg.in/telebot.v3/middleware"
)

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
	// TODO sameCommandHandlers

	barbers.Handle(&btnSettingsBarber, onSettingsBarber)

	barbers.Handle(&btnUpdPersonalBarber, onUpdPersonalBarber)
	barbers.Handle(&btnUpdNameBarber, onUpdNameBarber)
	barbers.Handle(&btnUpdPhoneBarber, onUpdPhoneBarber)

	barbers.Handle(&btnManageAccountBarber, onManageAccountBarber)
	barbers.Handle(&btnSetLastWorkDate, onSetLastWorkDate)
	barbers.Handle(&btnSelectLastWorkDate, onSelectLastWorkDate)

	barbers.Handle(&btnManageBarbers, onManageBarbers)
	barbers.Handle(&btnAddBarber, onAddBarber)
	barbers.Handle(&btnDeleteBarber, onDeleteBarber)
	barbers.Handle(&btnDeleteCertainBarber, onDeleteCertainBarber)

	barbers.Handle(&btnBackToMainBarber, onBackToMainBarber)

	barbers.Handle(tele.OnContact, onContactBarber)
	//TODO barberHandlers

	users.Handle("/user", onUser)
	//TODO userHandlers

	return bot
}
