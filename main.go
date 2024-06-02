package main

import (
	"barbershop-bot/config"
	"barbershop-bot/lib/e"
	"barbershop-bot/scheduler"
	"barbershop-bot/storage"
	"barbershop-bot/storage/sqlite"
	"barbershop-bot/telegram"
	"context"
	"log"
	"os"
	"strconv"
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
	bot.Stop()
}

// createRepository creates repository and prepares it for use
func createRepository() storage.Storage {
	rep, err := sqlite.New(config.SqliteStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutWrite)
	err = rep.Init(ctx)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	return rep
}

func getBarberIDs(rep storage.Storage) []int64 {
	ctx, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutRead)
	barberIDs, err := rep.FindAllBarberIDs(ctx)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	if len(barberIDs) == 0 {
		barberID, err := strconv.ParseInt(os.Getenv("BarberID"), 10, 64)
		if err != nil {
			log.Fatal(e.Wrap("can't get barberID from environment variable", err))
		}
		ctx, cancel = context.WithTimeout(context.Background(), config.DbQueryTimoutWrite)
		err = rep.CreateBarber(ctx, barberID)
		cancel()
		if err != nil {
			log.Fatal(err)
		}
		barberIDs = append(barberIDs, barberID)
	}
	return barberIDs
}

func makeBarbersSchedules(rep storage.Storage, barberIDs []int64) {
	err := scheduler.MakeSchedules(rep, barberIDs, config.ScheduledWeeks)
	if err != nil {
		log.Fatal(err)
	}
}
