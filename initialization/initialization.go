package initialization

import (
	cp "barbershop-bot/contextprovider"
	cfg "barbershop-bot/lib/config"
	"barbershop-bot/lib/e"
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

func InitGlobals() {
	once.Do(func() {
		cp.InitRepoWithContext(initStorage(createSQLite("data/sqlite/storage.db")))
		actualizeBarberIDs()
		cfg.InitBarberIDs(getBarberIDsFromRepo()...)
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
	storageWithContext := cp.NewStorageContextProvider(storage)
	if err := storageWithContext.Init(); err != nil {
		log.Fatal(e.Wrap("can't initialize storage", err))
	}
	return storage
}

func actualizeBarberIDs() {
	barberIDs := getBarberIDsFromRepo()
	if len(barberIDs) == 0 {
		createBarber(getBarberIDFromEnv())
	}
}

func getBarberIDsFromRepo() []int64 {
	barberIDs, err := cp.RepoWithContext.GetAllBarberIDs()
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
	if err := cp.RepoWithContext.CreateBarber(barberID); err != nil {
		log.Fatal(e.Wrap("can't create barber", err))
	}
}

func initBarbersSchedules() {
	if err := sched.MakeSchedules(); err != nil {
		log.Fatal(e.Wrap("can't make schedules", err))
	}
}
