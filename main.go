package main

import (
	"barbershop-bot/storage"
	"barbershop-bot/storage/sqlite"
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

const (
	sqliteStoragePath = "data/sqlite/storage.db"

	//scheduledDays is the number of days for which the barbershop schedule is compiled.
	scheduledDays uint16 = 183
)

// location is the time zone where the barbershop is located.
var location *time.Location

func init() {
	location = time.FixedZone("MSK", 3*60*60)
}

func main() {
	rep := createRepository()
	defer rep.Close()

	barberIDs := getBarberIDs(rep)

	makeBarbersScedules(rep, barberIDs)

	crn := scheduleEvents(rep, barberIDs)
	defer crn.Stop()

	bot := botWithMiddleware(rep)

	bot = setHandlers(bot, barberIDs)

	bot.Start()
	defer bot.Stop()
}

// createRepository creates repository and prepares it for use
func createRepository() (rep storage.Storage) {
	mutex := sync.Mutex{}
	rep, err := sqlite.New(sqliteStoragePath, location, &mutex)
	if err != nil {
		log.Fatal(err)
	}
	err = rep.Init(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	return
}

func getBarberIDs(rep storage.Storage) []int64 {
	barberIDs, err := rep.BarberIDs(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	return barberIDs
}

func makeBarbersScedules(rep storage.Storage, barberIDs []int64) {
	err := makeSchedules(rep, barberIDs, scheduledDays)
	if err != nil {
		log.Fatal(err)
	}
}

// scheduleEvents triggers events on a schedule.
// Triggered events:
//   - making schedules for barbers - every Monday at 03:00 AM
func scheduleEvents(rep storage.Storage, barberIDs []int64) *cron.Cron {
	crn := cron.New(cron.WithLocation(location))
	crn.AddFunc("0 3 * * 1",
		func() {
			err := makeSchedules(rep, barberIDs, scheduledDays)
			if err != nil {
				log.Print(err)
			}
		})
	crn.Start()
	return crn
}

// botWithMiddleware creates bot with Recover(), AutoRespond() and withStorage(rep) middleware.
func botWithMiddleware(rep storage.Storage) *tele.Bot {
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

func setHandlers(bot *tele.Bot, barberIDs []int64) *tele.Bot {
	barbers := bot.Group()
	barbers.Use(middleware.Whitelist(barberIDs...))

	//TODO barberHandlers

	users := bot.Group()
	users.Use(notInWhitelist(barberIDs...))

	//TODO userHandlers

	users.Handle("/user", onUser)

	// TODO sameCommandHandlers

	bot.Handle("/start", noAction, middleware.Restrict(middleware.RestrictConfig{
		Chats: barberIDs,
		In:    onStartBarber,
		Out:   onStartUser, //TODO
	}))

	bot.Handle(&btnBackToMain, noAction, middleware.Restrict(middleware.RestrictConfig{
		Chats: barberIDs,
		In:    onBackToMainBarber,
		Out:   onStartUser, //TODO
	}))

	bot.Handle(&btnUpdName, noAction, middleware.Restrict(middleware.RestrictConfig{
		Chats: barberIDs,
		In:    onUpdNameBarber,
		Out:   onStartUser, //TODO
	}))

	bot.Handle(&btnUpdPersonal, onUpdPersonal)
	return bot
}
