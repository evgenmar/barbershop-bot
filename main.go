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
	var s storage.Storage
	mutex := sync.Mutex{}
	s, err := sqlite.New(sqliteStoragePath, location, &mutex)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	err = s.Init(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	barberIDs, err := s.BarberIDs(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	err = makeSchedules(&s, barberIDs, scheduledDays)
	if err != nil {
		log.Fatal(err)
	}

	c := cron.New(cron.WithLocation(location))
	c.AddFunc("0 3 * * 1", //Triggers at 03:00 AM every Monday0 3 * * 1
		func() {
			err = makeSchedules(&s, barberIDs, scheduledDays)
			if err != nil {
				log.Fatal(err)
			}
		})
	c.Start()
	defer c.Stop()

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
