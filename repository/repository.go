package repository

import (
	ent "barbershop-bot/entities"
	"barbershop-bot/lib/e"
	st "barbershop-bot/repository/storage"
	"context"
	"errors"
	"time"
)

type Repository struct {
	storage st.Storage
}

var (
	ErrNoSavedBarber = st.ErrNoSavedBarber
	ErrNonUniqueData = st.ErrNonUniqueData
)

var Rep *Repository

func InitRepository(storage st.Storage) {
	Rep = New(storage)
}

func New(storage st.Storage) *Repository {
	return &Repository{
		storage: storage,
	}
}

func (r *Repository) CreateBarber(ctx context.Context, barberID int64) error {
	return r.storage.CreateBarber(ctx, barberID)
}

func (r *Repository) CreateWorkdays(ctx context.Context, wds ...ent.Workday) error {
	var workdays []st.Workday
	for _, workday := range wds {
		workdays = append(workdays, mapWorkdayToStorage(&workday))
	}
	return r.storage.CreateWorkdays(ctx, workdays...)
}

func (r *Repository) FindAllBarberIDs(ctx context.Context) ([]int64, error) {
	return r.storage.FindAllBarberIDs(ctx)
}

func (r *Repository) GetBarberByID(ctx context.Context, barberID int64) (barber ent.Barber, err error) {
	defer func() {
		if errors.Is(err, st.ErrNoSavedBarber) {
			err = ErrNoSavedBarber
		}
	}()
	br, err := r.storage.GetBarberByID(ctx, barberID)
	if err != nil {
		return ent.Barber{}, err
	}
	return mapBarberToEntity(&br)
}

func (r *Repository) GetLatestWorkDate(ctx context.Context, barberID int64) (date time.Time, err error) {
	defer func() { err = e.WrapIfErr("can't get latest work date", err) }()
	latestWD, err := r.storage.GetLatestWorkDate(ctx, barberID)
	if err != nil && !errors.Is(err, st.ErrNoSavedWorkdates) {
		return time.Time{}, err
	}
	return mapDateToEntity(latestWD)
}

func (r *Repository) UpdateBarberName(ctx context.Context, name string, barberID int64) (err error) {
	defer func() {
		if errors.Is(err, st.ErrNonUniqueData) {
			err = ErrNonUniqueData
		}
	}()
	return r.storage.UpdateBarberName(ctx, name, barberID)
}

func (r *Repository) UpdateBarberPhone(ctx context.Context, phone string, barberID int64) (err error) {
	defer func() {
		if errors.Is(err, st.ErrNonUniqueData) {
			err = ErrNonUniqueData
		}
	}()
	return r.storage.UpdateBarberPhone(ctx, phone, barberID)
}

func (r *Repository) UpdateBarberStatus(ctx context.Context, status ent.Status, barberID int64) error {
	return r.storage.UpdateBarberStatus(ctx, mapStatusToStorage(&status), barberID)
}
