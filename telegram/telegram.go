package telegram

import (
	"barbershop-bot/storage"
	"log"
	"os"
	"time"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

type state uint8

const (
	stateStart state = iota
	stateUpdName
)

// botWithMiddleware creates bot with Recover(), AutoRespond() and withStorage(rep) global middleware.
func BotWithMiddleware(rep storage.Storage) *tele.Bot {
	pref := tele.Settings{
		Token: os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second,
			AllowedUpdates: []string{"message", "callback_query"}},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	bot.Use(middleware.Recover())
	bot.Use(middleware.AutoRespond())
	bot.Use(withStorage(rep))
	return bot
}

func SetHandlers(bot *tele.Bot, barberIDs []int64) *tele.Bot {
	barbers := bot.Group()
	barbers.Use(middleware.Whitelist(barberIDs...))
	users := bot.Group()
	users.Use(notInWhitelist(barberIDs...))

	bot.Handle("/start", noAction, middleware.Restrict(middleware.RestrictConfig{
		Chats: barberIDs,
		In:    onStartBarber,
		Out:   onStartUser,
	}))
	// TODO sameCommandHandlers

	barbers.Handle(&btnUpdPersonalBarber, onUpdPersonalBarber)

	barbers.Handle(&btnUpdNameBarber, onUpdNameBarber)
	barbers.Handle(&btnUpdPhoneBarber, onUpdPhoneBarber)
	barbers.Handle(&btnBackToMainBarber, onBackToMainBarber)
	//TODO barberHandlers

	users.Handle("/user", onUser)
	//TODO userHandlers

	return bot
}

func noAction(tele.Context) error { return nil }

// store fetches storage.Storage from tele.Context
func store(ctx tele.Context) storage.Storage {
	rep, ok := ctx.Get("storage").(storage.Storage)
	if !ok {
		log.Print("can't get storage from Context")
		return nil
	}
	return rep
}

func newStatus(state state) storage.Status {
	var expiration string
	if state == stateStart {
		expiration = "3000-01-01 00:00:00"
	} else {
		expiration = time.Now().In(time.FixedZone("UTC", 0)).Add(24 * time.Hour).Format(time.DateTime)
	}
	return storage.Status{
		State:      uint8(state),
		Expiration: expiration,
	}
}
