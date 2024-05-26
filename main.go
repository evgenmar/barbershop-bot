package main

import (
	"barbershop-bot/config"
	"barbershop-bot/scheduler"
	"barbershop-bot/storage"
	"barbershop-bot/storage/sqlite"
	"barbershop-bot/telegram"
	"context"
	"log"
	"os"
	"strconv"
	"sync"
)

func main() {
	rep := createRepository()
	defer rep.Close()

	barberIDs := getBarberIDs(rep)
	makeBarbersSchedules(rep, barberIDs)

	bot := telegram.BotWithMiddleware(rep)
	bot = telegram.SetHandlers(bot, barberIDs)

	crn := scheduler.CronWithSettings(rep, barberIDs)
	crn.Start()
	defer crn.Stop()

	bot.Start()
	defer bot.Stop()
}

// createRepository creates repository and prepares it for use
func createRepository() storage.Storage {
	mutex := sync.Mutex{}
	rep, err := sqlite.New(config.SqliteStoragePath, &mutex)
	if err != nil {
		log.Fatal(err)
	}
	err = rep.Init(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	return rep
}

func getBarberIDs(rep storage.Storage) []int64 {
	barberIDs, err := rep.FindAllBarberIDs(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	if len(barberIDs) == 0 {
		barberID, err := strconv.ParseInt(os.Getenv("BarberID"), 10, 64)
		if err != nil {
			log.Fatal("can't get barberID from environment variable", err)
		}
		err = rep.CreateBarberID(context.TODO(), barberID)
		if err != nil {
			log.Fatal(err)
		}
		barberIDs = append(barberIDs, barberID)
	}
	return barberIDs
}

func makeBarbersSchedules(rep storage.Storage, barberIDs []int64) {
	err := scheduler.MakeSchedules(rep, barberIDs, config.ScheduledDays)
	if err != nil {
		log.Fatal(err)
	}
}
