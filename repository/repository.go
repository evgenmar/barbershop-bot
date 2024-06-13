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
	ErrNoSavedBarber      = st.ErrNoSavedBarber
	ErrNonUniqueData      = st.ErrNonUniqueData
	ErrAlreadyExists      = st.ErrAlreadyExists
	ErrAppointmentsExists = st.ErrAppointmentsExists
	ErrInvalidID          = errors.New("invalid ID")
	ErrInvalidWorkday     = errors.New("invalid workday")
	ErrInvalidName        = errors.New("invalid name")
	ErrInvalidPhone       = errors.New("invalid phone")
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

func (r *Repository) CreateBarber(ctx context.Context, barberID int64) (err error) {
	defer func() {
		if errors.Is(err, st.ErrAlreadyExists) {
			err = ErrAlreadyExists
		}
	}()
	return r.storage.CreateBarber(ctx, barberID)
}

func (r *Repository) CreateWorkdays(ctx context.Context, wds ...ent.Workday) (err error) {
	defer func() {
		if errors.Is(err, st.ErrAlreadyExists) {
			err = ErrAlreadyExists
		}
	}()
	var workdays []st.Workday
	for _, workday := range wds {
		wd, err := mapToStorage.workday(&workday)
		if err != nil {
			return err
		}
		workdays = append(workdays, wd)
	}
	return r.storage.CreateWorkdays(ctx, workdays...)
}

func (r *Repository) DeleteWorkdaysByDateRange(ctx context.Context, barberID int64, startDate, endDate time.Time) (err error) {
	defer func() {
		if errors.Is(err, st.ErrAppointmentsExists) {
			err = ErrAppointmentsExists
		}
	}()
	return r.storage.DeleteWorkdaysByDateRange(ctx, barberID, mapToStorage.date(startDate), mapToStorage.date(endDate))
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
	return mapToEntity.barber(&br)
}

func (r *Repository) GetLatestWorkDate(ctx context.Context, barberID int64) (date time.Time, err error) {
	defer func() { err = e.WrapIfErr("can't get latest work date", err) }()
	latestWD, err := r.storage.GetLatestWorkDate(ctx, barberID)
	if err != nil && !errors.Is(err, st.ErrNoSavedWorkdates) {
		return time.Time{}, err
	}
	return mapToEntity.date(latestWD)
}

// UpdateBarber updates only non-empty fields of Barber
func (r *Repository) UpdateBarber(ctx context.Context, barber ent.Barber) (err error) {
	defer func() {
		if errors.Is(err, st.ErrNonUniqueData) {
			err = ErrNonUniqueData
		}
	}()
	br, err := mapToStorage.barber(&barber)
	if err != nil {
		return err
	}
	return r.storage.UpdateBarber(ctx, br)
}