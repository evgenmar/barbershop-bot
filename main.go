package main

import (
	"barbershop-bot/lib/config"
	"barbershop-bot/lib/e"
	"barbershop-bot/repository/storage"
	"barbershop-bot/repository/storage/sqlite"
	"barbershop-bot/scheduler"
	"barbershop-bot/telegram"
	"context"
	"log"
	"os"
	"strconv"
)

func main() {
	storage, barberIDs := prepareStorage()

	telegram.SetBarberIDs(barberIDs...)
	bot := telegram.BotWithMiddleware(storage)
	bot = telegram.SetHandlers(bot)

	crn := scheduler.CronWithSettings(storage)
	crn.Start()

	bot.Start()
}

func prepareStorage() (storage.Storage, []int64) {
	storage := initStorage(createSQLite("data/sqlite/storage.db"))
	barberIDs := actualizeBarberIDs(storage)
	makeBarbersSchedules(storage)
	return storage, barberIDs
}

func createSQLite(path string) *sqlite.Storage {
	db, err := sqlite.New(path)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func initStorage(storage storage.Storage) storage.Storage {
	ctx, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutWrite)
	err := storage.Init(ctx)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	return storage
}

func actualizeBarberIDs(storage storage.Storage) []int64 {
	ctx, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutRead)
	barberIDs, err := storage.FindAllBarberIDs(ctx)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	if len(barberIDs) == 0 {
		barberID := getBarberIDFromEnv()
		createBarber(storage, barberID)
		barberIDs = append(barberIDs, barberID)
	}
	return barberIDs
}

func getBarberIDFromEnv() int64 {
	barberID, err := strconv.ParseInt(os.Getenv("BarberID"), 10, 64)
	if err != nil {
		log.Fatal(e.Wrap("can't get barberID from environment variable", err))
	}
	return barberID
}

func createBarber(storage storage.Storage, barberID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutWrite)
	err := storage.CreateBarber(ctx, barberID)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
}

func makeBarbersSchedules(storage storage.Storage) {
	err := scheduler.MakeSchedules(storage)
	if err != nil {
		log.Fatal(err)
	}
}
