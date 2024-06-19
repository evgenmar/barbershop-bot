package initialization

import (
	cp "barbershop-bot/contextprovider"
	cfg "barbershop-bot/lib/config"
	"barbershop-bot/lib/e"
	rep "barbershop-bot/repository"
	st "barbershop-bot/repository/storage"
	"barbershop-bot/repository/storage/sqlite"
	sched "barbershop-bot/scheduler"
	tg "barbershop-bot/telegram"
	"log"
	"os"
	"strconv"
	"sync"
)

var once sync.Once

var storageContextProvider cp.StorageContextProvider

func InitGlobals() {
	once.Do(func() {
		initStorageContextProvider(createSQLite("data/sqlite/storage.db"))
		initStorage()
		initBarbers()
		initRepository()
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

func initStorageContextProvider(storage st.Storage) {
	storageContextProvider = cp.NewStorageContextProvider(storage)
}

func initStorage() {
	if err := storageContextProvider.Init(); err != nil {
		log.Fatal(e.Wrap("can't initialize storage", err))
	}
}

func initBarbers() {
	cfg.InitBarberIDs(actualizeBarberIDs()...)
}

func actualizeBarberIDs() []int64 {
	barberIDs := getBarberIDsFromStorage()
	if len(barberIDs) == 0 {
		barberID := getBarberIDFromEnv()
		createBarber(barberID)
		barberIDs = append(barberIDs, barberID)
	}
	return barberIDs
}

func getBarberIDsFromStorage() []int64 {
	barberIDs, err := storageContextProvider.FindAllBarberIDs()
	if err != nil {
		log.Fatal(e.Wrap("can't get barberIDs from storage", err))
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

func createBarber(barberID int64) {
	if err := storageContextProvider.CreateBarber(barberID); err != nil {
		log.Fatal(e.Wrap("can't create barber", err))
	}
}

func initRepository() {
	rep.InitRepository(storageContextProvider.Storage)
}

func initBarbersSchedules() {
	if err := sched.MakeSchedules(); err != nil {
		log.Fatal(e.Wrap("can't make schedules", err))
	}
}
