package initialization

import (
	cp "barbershop-bot/contextprovider"
	ent "barbershop-bot/entities"
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
		testInit()
	})
}

func testInit() {
	barberID := cfg.Barbers.IDs()[0]
	services, err := cp.RepoWithContext.GetServicesByBarberID(barberID)
	if err != nil {
		log.Fatal(err)
	}
	if len(services) == 0 {
		err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: barberID, Name: "Евгений", Phone: "89187591136"})
		if err != nil {
			log.Fatal(err)
		}
		err = cp.RepoWithContext.CreateService(ent.Service{
			BarberID:   barberID,
			Name:       "Бокс и полубокс",
			Desciption: "Бокс, это ультракороткая макушка и выбритые виски с затылком. За названиями причёсок бокс и полубокс стоит одноименный вид спорта. Полубокс, это более длинная версия стрижки Бокс, для офиса она подойдёт больше. Все варианты боксов смотрятся очень стильно и подходят для более спортивных мужчин",
			Price:      500,
			Duration:   30,
		})
		if err != nil {
			log.Fatal(err)
		}
		err = cp.RepoWithContext.CreateService(ent.Service{
			BarberID:   barberID,
			Name:       "Теннис",
			Desciption: "Стрижка хорошо смотрится на молодых парнях и мужчинах среднего возраста. За счет спортивного стиля она хорошо молодит. Для этой модельной стрижки характерны короткие затылок и виски с постепенным увеличением длины на макушке",
			Price:      1000,
			Duration:   60,
		})
		if err != nil {
			log.Fatal(err)
		}
		err = cp.RepoWithContext.CreateService(ent.Service{
			BarberID:   barberID,
			Name:       "Модельная стрижка-Классика",
			Desciption: "Модельная мужская стрижка в классическом варианте с небольшой долей креатива в оформлении. Данная прическа легко подчеркивает индивидуальные особенности внешности и формирует стиль, это что-то среднее между чем-то эпатажным и новым, и спокойным консервативным. Данная стрижка моделируется и подстраивается под конкретного клиента и его внешние особенности.",
			Price:      1500,
			Duration:   90,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
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

func getBarberIDsFromRepo() (barberIDs []int64) {
	barbers, err := cp.RepoWithContext.GetAllBarbers()
	if err != nil {
		log.Fatal(e.Wrap("can't get barberIDs from storage", err))
	}
	for _, barber := range barbers {
		barberIDs = append(barberIDs, barber.ID)
	}
	return
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
