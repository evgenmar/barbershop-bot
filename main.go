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

const sqliteStoragePath = "data/sqlite/storage.db"

var (
	location      *time.Location
	scheduledDays uint16 = 183
)

func init() {
	location = time.FixedZone("MSK", 3*60*60)
}

func main() {
	s := createStorage()
	defer s.Close()
	err := s.Init(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	barberIDs := getBarberIDs(s)

	makeBarbersScedules(s, barberIDs)

	//Triggers at 03:00 AM every Monday
	c := scheduleMakingBarbersScedules(s, barberIDs, "0 3 * * 1")
	defer c.Stop()

	b := createBot()

	barbers := b.Group()
	barbers.Use(middleware.Whitelist(barberIDs...))

	//TODO barberHandlers

	barbers.Handle("/barber", func(c tele.Context) error {
		return c.Send("hello barber")
	})

	users := b.Group()
	users.Use(NotInWhitelist(barberIDs...))

	//TODO userHandlers

	users.Handle("/user", func(c tele.Context) error {
		return c.Send("hello user")
	})

	// TODO sameCommandHandlers

	b.Handle("/start", func(c tele.Context) error { return nil }, middleware.Restrict(middleware.RestrictConfig{
		Chats: barberIDs,
		In:    func(c tele.Context) error { return c.Send("wellcome barber") },
		Out:   func(c tele.Context) error { return c.Send("wellcome user") },
	}))

	b.Start()
}

func scheduleMakingBarbersScedules(s storage.Storage, barberIDs []int64, specOfCron string) *cron.Cron {
	c := cron.New(cron.WithLocation(location))
	c.AddFunc(specOfCron,
		func() {
			makeBarbersScedules(s, barberIDs)
		})
	c.Start()
	return c
}

func getBarberIDs(s storage.Storage) []int64 {
	barberIDs, err := s.BarberIDs(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	return barberIDs
}

func createStorage() storage.Storage {
	var s storage.Storage
	mutex := sync.Mutex{}
	s, err := sqlite.New(sqliteStoragePath, location, &mutex)
	if err != nil {
		log.Fatal(err)
	}
	return s
}

func makeBarbersScedules(s storage.Storage, barberIDs []int64) {
	err := makeSchedules(&s, barberIDs, scheduledDays)
	if err != nil {
		log.Fatal(err)
	}
}

func createBot() *tele.Bot {
	pref := tele.Settings{
		Token: os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second,
			AllowedUpdates: []string{"message", "callback_query"}},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	b.Use(middleware.Recover())
	return b
}
