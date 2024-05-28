package telegram

import (
	"barbershop-bot/lib/e"
	"barbershop-bot/storage"
	"errors"
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
	stateUpdPhone
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
	bot.Handle(tele.OnText, noAction, middleware.Restrict(middleware.RestrictConfig{
		Chats: barberIDs,
		In:    onTextBarber,
		Out:   onStartUser, // TODO
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
func store(ctx tele.Context) (storage.Storage, error) {
	rep, ok := ctx.Get("storage").(storage.Storage)
	if !ok {
		return nil, errors.New("can't get storage from tele.Context")
	}
	return rep, nil
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

// getState returns state. If the state has not expired yet, the second returned value is false.
// If the state has already expired, the second returned value is true.
func getState(status storage.Status) (state, bool, error) {
	expiration, err := time.Parse(time.DateTime, status.Expiration)
	if err != nil {
		return 0, false, e.Wrap("can't parse state expiration time", err)
	}
	if expiration.After(time.Now().In(time.FixedZone("UTC", 0))) {
		return state(status.State), false, nil
	}
	return state(status.State), true, nil
}

func getRepository(ctx tele.Context, errMsg string) storage.Storage {
	rep, err := store(ctx)
	if err != nil {
		log.Panicf("%s: %s", errMsg, err)
	}
	return rep
}
