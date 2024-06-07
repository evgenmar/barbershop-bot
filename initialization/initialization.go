package initialization

import (
	cfg "barbershop-bot/lib/config"
	"barbershop-bot/lib/e"
	rep "barbershop-bot/repository"
	st "barbershop-bot/repository/storage"
	"barbershop-bot/repository/storage/sqlite"
	sched "barbershop-bot/scheduler"
	tg "barbershop-bot/telegram"
	"context"
	"log"
	"os"
	"strconv"
	"sync"
)

var once sync.Once

func Globals() {
	once.Do(func() {
		storage := initStorage(createSQLite("data/sqlite/storage.db"))
		initBarbers(storage)
		initRepository(storage)
		initBarbersSchedules()
		tg.InitBot()
		sched.InitCron()
	})
}

func createSQLite(path string) *sqlite.Storage {
	db, err := sqlite.New(path)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func initStorage(storage st.Storage) st.Storage {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.TimoutWrite)
	err := storage.Init(ctx)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	return storage
}

func initBarbers(storage st.Storage) {
	cfg.InitBarberIDs(actualizeBarberIDs(storage)...)
}

func actualizeBarberIDs(storage st.Storage) []int64 {
	barberIDs := getBarberIDsFromStorage(storage)
	if len(barberIDs) == 0 {
		barberID := getBarberIDFromEnv()
		createBarber(storage, barberID)
		barberIDs = append(barberIDs, barberID)
	}
	return barberIDs
}

func getBarberIDsFromStorage(storage st.Storage) []int64 {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.TimoutRead)
	barberIDs, err := storage.FindAllBarberIDs(ctx)
	cancel()
	if err != nil {
		log.Fatal(err)
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

func createBarber(storage st.Storage, barberID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.TimoutWrite)
	err := storage.CreateBarber(ctx, barberID)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
}

func initRepository(storage st.Storage) {
	rep.InitRepository(storage)
}

func initBarbersSchedules() {
	if err := sched.MakeSchedules(); err != nil {
		log.Fatal(err)
	}
}
