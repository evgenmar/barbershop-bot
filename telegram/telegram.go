package telegram

import (
	"barbershop-bot/lib/e"
	"barbershop-bot/storage"
	"database/sql"
	"errors"
	"log"
	"os"
	"regexp"
	"sync"
	"time"
	"unicode"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

type state byte

const (
	stateStart state = iota
	stateUpdName
	stateUpdPhone
	stateAddBarber
)

type protectedIDs struct {
	ids     []int64
	rwMutex sync.RWMutex
}

var barberIDs protectedIDs

func SetBarberIDs(IDs ...int64) {
	barberIDs.ids = IDs
}

func (p *protectedIDs) iDs() []int64 {
	p.rwMutex.RLock()
	defer p.rwMutex.RUnlock()
	return p.ids
}

func (p *protectedIDs) setIDs(ids []int64) {
	p.rwMutex.Lock()
	defer p.rwMutex.Unlock()
	p.ids = ids
}

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

func SetHandlers(bot *tele.Bot) *tele.Bot {
	barbers := bot.Group()
	barbers.Use(whitelist())
	users := bot.Group()
	users.Use(notInWhitelist())

	bot.Handle("/start", noAction, onStartRestrict())
	bot.Handle(tele.OnText, noAction, onTextRestrict())
	// TODO sameCommandHandlers

	barbers.Handle(&btnSettingsBarber, onSettingsBarber)

	barbers.Handle(&btnUpdPersonalBarber, onUpdPersonalBarber)
	barbers.Handle(&btnUpdNameBarber, onUpdNameBarber)
	barbers.Handle(&btnUpdPhoneBarber, onUpdPhoneBarber)

	barbers.Handle(&btnManageBarbers, onManageBarbers)
	barbers.Handle(&btnAddBarber, onAddBarber)
	barbers.Handle(&btnDeleteBarber, onDeleteBarber)

	barbers.Handle(&btnBackToMainBarber, onBackToMainBarber)

	barbers.Handle(tele.OnContact, onContactBarber)
	//TODO barberHandlers

	users.Handle("/user", onUser)
	//TODO userHandlers

	return bot
}

func noAction(tele.Context) error { return nil }

// store fetches storage.Storage from tele.Context
func getRepository(ctx tele.Context) (storage.Storage, error) {
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
		State:      sql.NullByte{Byte: byte(state), Valid: true},
		Expiration: sql.NullString{String: expiration, Valid: true},
	}
}

// getState returns state. If the state has not expired yet, the second returned value is true.
// If the state has already expired, the second returned value is false.
func getState(status storage.Status) (state, bool, error) {
	if !status.State.Valid {
		return stateStart, true, nil
	}
	expiration, err := time.Parse(time.DateTime, status.Expiration.String)
	if err != nil {
		return stateStart, false, e.Wrap("can't parse state expiration time", err)
	}
	if expiration.After(time.Now().In(time.FixedZone("UTC", 0))) {
		return state(status.State.Byte), true, nil
	}
	return state(status.State.Byte), false, nil
}

func isValidName(text string) bool {
	namePattern := `^[a-zA-Zа-яА-Я0-9\s]{2,20}$`
	regex := regexp.MustCompile(namePattern)
	var hasLetter bool
	for _, r := range text {
		if unicode.IsLetter(r) {
			hasLetter = true
			break
		}
	}
	return regex.MatchString(text) && hasLetter
}

func isValidPhone(text string) bool {
	phonePattern := `^(\+?\d?[\s-]?)?\(?\d{3}\)?[\s-]?\d{3}[\s-]?\d{2}[\s-]?\d{2}$`
	regex := regexp.MustCompile(phonePattern)
	return regex.MatchString(text)
}

func normalizePhone(phone string) (normalized string) {
	for _, r := range phone {
		if unicode.IsDigit(r) {
			normalized = normalized + string(r)
		}
	}
	if len(normalized) == 11 {
		normalized = normalized[1:]
	}
	normalized = "+7" + normalized
	return normalized
}
