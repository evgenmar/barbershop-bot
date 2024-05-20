package main

import (
	"barbershop-bot/storage"
	"barbershop-bot/storage/sqlite"
	"context"
	"log"
	"os"
	"time"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

const sqliteStoragePath = "data/sqlite/storage.db"

func main() {
	var s storage.Storage
	s, err := sqlite.New(sqliteStoragePath)
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
